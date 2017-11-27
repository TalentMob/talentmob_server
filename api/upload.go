package api

import (
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
func (s *Server) PostVideo(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	video := models.Video{}

	r.DecodeJsonPayload(&video)

	video.UserID = currentUser.ID
	if err := video.Create(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(video)
}