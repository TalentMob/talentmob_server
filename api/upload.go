package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/rathvong/scheduler/task"
	"github.com/rathvong/talentmob_server/models"
)

// HTTP POST - users are able to create a video row in the database
//
//  Categories string `json:"categories"`
//  Thumbnail  string `json:"thumbnail"`
//  Key        string `json:"key"`
//  Title      string `json:"title"`
//
func (s *Server) PostVideo(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	video := models.Video{}

	if err := r.DecodeJsonPayload(&video); err != nil {
		response.SendError(err.Error())
		return
	}

	video.UserID = currentUser.ID
	if err := video.CreateForWeeklyEvents(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	if currentUser.AccountType != models.ACCOUNT_TYPE_TALENT {
		currentUser.AccountType = models.ACCOUNT_TYPE_TALENT
		if err := currentUser.Update(s.Db); err != nil {
			log.Println("PostVideo() Update AccountType ", err)
		}
	}

	response.SendSuccess(video)

	e := models.Event{}

	if err := e.GetAvailableWeeklyEvent(s.Db); err != nil {
		log.Println("weekly event error")
		return
	}

	if s.EventScheduler.Tasks[task.ID(fmt.Sprintf("%d", e.ID))] == nil {
		s.AddEventChannel <- e
	}
}

func (s *Server) PostVideo2(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	video := models.Video{}

	if err := r.DecodeJsonPayload(&video); err != nil {
		response.SendError(err.Error())
		return
	}

	video.UserID = currentUser.ID
	if err := video.Create(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	if currentUser.AccountType != models.ACCOUNT_TYPE_TALENT {
		currentUser.AccountType = models.ACCOUNT_TYPE_TALENT
		if err := currentUser.Update(s.Db); err != nil {
			log.Println("PostVideo() Update AccountType ", err)
		}
	}

	response.SendSuccess(video)

	e := models.Event{}

	if video.EventID == 0 {
		if err := e.GetAvailableWeeklyEvent(s.Db); err != nil {
			log.Println("getWeeklyEvent()", err)
			return
		}

	} else {
		if err := e.GetEventByID(s.Db, video.EventID); err != nil {
			return
		}
	}

	if s.EventScheduler.Tasks[task.ID(fmt.Sprintf("%d", e.ID))] == nil {
		s.AddEventChannel <- e
	}
}

func (s *Server) PostEvent(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	var event models.Event

	if err := r.DecodeJsonPayload(&event); err != nil {
		response.SendError(err.Error())
		return
	}

	env := os.Getenv("env")

	if env == "test" {
		res, err := addEventToProduction(currentUser, event)

		if err != nil {
			log.Println("PostEvent.addEventToProduction Error: ", err)
			response.SendError(err.Error())
			return
		}

		response.SendSuccess(res.Result)
		return
	}

	event.UserID = currentUser.ID

	if event.EventType == models.EventType.LeaderBoard {
		response.SendError("Cannot create leaderboard event here.")
		return
	}

	loc, _ := time.LoadLocation("America/Los_Angeles")

	start := time.Now()
	event.StartDate = start.In(loc)
	event.EndDate = event.StartDate.Add(time.Hour * 168)

	if err := event.Create(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	s.AddEventChannel <- event

	response.SendSuccess(event)

}

func addEventToProduction(user models.User, event models.Event) (*models.BaseResponse, error) {

	url := "https://talentmob.herokuapp.com/api/1/event"

	req, err := http.NewRequest(http.MethodPost, url, NewReader(event))

	req.Header.Add("Authorization", user.Api.Token)

	if err != nil {
		return nil, err
	}

	res, err := Client.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {

		b, err := ioutil.ReadAll(res.Body)

		if err != nil {
			return nil, err
		}

		return nil, errors.New(fmt.Sprintf("request was not successful error: %s statusCode: %d", string(b), res.StatusCode))
	}

	var response models.BaseResponse

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New(response.Info)
	}

	return &response, err

}

func NewReader(data interface{}) io.Reader {
	var buf io.ReadWriter
	buf = new(bytes.Buffer)
	json.NewEncoder(buf).Encode(data)
	return buf
}

var Client = &http.Client{
	Timeout: time.Second * 10,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}
