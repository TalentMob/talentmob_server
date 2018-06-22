package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/rathvong/talentmob_server/models"
	"github.com/rathvong/talentmob_server/system"
)

var (
	aws_access_key_id     = os.Getenv("AWS_ACCESS_KEY")
	aws_secret_access_key = os.Getenv("AWS_SECRET_KEY")
)

var (
	presetID          = "1529585755448-pzxaue"
	transcodedPreset  = "1529672652778-7mh96k"
	waterMarkInputKey = "large_watermark.png"
	waterMarkPresetId = "BottomRight"
)

var SystemTaskType = SystemTaskTypes{
	AddPointsToUsers:                "add_points_to_users",
	AddEmailSignUp:                  "add_email_signup",
	TranscodeWithWatermarkAllVideos: "transcode_with_watermark_all_videos",
	TranscodeWithWatermarkVideo:     "transcode_with_watermark_video",
	TranscodeAllVideos:              "transcode_all_videos",
	TranscodeVideo:                  "transcode_video",
}

var transcodingAllWithWatermarkRunning bool

var transcodingAllRunning bool

func initTranscoder() *elastictranscoder.ElasticTranscoder {
	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, "")

	sess, err := session.NewSession(aws.NewConfig().WithCredentials(creds).WithRegion("us-west-2"))

	if err != nil {
		log.Fatal(err)
	}

	return elastictranscoder.New(sess)
}

type SystemTaskTypes struct {
	AddPointsToUsers                string
	AddEmailSignUp                  string
	TranscodeWithWatermarkAllVideos string
	TranscodeWithWatermarkVideo     string
	TranscodeVideo                  string
	TranscodeAllVideos              string
}

type SystemTaskParams struct {
	Task     string `json:"task"`
	Extra    string `json:"extra"`
	db       *system.DB
	response *models.BaseResponse
}

func (s *Server) PostPerformSystemTask(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	if !s.AuthenticateHeaderForAdmin(r) {
		response.SendError("You do not have access")
		return
	}

	params := SystemTaskParams{}
	r.DecodeJsonPayload(&params)
	params.Init(&response, s.Db)

	if err := params.validateTasks(); err != nil {
		response.SendError(err.Error())
		return
	}

}

// Initialise params with ability to respond to tasks
func (tp *SystemTaskParams) Init(response *models.BaseResponse, db *system.DB) {
	tp.response = response
	tp.db = db
}

func (st *SystemTaskParams) validateTasks() (err error) {

	switch st.Task {
	case SystemTaskType.AddPointsToUsers:
		st.addPointsToUsers()
	case SystemTaskType.AddEmailSignUp:
		st.addEmailSignup()
	case SystemTaskType.TranscodeWithWatermarkAllVideos:
		st.transcodeWithWatermarkAllVideos()
	case SystemTaskType.TranscodeWithWatermarkVideo:
		st.transcodeWithWatermarkVideo()
	case SystemTaskType.TranscodeVideo:
		st.transcodeVideo()
	case SystemTaskType.TranscodeAllVideos:
		st.transcodeAllVideos()
	default:
		return errors.New(ErrorActionIsNotSupported + fmt.Sprintf(" Task Available: %+v", SystemTaskType))
	}

	return
}

func (st *SystemTaskParams) addEmailSignup() {
	address := st.Extra

	if address == "" {
		st.response.SendError("Email address is empty")
		return
	}

	ne := models.NotificationEmail{}

	ne.Address = address

	if err := ne.Create(st.db); err != nil {
		st.response.SendError(err.Error())
		return
	}

	st.response.SendSuccess("Email Saved.")

}

func (st *SystemTaskParams) addPointsToUsers() {
	p := models.Point{}
	if err := p.AddToUsers(st.db); err != nil {
		st.response.SendError(err.Error())
		return
	}

	st.response.SendSuccess("update finished.")
}

