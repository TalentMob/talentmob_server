package general

import (
	"github.com/ant0ine/go-json-rest/rest"

	"errors"

	"encoding/json"
	"github.com/rathvong/talentmob_server/models"

	"database/sql"
	"github.com/rathvong/talentmob_server/system"
	"github.com/rathvong/util"

	"log"
)

const (
	ErrorMissingID            = "error missing id"
	ErrorMissingExtra         = "error missing extra"
	ErrorActionIsNotSupported = "action is not supported for this model"
	ErrorUnauthorizedAction   = "action is not authorized"
	ErrorModelIsNotFound      = "model is not found"
)

// Task action will handle what users will be capable of
// performing in the app
type TaskAction struct {
	follow      string
	unfollow    string
	upvote      string
	downvote    string
	like        string
	unlike      string
	create      string
	delete      string
	update      string
	exists      string
	updateFCM   string
	logout      string
	accountType string
	get         string
	top         string
	add         string
}

// register values for each action field
var taskAction = TaskAction{
	upvote:      "upvote",
	downvote:    "downvote",
	follow:      "follow",
	unfollow:    "unfollow",
	like:        "like",
	unlike:      "unlike",
	create:      "create",
	delete:      "delete",
	update:      "update",
	exists:      "exists",
	updateFCM:   "update_fcm",
	logout:      "logout",
	accountType: "account_type",
	get:         "get",
	top:         "top",
	add:         "add"}

// Handle what type of models tasks can be performed on
type TaskModel struct {
	user     string
	vote     string
	video    string
	view     string
	bio      string
	comment  string
	category string
	point    string
	boost    string
	event    string
}

// register values for each model field
var taskModel = TaskModel{
	user:     "user",
	vote:     "vote",
	video:    "video",
	view:     "view",
	bio:      "bio",
	comment:  "comment",
	category: "category",
	point:    "point",
	boost:    "boost",
	event:    "event",
}

// Will handle all requests from user
type TaskParams struct {
	Model       string `json:"model"`
	Action      string `json:"action"`
	ID          uint64 `json:"id"`
	Extra       string `json:"extra"`
	response    *models.BaseResponse
	currentUser *models.User
	db          *system.DB
}

// HTTP POST - Handle all micro services to update simple models
// params - action, model, id, extra
//
//
// Model       string `json:"model"`
// Action      string `json:"action"`
// ID          uint64 `json:"id"`
// Extra       string `json:"extra"`
//
func (s *Server) PostPerformTask(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	params := TaskParams{}
	r.DecodeJsonPayload(&params)

	if err := params.validateTasks(); err != nil {
		response.SendError(err.Error())
		return
	}

	params.Init(&response, &currentUser, s.Db)
	params.HandleTasks()

}

// Initialise params with ability to respond to tasks
func (tp *TaskParams) Init(response *models.BaseResponse, user *models.User, db *system.DB) {
	tp.response = response
	tp.currentUser = user
	tp.db = db
}

// Validate if proper tasks are requested
func (tp *TaskParams) validateTasks() (err error) {
	if tp.Model == "" {
		return errors.New("missing model")
	}

	if tp.Action == "" {
		return errors.New("missing action")
	}

	return
}

// Handle which models to be performed on
func (tp *TaskParams) HandleTasks() {
	switch tp.Model {
	case taskModel.video:
		tp.HandleVideoTasks()
	case taskModel.user:
		tp.HandleUserTasks()
	case taskModel.vote:
		tp.HandleVoteTasks()
	case taskModel.view:
		tp.HandleViewTasks()
	case taskModel.bio:
		tp.HandleBioTasks()
	case taskModel.comment:
		tp.HandleCommentTasks()
	case taskModel.category:
		tp.HandleCategoryTasks()
	case taskModel.boost:
		tp.HandleBoostTasks()
	case taskModel.point:
		tp.HandlePointTasks()
	case taskModel.event:
		tp.HandleEventTasks()
	default:
		tp.response.SendError(ErrorModelIsNotFound)
	}
}

func (tp *TaskParams) HandleEventTasks() {
	switch tp.Action {
	case taskAction.get:
		tp.HandleEventGet()
	default:
		tp.response.SendError(ErrorModelIsNotFound)
	}
}

func (tp *TaskParams) HandleEventGet() {
	switch tp.Extra {
	case "winner_last_closed_event":
		tp.HandleGetWinnerLastClosedEvent()
	default:
		tp.response.SendError(ErrorModelIsNotFound)
	}
}

