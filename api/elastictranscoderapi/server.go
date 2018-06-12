package elastictranscoderapi

import (
	"net/http"

	"github.com/go-zoo/bone"
	"github.com/go-zoo/claw"


	"github.com/rathvong/talentmob_server/system"
	"github.com/go-zoo/claw/middleware"
	"os"
	"github.com/rathvong/talentmob_server/models"
	"strconv"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"log"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
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

func mux(db *system.DB) *bone.Mux{
	mux := bone.New()

	c := claw.New(middleware.Recovery)

	service := Service{db: db, et: initTranscoder()}

	mux.Get("/transcode/:video_id", c.Use(service.transcode).Add(Authenticate))
	mux.Get("/transcode/all", c.Use(service.transcodeAll).Add(Authenticate))

	return mux
}

type Service struct {
	db *system.DB
	et *elastictranscoder.ElasticTranscoder
	transcodingAllRunning bool
}

func (s *Service) Serve(port int, db *system.DB){
	http.ListenAndServe(fmt.Sprintf(":%d", port), mux(db))
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

func (s *Service) transcode(w http.ResponseWriter, r *http.Request) {



	response := models.BaseResponse{}
	response.InitV2(w)



	idString := bone.GetValue(r, "video_id")

	_, err := strconv.Atoi(idString)

	if err != nil {
		response.SendError("bad request: " + err.Error())
		return
	}

}

func (s *Service) transcodeAll(w http.ResponseWriter, r *http.Request) {
	var t models.Transcoded
	var response models.BaseResponse


	response.InitV2(w)

	if s.transcodingAllRunning {
		response.SendError("job is already in progress")
		return
	}

	videos, err := t.GetNeedsTranscodedWatermarkVideos(s.db)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	s.transcodingAllRunning = true

	for _, video := range *videos {

		outputKey := video.Key + ".mp4"
		thumbnailPattern := "thumb_" + video.Key

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
				PresetId:         aws.String("1528607447282-z2dgbc"), // Generic 1080p H.264
				Rotate:           aws.String("auto"),
				ThumbnailPattern: aws.String(thumbnailPattern),
			},
		}

		if err := params.Validate(); err != nil {
			continue
		}

		res, err := s.et.CreateJob(params)

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

		if err := trancoded.Create(s.db); err != nil {
			log.Printf("Transcode All: video_id: %v Error %v", video.ID, err)
		}



	}

	s.transcodingAllRunning = false
}
