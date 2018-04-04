package api

import (

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/rathvong/talentmob_server/models"
)

//HTTP GET - retrieve comments for video
// comments will be returned 9 at a time
// params - page, video_id
func (s *Server) GetComments(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	_, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	page := s.GetPageFromParams(r)
	videoID, err := s.GetVideoIDFromParams(r)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	comment := models.Comment{}

	comments, err := comment.GetForVideo(s.Db, videoID, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(comments)

}


func (s *Server) GetUpVotedUsersOnVideo(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	page := s.GetPageFromParams(r)

	videoID, err := s.GetVideoIDFromParams(r)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	video := models.Video{}
	relationship := models.Relationship{}
	users, err := video.UpVotedUsers(s.Db, videoID, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}


	relationships, err  := relationship.PopulateFollowingData(s.Db, currentUser.ID, users)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(relationships)


}


func (s *Server) PostComment(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	comment := models.Comment{}
	r.DecodeJsonPayload(&comment)

	comment.UserID = currentUser.ID

	if err := comment.Create(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	if err := comment.Publisher.GetUser(s.Db, currentUser.ID); err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(comment)
}