func (tp *TaskParams) HandleGetWinnerLastClosedEvent() {
	event := models.Event{}

	events, err := event.GetAllEvents(tp.db, 3, 0)

	var competition models.Competitor

	var topVideo models.Video

	if err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if len(events) == 3 {
		videos, err := competition.GetHistory(tp.db, events[2].ID, tp.currentUser.ID, 1, 0)

		if err != nil {
			tp.response.SendError(err.Error())
			return
		}

		topVideo = videos[0]
	}

	tp.response.SendSuccess(topVideo)
}

func (tp *TaskParams) HandlePointTasks() {
	switch tp.Action {
	case taskAction.get:
		tp.HandlePointGet()
	case taskAction.add:
		tp.HandlePointAdd()
	default:
		tp.response.SendError(ErrorUnauthorizedAction)
	}

}

func (tp *TaskParams) HandlePointGet() {
	switch tp.Extra {
	case "ads_watched_today":
		tp.HandleGetAdsWatched()
	default:
		tp.HandleGetPoints()
	}
}

func (tp *TaskParams) HandlePointAdd() {
	switch tp.Extra {
	case models.POINT_ADS:
		if tp.Extra == "" {
			tp.response.SendError(ErrorMissingExtra)
			return
		}
		tp.HandleAdPoints()
	default:
		tp.response.SendError(ErrorActionIsNotSupported)
		return
	}
}

func (tp *TaskParams) HandleGetAdsWatched() {
	log.Println("HandleGetAdsWatched()")
	ap := models.AdPoint{}

	count, err := ap.GetAdsWatched(tp.db, tp.currentUser.ID)

	if err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(count)
}

