package api

import (
	"errors"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/rathvong/talentmob_server/models"
	"github.com/rathvong/talentmob_server/system"
)

var SystemTaskType = SystemTaskTypes{
	addPointsToUsers: "add_points_to_users",
	addEmailSignUp:   "add_email_signup",
}

type SystemTaskTypes struct {
	addPointsToUsers string
	addEmailSignUp   string
}

type SystemTaskParams struct {
	Task     string `json:"task"`
	Extra    string `json:"extra"`
	db       *system.DB
	response *models.BaseResponse
}

func (s *Server) PostPerformSystemTask(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	if !s.AuthenticateHeaderForAdmin(r) {
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
func (tp *SystemTaskParams) Init(response *models.BaseResponse, db *system.DB) {
	tp.response = response
	tp.db = db
}

func (st *SystemTaskParams) validateTasks() (err error) {

	switch st.Task {
	case SystemTaskType.addPointsToUsers:
		st.addPointsToUsers()
	case SystemTaskType.addEmailSignUp:
		st.addEmailSignup()
	default:
		return errors.New(ErrorActionIsNotSupported)

	}

	return
}

func (st *SystemTaskParams) addEmailSignup() {
	address := st.Extra

	if address == "" {
		st.response.SendError("Email address is empty")
		return
	}

	ne := models.NotificationEmail{}

	ne.Address = address

	if err := ne.Create(st.db); err != nil {
		st.response.SendError(err.Error())
		return
	}

	st.response.SendSuccess("Email Saved.")

}

func (st *SystemTaskParams) addPointsToUsers() {
	p := models.Point{}
	if err := p.AddToUsers(st.db); err != nil {
		st.response.SendError(err.Error())
		return
	}

	st.response.SendSuccess("update finished.")
}
