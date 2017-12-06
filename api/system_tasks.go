package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/rathvong/talentmob_server/models"
	"github.com/rathvong/talentmob_server/system"
	"errors"
)


var SystemTaskType = SystemTaskTypes{
	addPointsToUsers:"add_points_to_users",
}

type SystemTaskTypes struct {
	addPointsToUsers string
}

type SystemTaskParams struct {
	Task string `json:"task"`
	db *system.DB
	response *models.BaseResponse
}

func (s *Server) PostPerformSystemTask(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	if !s.AuthenticateHeaderForAdmin(r){
		response.SendError("You do not have access")
		return
	}

	params := SystemTaskParams{}
	r.DecodeJsonPayload(&params)
	params.Init(&response, s.Db)

	if err := params.validateTasks(); err != nil {
		response.SendError(err.Error())
		return
	}

}

// Initialise params with ability to respond to tasks
func (tp *SystemTaskParams) Init(response *models.BaseResponse, db *system.DB){
	tp.response = response
	tp.db = db
}

func (st *SystemTaskParams) validateTasks() (err error){

	switch st.Task {
	case SystemTaskType.addPointsToUsers:
		st.addPointsToUsers()
	default:
		return errors.New(ErrorActionIsNotSupported)

	}

	return
}

func (st *SystemTaskParams) addPointsToUsers(){
	p := models.Point{}
	if err := p.AddToUsers(st.db); err != nil {
		st.response.SendError(err.Error())
		return
	}

	st.response.SendSuccess("update finished.")
}