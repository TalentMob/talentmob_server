package api

import (
	"log"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

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