func (tp *TaskParams) HandleGetPoints() {
	p := models.Point{}

	if err := p.GetByUserID(tp.db, tp.currentUser.ID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(p)
}

func (tp *TaskParams) HandleAdPoints() {

	ap := models.AdPoint{}

	ap.UserID = tp.currentUser.ID

	if err := ap.Create(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(ap)
}

func (tp *TaskParams) HandleBoostTasks() {
	if tp.ID == 0 {
		tp.response.SendError(ErrorMissingID)
		return
	}

	if tp.Extra == "" {
		tp.response.SendError(ErrorMissingID)
		return

	}

	b := models.Boost{}

	if !b.IsBoost(tp.Extra) {
		tp.response.SendError("Is not a boost")
		return
	}

	p := models.Point{}

	if err := p.GetByUserID(tp.db, tp.currentUser.ID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if valid := b.IsPointsValid(tp.Extra, p.Total); !valid {
		tp.response.SendError("Not enough points")
		return
	}

	if exists, err := b.ExistsForVideo(tp.db, tp.ID); exists || err != nil {
		if err != nil {
			tp.response.SendError(err.Error())
			return
		}

		tp.response.SendError("boost is not available for this video")
		return
	}

	b.BoostType = tp.Extra
	b.UserID = tp.currentUser.ID
	b.VideoID = tp.ID

	if err := b.Create(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(b)
}

func (tp *TaskParams) HandleCategoryTasks() {
	switch tp.Action {
	case taskAction.top:
		tp.retrieveTopCategories()
	case taskAction.get:
		tp.retrieveMainCategories()
	case taskAction.create:
		tp.response.SendError(ErrorUnAuthorized)
	case taskAction.update:
		tp.response.SendError(ErrorUnAuthorized)
	default:
		tp.response.SendError(ErrorActionIsNotSupported)
	}
}

func (tp *TaskParams) retrieveTopCategories() {

	page := util.ConvertPageParamsToInt(tp.Extra)

	category := models.Category{}
	categories, err := category.GetTopCategories(tp.db, page)

	if err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(categories)

}

func (tp *TaskParams) retrieveMainCategories() {
	category := models.Category{}
	categories, err := category.GetMainCategories(tp.db)

	if err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(categories)
}

func (tp *TaskParams) HandleCommentTasks() {
	if tp.ID == 0 {
		tp.response.SendError(ErrorMissingID)
		return
	}

	switch tp.Action {
	case taskAction.update:
		tp.updateComment()
	case taskAction.delete:
		tp.deleteComment()
	default:
		tp.response.SendError(ErrorActionIsNotSupported)
	}
}

func (tp *TaskParams) updateComment() {
	comment := models.Comment{}

	if err := json.Unmarshal([]byte(tp.Extra), &comment); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if err := comment.Update(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if err := comment.Publisher.GetUser(tp.db, comment.UserID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(comment)
	return

}

func (tp *TaskParams) deleteComment() {
	comment := models.Comment{}

	if err := comment.Get(tp.db, tp.ID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	comment.IsActive = false

	if err := comment.Update(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(comment)
	return
}

// Handle users bio tasks
func (tp *TaskParams) HandleBioTasks() {
	if tp.ID == 0 {
		tp.response.SendError(ErrorMissingID)
		return
	}

	switch tp.Action {
	case taskAction.update:
		tp.updateUsersBio()
	default:
		tp.response.SendError(ErrorActionIsNotSupported)
	}

}

func (tp *TaskParams) updateUsersBio() {
	if tp.Extra == "" {
		tp.response.SendError(ErrorMissingExtra)
		return
	}

	bio := models.Bio{}

	if err := json.Unmarshal([]byte(tp.Extra), &bio); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if err := bio.Update(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(bio)

}

// Perform tasks for video
func (tp *TaskParams) HandleVideoTasks() {
	if tp.ID == 0 {
		tp.response.SendError(ErrorMissingID)
		return
	}

	switch tp.Action {
	case taskAction.upvote:
		tp.performVideoUpvote()
	case taskAction.downvote:
		tp.performVideoDownvote()
	case taskAction.update:
	case taskAction.delete:
		tp.performVideoDelete()
	case taskAction.get:
		tp.performVideoGet()
	default:
		tp.response.SendError(ErrorActionIsNotSupported)
	}

}

func (tp *TaskParams) performVideoGet() {
	var video models.Video
	var vote models.Vote
	var user models.ProfileUser
	var err error

	if err := video.GetVideoByID(tp.db, tp.ID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if video.IsUpvoted, err = vote.HasUpVoted(tp.db, tp.currentUser.ID, video.ID, 0); err != nil {
		return
	}

	if video.IsDownvoted, err = vote.HasDownVoted(tp.db, tp.currentUser.ID, video.ID, 0); err != nil {
		return
	}

	if err := user.GetUser(tp.db, video.UserID); err != nil {
		return
	}

	video.Publisher = user

	tp.response.SendSuccess(video)
}

func (tp *TaskParams) performVideoDelete() {
	video := models.Video{}

	if err := video.GetVideoByID(tp.db, tp.ID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if tp.currentUser.ID != video.UserID {
		tp.response.SendError(ErrorUnauthorizedAction)
		return
	}

	if err := video.SoftDelete(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	competition := models.Competitor{}
	event := models.Event{}

	if err := competition.GetByVideoID(tp.db, video.ID); err != nil {
		log.Println(err)

	}

	if err := event.Get(tp.db, competition.EventID); err != nil {
		log.Println(err)

	}

	event.CompetitorsCount--

	if err := event.Update(tp.db); err != nil {
		log.Println(err)

	}

	tp.response.SendSuccess("video deleted")
}

// Add an upvote for a user to a video
func (tp *TaskParams) performVideoUpvote() {
	vote := models.Vote{}

	if exists, err := vote.Exists(tp.db, tp.currentUser.ID, tp.ID); exists || err != nil {
		if err == nil {
			err = vote.Errors(models.ErrorExists, "id")
		}

		tp.response.SendError(err.Error())
		return
	}

	vote.VideoID = tp.ID
	vote.UserID = tp.currentUser.ID
	vote.Upvote = 1

	video := models.Video{}
	point := models.Point{}

	if err := point.GetByUserID(tp.db, tp.currentUser.ID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if err := video.GetVideoByID(tp.db, vote.VideoID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	var pointsGained models.PointActivity
	if video.Upvotes > video.Downvotes {
		point.AddPoints(models.POINT_ACTIVITY_CORRECT_VOTE)
		pointsGained = models.POINT_ACTIVITY_CORRECT_VOTE
	} else if video.Upvotes == video.Downvotes {
		point.AddPoints(models.POINT_ACTIVITY_TIE_VOTE)
		pointsGained = models.POINT_ACTIVITY_TIE_VOTE
	} else {
		point.AddPoints(models.POINT_ACTIVITY_INCORRECT_VOTE)
		pointsGained = models.POINT_ACTIVITY_INCORRECT_VOTE
	}

	point.Update(tp.db)

	if err := vote.Create(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	compete := models.Competitor{}

	if err := compete.GetByVideoID(tp.db, tp.ID); err != nil && err != sql.ErrNoRows {
		tp.response.SendError(err.Error())
		return
	}

	if compete.IsVoteUpdateable() {

		if err := compete.AddUpvote(tp.db); err != nil {
			tp.response.SendError(err.Error())
			return
		}
		log.Println("vote added for competitor", compete.ID)

	} else {
		log.Println("Unable to add any more votes for this event", compete.ID)

	}

	//Send push notification to video publisher
	if tp.currentUser.ID != compete.UserID {
		models.Notify(tp.db, tp.currentUser.ID, video.UserID, models.VERB_UPVOTED, vote.VideoID, models.OBJECT_VIDEO)
	}

	tp.response.Info = util.ConvertToString(pointsGained.Value())
	tp.response.SendSuccess(vote)
}

// Add a downvote for a user to a video
func (tp *TaskParams) performVideoDownvote() {
	vote := models.Vote{}

	if exists, err := vote.Exists(tp.db, tp.currentUser.ID, tp.ID); exists || err != nil {
		if err == nil {
			err = vote.Errors(models.ErrorExists, "id")
		}

		tp.response.SendError(err.Error())
		return
	}

	vote.VideoID = tp.ID
	vote.UserID = tp.currentUser.ID
	vote.Downvote = 1

	video := models.Video{}
	point := models.Point{}

	if err := point.GetByUserID(tp.db, tp.currentUser.ID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if err := video.GetVideoByID(tp.db, vote.VideoID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	var pointsGained models.PointActivity

	log.Printf("video.downvotes: %v video.upvotes: %v", video.Downvotes, video.Upvotes)

	totalVotes := video.Upvotes + video.Downvotes

	upVotePercentage := (float32(video.Upvotes) / float32(totalVotes)) * 100

	if upVotePercentage < 50 {
		point.AddPoints(models.POINT_ACTIVITY_CORRECT_VOTE)
		pointsGained = models.POINT_ACTIVITY_CORRECT_VOTE
	} else if video.Upvotes == video.Downvotes {
		point.AddPoints(models.POINT_ACTIVITY_TIE_VOTE)
		pointsGained = models.POINT_ACTIVITY_TIE_VOTE
	} else {
		point.AddPoints(models.POINT_ACTIVITY_INCORRECT_VOTE)
		pointsGained = models.POINT_ACTIVITY_INCORRECT_VOTE
	}

	point.Update(tp.db)

	if err := vote.Create(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	compete := models.Competitor{}

	if err := compete.GetByVideoID(tp.db, tp.ID); err != nil && err != sql.ErrNoRows {
		tp.response.SendError(err.Error())
		return
	}

	if compete.IsVoteUpdateable() {

		if err := compete.AddDownvote(tp.db); err != nil {
			tp.response.SendError(err.Error())
			return
		}

		log.Println("vote added for competitor", compete.ID)

	} else {
		log.Println("Unable to add any more votes for this event", compete.ID)

	}

	tp.response.Info = util.ConvertToString(pointsGained.Value())
	tp.response.SendSuccess(vote)
}

// Perform tasks for users
func (tp *TaskParams) HandleUserTasks() {
	if tp.ID == 0 {
		tp.response.SendError(ErrorMissingID)
		return
	}

	switch tp.Action {
	case taskAction.logout:
		tp.performUserLogout()
	case taskAction.follow:
		tp.performFollowOtherUser()
	case taskAction.updateFCM:
		tp.performUpdateFCM()
	case taskAction.unfollow:
		tp.performUnfollowOtherUser()
	case taskAction.accountType:
		tp.performUpdateAccountType()
	case taskAction.get:
		tp.performUserGet()
	case taskAction.exists:
		tp.performUserExists()

	default:
		tp.response.SendError(ErrorActionIsNotSupported)
	}
}

const (
	UsernameExists    = 1
	EmailExists       = 2
	PhoneNumberExists = 3
)

func (tp *TaskParams) performUserExists() {
	if len(tp.Extra) == 0 {
		tp.response.SendError(ErrorMissingExtra)
		return
	}

	switch tp.ID {
	case UsernameExists:
		tp.performCheckOnUserName()
	case EmailExists:
		tp.performCheckOnEmail()
	case PhoneNumberExists:
		tp.performCheckOnPhoneNumber()
	default:
		tp.response.SendError(ErrorActionIsNotSupported)

	}
}

func (tp *TaskParams) performCheckOnUserName() {
	user := models.User{}

	exists, err := user.NameExists(tp.db, tp.Extra)

	if err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(exists)

}

func (tp *TaskParams) performCheckOnEmail() {
	user := models.User{}

	exists, err := user.EmailExists(tp.db, tp.Extra)

	if err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(exists)
}

func (tp *TaskParams) performCheckOnPhoneNumber() {
	user := models.ContactInformation{}

	exists := user.ExistsPhone(tp.db, tp.Extra)

	tp.response.SendSuccess(exists)
}

func (tp *TaskParams) performUserGet() {
	switch tp.Extra {
	case "total_mob":
		tp.getTotalMob()
	case "total_talent":
		tp.getTotalTalent()
	default:
		tp.response.SendError(ErrorActionIsNotSupported)
		return

	}
}

func (tp *TaskParams) getTotalMob() {
	u := models.User{}

	total, err := u.TotalMobCount(tp.db)

	if err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(total)
}

func (tp *TaskParams) getTotalTalent() {
	u := models.User{}

	total, err := u.TotalTalentCount(tp.db)

	if err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(total)
}

/**
Create a relationship if not found
*/
func (tp *TaskParams) performFollowOtherUser() {
	relationship := models.Relationship{}

	if err := relationship.New(tp.db, tp.ID, tp.currentUser.ID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(relationship)
}

/**
Will validate if a relation exists. Its important their is a relationship existing to unfollow a user
*/
func (tp *TaskParams) performUnfollowOtherUser() {
	relationship := models.Relationship{}

	if exists, err := relationship.Exists(tp.db, tp.ID, tp.currentUser.ID); !exists || err != nil {
		if err == nil {
			err = errors.New("relationship does not exist")
		}

		tp.response.SendError(err.Error())
		return
	}

	if err := relationship.Get(tp.db, tp.ID, tp.currentUser.ID); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	relationship.IsActive = false

	if err := relationship.Update(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(relationship)

}

func (tp *TaskParams) performUpdateFCM() {
	if tp.Extra == "" {
		tp.response.SendError(ErrorMissingExtra)
		return
	}

	userApi := models.Api{}

	if err := json.Unmarshal([]byte(tp.Extra), &userApi); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if userApi.UserID != tp.currentUser.ID {
		tp.response.SendError(ErrorUnauthorizedAction)
		return
	}

	api := models.Api{}

	if err := api.GetByAPIToken(tp.db, tp.currentUser.Api.Token); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	api.PushNotificationToken = userApi.PushNotificationToken
	api.PushNotificationService = userApi.PushNotificationService
	api.ManufacturerVersion = userApi.ManufacturerVersion
	api.ManufacturerModel = userApi.ManufacturerModel
	api.ManufacturerName = userApi.ManufacturerName
	api.DeviceID = userApi.DeviceID

	if !api.IsPushServiceValid() {
		tp.response.SendError("Push Notification can only support apple or google")
		return
	}

	if err := api.Update(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(api)
}

func (tp *TaskParams) performUpdateAccountType() {
	if tp.ID == 0 {
		tp.ID = 2
	}

	tp.currentUser.AccountType = int(tp.ID)
	if err := tp.currentUser.Update(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(tp.currentUser)

}

// Log current user out of access
func (tp *TaskParams) performUserLogout() {
	if tp.Extra == "" {
		tp.response.SendError(ErrorMissingExtra)
		return
	}

	userApi := models.Api{}

	if err := userApi.GetByAPIToken(tp.db, tp.Extra); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	if userApi.UserID != tp.currentUser.ID {
		tp.response.SendError(ErrorUnauthorizedAction)
		return
	}

	if err := userApi.Delete(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(userApi)
}

// Perform tasks for vote
func (tp *TaskParams) HandleVoteTasks() {
	if tp.ID == 0 {
		tp.response.SendError(ErrorMissingID)
		return
	}

	switch tp.Action {
	case taskAction.update:
	case taskAction.delete:
	default:
		tp.response.SendError(ErrorActionIsNotSupported)
	}
}

// Perform view tasks
func (tp *TaskParams) HandleViewTasks() {
	if tp.ID == 0 {
		tp.response.SendError(ErrorMissingID)
		return
	}

	switch tp.Action {
	case taskAction.create:
		tp.performCreateView()
	default:
		tp.response.SendError(ErrorActionIsNotSupported)
	}
}

// Perform create a view

func (tp *TaskParams) performCreateView() {
	view := models.View{}

	if exists, err := view.Exists(tp.db, tp.currentUser.ID, tp.ID); exists || err != nil {
		if err == nil {
			err = errors.New("view already exists")
		}

		tp.response.SendError(err.Error())
		return
	}

	view.UserID = tp.currentUser.ID
	view.VideoID = tp.ID

	if err := view.Create(tp.db); err != nil {
		tp.response.SendError(err.Error())
		return
	}

	tp.response.SendSuccess(view)

}