func (st *SystemTaskParams) transcodeWithWatermarkVideo() {

	if st.Extra == "" {
		st.response.SendError("missing extra={video_id}")
		return
	}

	videoID, err := strconv.Atoi(st.Extra)
	if err != nil {
		st.response.SendError(err.Error())
		return
	}

	var video models.Video

	if err := video.GetVideoByID(st.db, uint64(videoID)); err != nil {
		st.response.SendError(err.Error())
		return
	}

	log.Printf("transcoding video: %+v\n ", video)
	et := initTranscoder()

	outputKey := video.Key + ".mp4"
	thumbnailPattern := video.Key + "-{count}"

	waterMark := &elastictranscoder.JobWatermark{InputKey: &waterMarkInputKey, PresetWatermarkId: &waterMarkPresetId}

	params := &elastictranscoder.CreateJobInput{
		Input: &elastictranscoder.JobInput{
			AspectRatio: aws.String("auto"),
			Container:   aws.String("auto"),
			FrameRate:   aws.String("auto"),
			Interlaced:  aws.String("auto"),
			Key:         aws.String(video.Key), // the "filename" in S3
			Resolution:  aws.String("auto"),
		},
		PipelineId: aws.String("1528550420987-fmmf1s"), // Pipeline can be created via console
		Output: &elastictranscoder.CreateJobOutput{
			Key:              aws.String(outputKey),
			PresetId:         aws.String(presetID), // Generic 1080p H.264
			Rotate:           aws.String("auto"),
			ThumbnailPattern: aws.String(thumbnailPattern),
			Watermarks:       []*elastictranscoder.JobWatermark{waterMark},
		},
	}

	if err := params.Validate(); err != nil {
		st.response.SendError(err.Error())
		return
	}

	res, err := et.CreateJob(params)

	if err != nil {
		st.response.SendError(err.Error())
		return
	}

	log.Printf("Job Response: %v\n", res.Job)

	var trancoded = models.Transcoded{
		VideoID:                video.ID,
		TranscodedWatermarkKey: outputKey,
		TranscodedThumbnailKey: thumbnailPattern,
		TranscodedKey:          outputKey,
	}

	if exists, err := trancoded.Exists(st.db, video.ID); err != nil || exists {

		if err != nil {
			st.response.SendError(err.Error())
			return
		}

		st.response.SendSuccess("Transcoding Job has started")
		return
	}

	if err := trancoded.Create(st.db); err != nil {
		log.Printf("Transcode All: video_id: %v Error %v", video.ID, err)
		st.response.SendError(err.Error())
	}

	st.response.SendSuccess("Transcoding Job has started")
}

func (st *SystemTaskParams) transcodeWithWatermarkAllVideos() {
	log.Println("Transcode all: start")
	var t models.Transcoded

	if transcodingAllWithWatermarkRunning {
		st.response.SendError("job is already in progress")
		return
	}

	videos, err := t.GetAllVideos(st.db)

	if err != nil {
		st.response.SendError(err.Error())
		return
	}

	log.Printf("Number of videos to transcode: %d", len(videos))

	transcodingAllWithWatermarkRunning = true

	et := initTranscoder()

	go func() {

		for _, video := range videos {

			log.Printf("transcoding video: %+v\n ", video)
			outputKey := video.Key + ".mp4"
			thumbnailPattern := video.Key + "-{count}"

			waterMark := &elastictranscoder.JobWatermark{InputKey: &waterMarkInputKey, PresetWatermarkId: &waterMarkPresetId}

			params := &elastictranscoder.CreateJobInput{
				Input: &elastictranscoder.JobInput{
					AspectRatio: aws.String("auto"),
					Container:   aws.String("auto"),
					FrameRate:   aws.String("auto"),
					Interlaced:  aws.String("auto"),
					Key:         aws.String(video.Key), // the "filename" in S3
					Resolution:  aws.String("auto"),
				},
				PipelineId: aws.String("1528550420987-fmmf1s"), // Pipeline can be created via console
				Output: &elastictranscoder.CreateJobOutput{
					Key:              aws.String(outputKey),
					PresetId:         aws.String(presetID), // Generic 1080p H.264
					Rotate:           aws.String("auto"),
					ThumbnailPattern: aws.String(thumbnailPattern),
					Watermarks:       []*elastictranscoder.JobWatermark{waterMark},
				},
			}

			if err := params.Validate(); err != nil {
				continue
			}

			res, err := et.CreateJob(params)

			if err != nil {
				log.Println("Failed to create job: ", err)
				continue
			}

			log.Printf("Job Response: %v\n", res.Job)

			var trancoded = models.Transcoded{
				VideoID:                video.ID,
				TranscodedWatermarkKey: outputKey,
				TranscodedThumbnailKey: thumbnailPattern,
				TranscodedKey:          outputKey,
			}

			if exists, err := trancoded.Exists(st.db, video.ID); err != nil || exists {

				if err != nil {
					log.Print(err)
					continue
				}

				continue
			}

			if err := trancoded.Create(st.db); err != nil {
				log.Printf("Transcode All: video_id: %v Error %v", video.ID, err)
			}

			st.response.SendSuccess("Transcoding Job has started")
		}

		transcodingAllWithWatermarkRunning = false
	}()

	st.response.SendSuccess("Transcoding Job has started")
}

