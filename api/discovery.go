package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/rathvong/talentmob_server/models"
)

// Handles requests for discovery screen
// Will accept params
// page int,
// query string,
// query_type string

func (s *Server) HandleQueries(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	page := s.GetPageFromParams(r)

	qry := models.Query{}
	qry.SetQueryType(s.GetQueryTypeFromParams(r))
	qry.Categories = s.GetCategoryFromParams(r)
	qry.Qry = s.GetQueryFromParams(r)
	qry.UserID = currentUser.ID

	result, err := qry.Find(s.Db, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(result)
}

func (s *Server) HandleQueries2(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	page := s.GetPageFromParams(r)

	qry := models.Query{}
	qry.SetQueryType(s.GetQueryTypeFromParams(r))
	qry.Categories = s.GetCategoryFromParams(r)
	qry.Qry = s.GetQueryFromParams(r)
	qry.UserID = currentUser.ID

	result, err := qry.Find2(s.Db, page)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(result)
}
