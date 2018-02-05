package api

import (
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/rathvong/talentmob_server/models"
	"errors"
)




func (s *Server) GetProfile(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	userID, err := s.GetUserIDFromParams(r)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	if userID == 0 {
		userID = currentUser.ID
	}

	user := models.ProfileUser{}

	if err = user.GetUser(s.Db, userID); err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(user)

}

// HTTP GET - retrieve all users import videos
// params - page
func (s *Server) GetImportedVideos(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	page := s.GetPageFromParams(r)
	userID, err := s.GetUserIDFromParams(r)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	if userID == 0 {
		userID = currentUser.ID
	}

	video := models.Video{}
	videos, err := video.GetImportedVideos(s.Db, userID, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(videos)
}

// HTTP GET - retrieve all users favourite videos
// params - page
func (s *Server) GetFavouriteVideos(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	page := s.GetPageFromParams(r)
	userID, err := s.GetUserIDFromParams(r)


	if err != nil {
		response.SendError(err.Error())
		return
	}

	if userID == 0 {
		userID = currentUser.ID
	}

	video := models.Video{}
	videos, err := video.GetFavouriteVideos(s.Db, userID, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(videos)
}

// HTTP POST - update user items

func (s *Server) PostUpdateUser(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	user := models.User{}
	r.DecodeJsonPayload(&user)


	if user.ID != currentUser.ID {
		response.SendError(ErrorUnauthorizedAction)
		return
	}


	if err = user.Update(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(currentUser)
}

/**
	Retrieve users relationships list. If the user is not the user in the profile than
	the server will populate the list with relationship data for the current user.
 */
func (s *Server) GetRelationships(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	page := s.GetPageFromParams(r)


	userID, err := s.GetUserIDFromParams(r)

	if err != nil || userID == 0{

		if (err == nil){
			err = errors.New("missing user_id")
		}
		
		response.SendError(err.Error())

		return
	}

	relationshipName, err := s.GetRelationshipFromParams(r)

	if err != nil {
		response.SendError(err.Error())

		return
	}

	relationship := models.Relationship{}

	var relationships []models.User

	response.SendSuccess(relationships)
	return

	switch relationshipName {
	case "followers":
		relationships, err = relationship.GetFollowers(s.Db, userID, page)


	case "followings":
		relationships, err = relationship.GetFollowing(s.Db, userID, page)

	default:
		response.SendError(err.Error())
		return

	}

	if err != nil {
		response.SendError(err.Error())
		return
	}

	if currentUser.ID != userID {
		relationships, err  = relationship.PopulateFollowingData(s.Db, currentUser.ID, relationships)

	}

	response.SendSuccess(relationships)

}