func (st *SystemTaskParams) transcodeVideo() {

	if st.Extra == "" {
		st.response.SendError("missing extra={video_id}")
		return
	}

	videoID, err := strconv.Atoi(st.Extra)
	if err != nil {
		st.response.SendError(err.Error())
		return
	}

	var video models.Video

	if err := video.GetVideoByID(st.db, uint64(videoID)); err != nil {
		st.response.SendError(err.Error())
		return
	}

	log.Printf("transcoding video: %+v\n ", video)
	et := initTranscoder()

	outputKey := video.Key + ".mp4"
	thumbnailPattern := video.Key + "-{count}"

	params := &elastictranscoder.CreateJobInput{
		Input: &elastictranscoder.JobInput{
			AspectRatio: aws.String("auto"),
			Container:   aws.String("auto"),
			FrameRate:   aws.String("auto"),
			Interlaced:  aws.String("auto"),
			Key:         aws.String(video.Key), // the "filename" in S3
			Resolution:  aws.String("auto"),
		},
		PipelineId: aws.String("1529303979535-ru9lk4"), // Pipeline can be created via console
		Output: &elastictranscoder.CreateJobOutput{
			Key:              aws.String(outputKey),
			PresetId:         aws.String(transcodedPreset), // Generic 1080p H.264
			Rotate:           aws.String("auto"),
			ThumbnailPattern: aws.String(thumbnailPattern),
		},
	}

	if err := params.Validate(); err != nil {
		st.response.SendError(err.Error())
		return
	}

	res, err := et.CreateJob(params)

	if err != nil {
		st.response.SendError(err.Error())
		return
	}

	log.Printf("Job Response: %v\n", res.Job)

	var trancoded = models.Transcoded{
		VideoID:                video.ID,
		TranscodedWatermarkKey: outputKey,
		TranscodedThumbnailKey: thumbnailPattern,
		TranscodedKey:          outputKey,
	}

	if exists, err := trancoded.Exists(st.db, video.ID); err != nil || exists {

		if err != nil {
			st.response.SendError(err.Error())
			return
		}

		st.response.SendSuccess("Transcoding Job has started")
		return
	}

	if err := trancoded.Create(st.db); err != nil {
		log.Printf("Transcode All: video_id: %v Error %v", video.ID, err)
		st.response.SendError(err.Error())
		return
	}

	st.response.SendSuccess("Transcoding Job has started")
}

func (st *SystemTaskParams) transcodeAllVideos() {
	log.Println("Transcode all: start")
	var t models.Transcoded

	if transcodingAllRunning {
		st.response.SendError("job is already in progress")
		return
	}

	videos, err := t.GetAllVideos(st.db)

	if err != nil {
		st.response.SendError(err.Error())
		return
	}

	log.Printf("Number of videos to transcode: %d", len(videos))

	transcodingAllRunning = true

	et := initTranscoder()

	go func() {

		for _, video := range videos {

			log.Printf("transcoding video: %+v\n ", video)

			outputKey := video.Key + ".mp4"
			thumbnailPattern := video.Key + "-{count}"

			params := &elastictranscoder.CreateJobInput{
				Input: &elastictranscoder.JobInput{
					AspectRatio: aws.String("auto"),
					Container:   aws.String("auto"),
					FrameRate:   aws.String("auto"),
					Interlaced:  aws.String("auto"),
					Key:         aws.String(video.Key), // the "filename" in S3
					Resolution:  aws.String("auto"),
				},
				PipelineId: aws.String("1529303979535-ru9lk4"), // Pipeline can be created via console
				Output: &elastictranscoder.CreateJobOutput{
					Key:              aws.String(outputKey),
					PresetId:         aws.String(transcodedPreset), // Generic 1080p H.264
					Rotate:           aws.String("auto"),
					ThumbnailPattern: aws.String(thumbnailPattern),
				},
			}

			if err := params.Validate(); err != nil {
				continue
			}

			res, err := et.CreateJob(params)

			if err != nil {
				log.Println("Failed to create job: ", err)
				continue
			}

			log.Printf("Job Response: %v\n", res.Job)

			var trancoded = models.Transcoded{
				VideoID:                video.ID,
				TranscodedWatermarkKey: outputKey,
				TranscodedThumbnailKey: thumbnailPattern,
				TranscodedKey:          outputKey,
			}

			if exists, err := trancoded.Exists(st.db, video.ID); err != nil || exists {

				if err != nil {
					log.Print(err)
					continue
				}

				continue
			}

			if err := trancoded.Create(st.db); err != nil {
				log.Printf("Transcode All: video_id: %v Error %v", video.ID, err)
			}

		}

		transcodingAllRunning = false
	}()

	st.response.SendSuccess("Transcoding Job has started")
}

type ElasticTranscoderResponse struct {
	State      string `json:"state"`
	Version    string `json:"version"`
	JobID      string `json:"jobId"`
	PipelineID string `json:"pipelineId"`
	Input      Input  `json:"input"`
	Outputs    string `json:"output"`
}

type Input struct {
	Key string `json:"key"`
}

