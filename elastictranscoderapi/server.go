package elastictranscoderapi

import (
	"net/http"
	"log"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"

	"github.com/rathvong/talentmob_server/system"
	"github.com/rathvong/talentmob_server/models"
	"github.com/ant0ine/go-json-rest/rest"
)

var (
	aws_access_key_id = os.Getenv("AWS_ACCESS_KEY")
	aws_secret_access_key = os.Getenv("AWS_SECRET_KEY")
)

func initTranscoder() *elastictranscoder.ElasticTranscoder {
	creds := credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, "")

	sess, err := session.NewSession(aws.NewConfig().WithCredentials(creds).WithRegion("us-west-2"))

	if err != nil {
		log.Fatal(err)
	}

	return elastictranscoder.New(sess)
}

func mux(db *system.DB) http.Handler{
	service := rest.NewApi()

	api := Service{db: db}

	var DefaultDevStack = []rest.Middleware{
		&rest.AccessLogApacheMiddleware{
			Format: rest.CombinedLogFormat,
		},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.PoweredByMiddleware{},
		&rest.RecoverMiddleware{},
		&rest.GzipMiddleware{},

	}

	service.Use(DefaultDevStack...)
	router, err := rest.MakeRouter(
	//	rest.Get("/transcode/:video_id", api.transcode),
		rest.Get("/transcode/all", api.transcodeAll),
	)


	if err != nil {
		log.Fatal(err)
	}

	service.SetApp(router)


	return service.MakeHandler()
}

type Service struct {
	db *system.DB
	transcodingAllRunning bool
}



func (s *Service) Serve(port int, db *system.DB){

	if db == nil {
		panic("database is nil")
	}

	if aws_access_key_id == "" {
		panic("missing AWS_ACCESS_KEY")
	}

	if aws_secret_access_key == "" {
		panic("missing AWS_SECRET_KEY")
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux(db)))
}

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		response := models.BaseResponse{}
		response.InitV2(w)
		authorization := r.Header.Get("Authorization")

		adminToken := os.Getenv("ADMIN_TOKEN")

		if authorization != adminToken {
			response.SendError(models.ErrorUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

//func (s *Service) transcode(w rest.ResponseWriter, r rest.Request) {
//
//	response := models.BaseResponse{}
//	response.InitV2(w)
//
//	idString := bone.GetValue(r, "video_id")
//
//	_, err := strconv.Atoi(idString)
//
//	if err != nil {
//		response.SendError("bad request: " + err.Error())
//		return
//	}
//
//}

func (s *Service) transcodeAll(w rest.ResponseWriter, r *rest.Request) {
	log.Println("Transcode all: start")
	var t models.Transcoded
	var response models.BaseResponse


	response.Init(w)

	if s.transcodingAllRunning {
		response.SendError("job is already in progress")
		return
	}

	videos, err := t.GetNeedsTranscodedWatermarkVideos(s.db)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	log.Printf("Number of videos to transcode: %d", len(videos))

	s.transcodingAllRunning = true

	et := initTranscoder()

		for _, video := range videos {

			log.Printf("transcoding video: %+v\n ", video)


			outputKey := video.Key + ".mp4"
			thumbnailPattern := "thumb_" + video.Key + "-{count}"

			waterMarkInputKey := "large_watermark.png"
			waterMarkPresetId := "BottomRight"

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
					PresetId:         aws.String("1529065895427-3219z0"), // Generic 1080p H.264
					Rotate:           aws.String("auto"),
					ThumbnailPattern: aws.String(thumbnailPattern),
					Watermarks: []*elastictranscoder.JobWatermark{waterMark},
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
				VideoID: video.ID,
				TranscodedCompleted: false,
				TranscodedWatermarkKey: outputKey,
				TranscodedThumbnailKey: thumbnailPattern,
				TranscodedKey: outputKey,
				WatermarkCompleted:true,
			}

			if exists, err := trancoded.Exists(s.db, video.ID); err != nil || exists {

				if err != nil {
					log.Print(err)
					continue
				}

				if err := trancoded.GetByVideoID(s.db, video.ID); err != nil {
					log.Print(err)
					continue
				}

				trancoded.WatermarkCompleted = true

				if err := trancoded.Update(s.db); err != nil {
					log.Println(err)
					
				}

				continue
			}

			if err := trancoded.Create(s.db); err != nil {
				log.Printf("Transcode All: video_id: %v Error %v", video.ID, err)
			}

		}
		s.transcodingAllRunning = false


	response.SendSuccess("Transcoding Job has started")
}
