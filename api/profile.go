package api

import (
	"fmt"

	"github.com/ant0ine/go-json-rest/rest"

	"database/sql"
	"errors"

	"github.com/rathvong/talentmob_server/badgecontroller"
	"github.com/rathvong/talentmob_server/models"
)

func (s *Server) GetProfile(w rest.ResponseWriter, r *rest.Request) {
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

	relationship := models.Relationship{}

	user.IsFollowing, err = relationship.IsFollowing(s.Db, user.ID, currentUser.ID)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	_, user.RankMob, err = currentUser.RankAgainstMob(s.Db, user.ID)

	if err != nil && err != sql.ErrNoRows {
		response.SendError(err.Error())
		return
	}

	_, user.RankTalent, err = currentUser.RankAgainstTalent(s.Db, user.ID)

	if err != nil && err != sql.ErrNoRows {
		response.SendError(err.Error())
		return
	}

	qryImport := fmt.Sprintf("SELECT COUNT(*) FROM videos WHERE user_id=%d AND is_active=true", user.ID)

	if err := s.Db.QueryRow(qryImport).Scan(&user.ImportedVideosCount); err != nil {
		response.SendError(err.Error())
		return
	}

	qryFavourite := fmt.Sprintf("SELECT COUNT(*) FROM votes WHERE user_id=%d AND upvote > 0", user.ID)

	if err := s.Db.QueryRow(qryFavourite).Scan(&user.FavouriteVideosCount); err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(user)
}

func (s *Server) GetProfile2(w rest.ResponseWriter, r *rest.Request) {
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

	if err = user.GetUser2(s.Db, userID, currentUser.ID); err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(user)
}

// HTTP GET - retrieve all users import videos
// params - page
func (s *Server) GetImportedVideos(w rest.ResponseWriter, r *rest.Request) {
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
func (s *Server) GetFavouriteVideos(w rest.ResponseWriter, r *rest.Request) {
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

func (s *Server) GetImportedVideos2(w rest.ResponseWriter, r *rest.Request) {
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
	videos, err := video.GetImportedVideos2(s.Db, userID, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(videos)
}

// HTTP GET - retrieve all users favourite videos
// params - page
func (s *Server) GetFavouriteVideos2(w rest.ResponseWriter, r *rest.Request) {
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
	videos, err := video.GetFavouriteVideos2(s.Db, userID, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(videos)
}

// HTTP GET - retrieve all users achievements
// params - page
func (s *Server) GetStats(w rest.ResponseWriter, r *rest.Request) {
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

	b := new(badgecontroller.Badge)

	stats, err := b.List(s.Db, userID)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(stats)
}

// HTTP POST - update user items

func (s *Server) PostUpdateUser(w rest.ResponseWriter, r *rest.Request) {
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

	currentUser.Avatar = user.Avatar
	currentUser.Name = user.Name

	if err = currentUser.Update(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	currentUser.IsReturning = true

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

	if err != nil || userID == 0 {

		if err == nil {
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

	switch relationshipName {
	case "followers":
		relationships, err = relationship.GetFollowers(s.Db, userID, page)

	case "followings":
		relationships, err = relationship.GetFollowing(s.Db, userID, page)

	default:

		err = errors.New("unrecognized relationship")

		response.SendError(err.Error())
		return

	}

	if err != nil {
		response.SendError(err.Error())
		return
	}

	relationships, err = relationship.PopulateFollowingData(s.Db, currentUser.ID, relationships)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(relationships)

}

func (s *Server) GetRelationships2(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	_, err := s.LoginProcess(response, r)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	page := s.GetPageFromParams(r)

	userID, err := s.GetUserIDFromParams(r)

	if err != nil || userID == 0 {

		if err == nil {
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

	switch relationshipName {
	case "followers":
		relationships, err = relationship.GetFollowers2(s.Db, userID, page)

	case "followings":
		relationships, err = relationship.GetFollowing2(s.Db, userID, page)

	default:

		err = errors.New("unrecognized relationship")

		response.SendError(err.Error())
		return

	}

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(relationships)

}