type Output struct {
	ID               string `json:"id"`
	PresetID         string `json:"presedId"`
	Key              string `json:"key"`
	ThumbnailPattern string `json:"thumbnailPattern"`
	Rotate           string `json:"rotate"`
	Status           string `json:"status"`
	Duration         int    `json:"duration"`
}

func (s *Server) PostElasticTranscoding(w rest.ResponseWriter, r *rest.Request) {

	var er ElasticTranscoderResponse
	var en ElasticTranscoderNotification
	response := models.BaseResponse{}

	response.Init(w)

	if err := r.DecodeJsonPayload(&er); err != nil {
		log.Println("PostElasticTransoding() Error: ", err)
		response.SendError(err.Error())
		return
	}

	if err := en.SetResponse(er); err != nil {
		log.Println("PostElasticTransoding() Error: ", err)

		response.SendError(err.Error())
		return
	}

	var transcoded models.Transcoded

	if err := transcoded.GetByTranscodedKey(s.Db, en.Key); err != nil {
		log.Println("PostElasticTransoding() Error: ", err)

		response.SendError(err.Error())
		return
	}

	en.TranscodedID = transcoded.ID

	if err := en.Create(s.Db); err != nil {
		log.Println("PostElasticTransoding() Error: ", err)

		response.SendError(err.Error())
		return
	}

	response.SendSuccess(en)

}

type ElasticTranscoderNotification struct {
	models.BaseModel
	JobID        string `json:"job_id"`
	TranscodedID uint64 `json:"transcoded_id"`
	PipelineID   string `json:"pipeline_id"`
	Key          string `json:"key"`
	State        string `json:"state"`
	Status       string `json:"status"`
	IsActive     bool   `json:"is_active"`
}

func (e *ElasticTranscoderNotification) SetResponse(r ElasticTranscoderResponse) error {
	e.JobID = r.JobID
	e.IsActive = true
	e.PipelineID = r.PipelineID

	var outputs []Output

	outputs = make([]Output, 0)

	if err := json.Unmarshal([]byte(r.Outputs), &outputs); err != nil {
		return err
	}

	e.Key = outputs[0].Key
	e.State = r.State
	e.Status = outputs[0].Status

	return nil
}

func (e *ElasticTranscoderNotification) queryCreate() string {
	return `INSERT INTO elastic_transcoder_notifications 
						(job_id, transcoded_id, pipeline_id, key, state, status, is_active, created_at, updated_at)
						VALUES
						($1, $2, $3, $4, $5, $6, $7, $8, $9)
						RETURNING id`
}

func (e *ElasticTranscoderNotification) queryUpdate() string {
	return `UPDATE elastic_transcoder_notifications SET 
			transcoded_id = $2,
			pipeline_id = $3,
			key = $4,
			state = $5,
			status = $6,
			is_active = $7,
			updated_at = $8
			WHERE job_id = $1`
}

func (e *ElasticTranscoderNotification) Create(db *system.DB) error {

	if e.JobID == "" {
		return e.Errors(models.ErrorMissingValue, "job_id")
	}

	if e.Key == "" {
		return e.Errors(models.ErrorMissingID, "key")
	}

	if e.PipelineID == "" {
		return e.Errors(models.ErrorMissingValue, "pipeline_id")
	}

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}()

	e.IsActive = true
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()

	err = tx.QueryRow(
		e.queryCreate(),
		e.JobID,
		e.TranscodedID,
		e.PipelineID,
		e.Key,
		e.State,
		e.Status,
		e.IsActive,
		e.CreatedAt,
		e.UpdatedAt,
	).Scan(&e.ID)

	if err != nil {
		log.Printf("ElasticTranscoderNotification.Create() Query: %v Error: %v", e.queryCreate(), err)
	}

	return nil
}

func (e *ElasticTranscoderNotification) Update(db *system.DB) error {

	if e.ID == 0 {
		return e.Errors(models.ErrorMissingID, "id")
	}

	if e.JobID == "" {
		return e.Errors(models.ErrorMissingValue, "job_id")
	}

	if e.Key == "" {
		return e.Errors(models.ErrorMissingID, "key")
	}

	if e.PipelineID == "" {
		return e.Errors(models.ErrorMissingValue, "pipeline_id")
	}

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}()

	e.UpdatedAt = time.Now()

	_, err = tx.Exec(
		e.queryUpdate(),
		e.JobID,
		e.TranscodedID,
		e.PipelineID,
		e.Key,
		e.State,
		e.Status,
		e.IsActive,
		e.UpdatedAt,
	)

	if err != nil {
		log.Printf("ElasticTranscoderNotification.Update() Query: %v Error: %v", e.queryUpdate(), err)
	}

	return nil
}
