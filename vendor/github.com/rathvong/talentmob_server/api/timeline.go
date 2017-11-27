package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/rathvong/talentmob_server/models"
)

//HTTP GET - retrieve users time-line for videos to vote on
// videos will be returned 9 at a time
func (s *Server) GetTimeLine(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	video := models.Video{}
	videos, err := video.GetTimeLine(s.Db, currentUser.ID, 1)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(videos)

}

//HTTP GET - retrieve leader board list
// videos will be returned 9 at a time
// params - page
func (s *Server) GetLeaderBoard(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	page := s.GetPageFromParams(r)


	video := models.Video{}
	videos, err := video.GetLeaderBoard(s.Db, page, currentUser.ID)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(videos)

}

//HTTP GET - retrieve users voting history
// videos will be returned 9 at a time
// params - page
func (s *Server) GetHistory(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	page := s.GetPageFromParams(r)
	video := models.Video{}
	videos, err := video.GetHistory(s.Db, currentUser.ID, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(videos)

}

//HTTP GET - retrieve users voting history
// videos will be returned 9 at a time
// params - page
func (s *Server) GetLeaderBoardHistory(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	page := s.GetPageFromParams(r)
	eventID, err := s.GetEventIDFromParams(r)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	compete := models.Competitor{}
	videos, err := compete.GetHistory(s.Db, eventID, currentUser.ID, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(videos)

}


// Return all events
func (s *Server) GetEvents(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	_, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}



	if err != nil {
		response.SendError(err.Error())
		return
	}

	page := s.GetPageFromParams(r)

	event := models.Event{}

	events, err := event.GetAllEvents(s.Db, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(events)

}