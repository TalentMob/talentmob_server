package api

import (
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"

	"errors"
	"net/url"
	"os"

	"github.com/rathvong/talentmob_server/models"
	"github.com/rathvong/talentmob_server/system"
	"github.com/rathvong/util"
)

const (
	ErrorUnAuthorized = "unauthorized"
)

// Url address and connection port to api services
// Port is set to 8080
// POST registration - /api/1/u/registration
// POST login - /api/1/u/login
// POST facebook login - /api/1/u/facebook
// POST update user - /api/1/u/update
// GET imported videos - /api/1/u/videos/imported/
// GET favourite videos - /api/1/u/videos/favourite
// GET time-line  - /api/1/u/time-line/
// POST upload video - /api/1/video
// POST perform tasks - /api/1/tasks
const (
	Version                    = "1"
	UrlMakeHandle              = "/"
	UrlPostUserRegistration    = "/api/" + Version + "/u/registration"
	UrlPostUserLogin           = "/api/" + Version + "/u/login"
	UrlPostUserFacebookLogin   = "/api/" + Version + "/u/facebook"
	UrlPostUserFireBaseLogin   = "/api/" + Version + "/u/login/firebase"
	UrlPostUserInstagramLogin  = "/api/" + Version + "/u/login/instagram"
	UrlPostUserUpdate          = "/api/" + Version + "/u/update"
	UrlGetUserImportedVideos   = "/api/" + Version + "/u/videos/imported/:params"
	UrlGetUserFavouriteVideos  = "/api/" + Version + "/u/videos/favourite/:params"
	UrlGetUserImportedVideos2  = "/api/" + "2" + "/u/videos/imported/:params"
	UrlGetUserFavouriteVideos2 = "/api/" + "2" + "/u/videos/favourite/:params"

	UrlGetUserProfile  = "/api/" + Version + "/u/:params"
	UrlGetUserProfile2 = "/api/" + "2" + "/u/:params"

	UrlGetRelationship  = "/api/" + Version + "/u/relationships/:params"
	UrlGetRelationship2 = "/api/" + "2" + "/u/relationships/:params"

	UrlGetStats     = "/api/" + Version + "/u/stats/:params"
	UrlGetTimeLine  = "/api/" + Version + "/time-line/:params"
	UrlGetTimeLine2 = "/api/" + "2" + "/time-line/:params"

	UrlGetHistory      = "/api/" + Version + "/history/:params"
	UrlGetLeaderBoard  = "/api/" + Version + "/leaderboard/:params"
	UrlGetLeaderBoard2 = "/api/" + "2" + "/leaderboard/:params"

	URLGetLeaderBoardHistory  = "/api/" + Version + "/leaderboard/history/:params"
	URLGetLeaderBoardHistory2 = "/api/" + "2" + "/leaderboard/history/:params"

	UrlGetEvents  = "/api/" + Version + "/events/:params"
	UrlGetEvents2 = "/api/" + "2" + "/events/:params"
	UrlPostVideo  = "/api/" + Version + "/video"

	UrlGetTopVideo  = "/api/" + Version + "/video/top/"
	UrlGetTopVideo2 = "/api/" + "2" + "/video/top/"

	UrlGetVideo  = "/api/" + Version + "/video/:params"
	UrlGetVideo2 = "/api/" + "2" + "/video/:params"

	UrlGetUpVotedUsersOnVideo  = "/api/" + Version + "/video/upvote/:params"
	UrlGetUpVotedUsersOnVideo2 = "/api/" + "2" + "/video/upvote/:params"

	UrlGetComments  = "/api/" + Version + "/comments/:params"
	UrlGetComments2 = "/api/" + "2" + "/comments/:params"

	UrlPostComment     = "/api/" + Version + "/comments"
	UrlPostPerformTask = "/api/" + Version + "/tasks"
	UrlGetDiscovery    = "/api/" + Version + "/discovery/:params"
	UrlGetDiscovery2   = "/api/" + "2" + "/discovery/:params"

	UrlPostSystemTask = "/api/" + Version + "/admin/system"
	UrlGetTopUsers    = "/api/" + Version + "/history/users/:params"
	UrlGetTopUsers2   = "/api/" + "2" + "/history/users/:params"

	UrlPostElasticTranscoding = "/api/" + Version + "/elastictranscoding"
	UrlPostTransaction        = "/api/" + Version + "/starpower/transaction"
	UrlGetTransactions        = "/api/" + Version + "/starpower/transaction/:params"

	UrlPostEvent = "/api/" + Version + "/event"
	UrlGetEvent  = "/api" + Version + "/event/:params"

	DefaultAddressPort = "8080"
)

// Server to handle micro services.
// will hold a reference to database
// for all DB calls
type Server struct {
	Db              *system.DB
	AddEventChannel chan models.Event
}

func (s *Server) AddEvent(event models.Event) {
	s.AddEventChannel <- event
}

// The address port used to connect to REST service
func (s *Server) getAddressPort() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = DefaultAddressPort
	}

	return ":" + port
}

func (s *Server) Serve() {
	service := rest.NewApi()

	var DefaultDevStack = []rest.Middleware{
		&rest.AccessLogApacheMiddleware{
			Format: rest.CombinedLogFormat,
		},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.PoweredByMiddleware{},
		&rest.RecoverMiddleware{},
		&rest.GzipMiddleware{},
	}

	service.Use(DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post(UrlPostUserLogin, s.UserLogin),
		rest.Post(UrlPostUserRegistration, s.UserRegistrations),
		rest.Post(UrlPostUserFacebookLogin, s.UserFacebookLogin),
		rest.Post(UrlPostUserUpdate, s.PostUpdateUser),
		rest.Get(UrlGetUserImportedVideos, s.GetImportedVideos),
		rest.Get(UrlGetUserFavouriteVideos, s.GetFavouriteVideos),
		rest.Get(UrlGetUserImportedVideos2, s.GetImportedVideos2),
		rest.Get(UrlGetUserFavouriteVideos2, s.GetFavouriteVideos2),

		rest.Get(UrlGetUserProfile, s.GetProfile),
		rest.Get(UrlGetUserProfile2, s.GetProfile2),
		rest.Get(UrlGetRelationship, s.GetRelationships),
		rest.Get(UrlGetRelationship2, s.GetRelationships2),

		rest.Get(UrlGetTimeLine, s.GetTimeLine),
		rest.Get(UrlGetTimeLine2, s.GetTimeLine2),
		rest.Get(UrlGetHistory, s.GetHistory),
		rest.Get(UrlGetLeaderBoard, s.GetLeaderBoard),
		rest.Get(UrlGetLeaderBoard2, s.GetLeaderBoard2),
		rest.Get(URLGetLeaderBoardHistory, s.GetLeaderBoardHistory),
		rest.Get(URLGetLeaderBoardHistory2, s.GetLeaderBoardHistory2),
		rest.Get(UrlGetEvents, s.GetEvents),
		rest.Get(UrlGetEvents2, s.GetEvents2),
		rest.Post(UrlPostVideo, s.PostVideo),
		rest.Get(UrlGetComments, s.GetComments),
		rest.Get(UrlGetComments2, s.GetComments2),

		rest.Post(UrlPostComment, s.PostComment),
		rest.Post(UrlPostPerformTask, s.PostPerformTask),
		rest.Get(UrlGetDiscovery, s.HandleQueries),
		rest.Get(UrlGetDiscovery2, s.HandleQueries2),
		rest.Post(UrlPostSystemTask, s.PostPerformSystemTask),
		rest.Get(UrlGetTopUsers, s.GetTopUsers),
		rest.Get(UrlGetTopUsers2, s.GetTopUsers2),

		rest.Get(UrlGetTopVideo, s.GetLastWeeksWinner),
		rest.Get(UrlGetTopVideo2, s.GetLastWeeksWinner2),

		rest.Get(UrlGetVideo, s.GetVideo),
		rest.Get(UrlGetVideo2, s.GetVideo2),

		rest.Post(UrlPostUserFireBaseLogin, s.UserFirebaseLogin),
		//	rest.Post(UrlPostUserInstagramLogin, s.LoginWithInstagram),
		rest.Get(UrlGetUpVotedUsersOnVideo, s.GetUpVotedUsersOnVideo),
		rest.Get(UrlGetUpVotedUsersOnVideo2, s.GetUpVotedUsersOnVideo2),

		rest.Get(UrlGetStats, s.GetStats),
		rest.Post(UrlPostElasticTranscoding, s.PostElasticTranscoding),
		rest.Post(UrlPostTransaction, s.PostTransaction),
		rest.Get(UrlGetTransactions, s.GetTransactions),
		rest.Post(UrlPostEvent, s.PostEvent),
	)

	if err != nil {
		log.Fatal(err)
	}

	service.SetApp(router)

	//***** Handle API
	http.Handle(UrlMakeHandle, service.MakeHandler())
	log.Fatal(http.ListenAndServe(s.getAddressPort(), nil))

}

//Authenticated request headers for JWT
func (s *Server) AuthenticateHeadersForJWT(r *rest.Request) (isAuthenticated bool, token string) {
	token = r.Header.Get("Authorization")

	return system.UserIsAuthenticated(token), token
}

func (s *Server) AuthenticateHeaderForIDToken(r *rest.Request) (token string, err error) {
	token = r.Header.Get("Authorization")

	if token == "" {
		err = errors.New("missing ID Token")
	}

	return
}

// validated JWT token and retrieve user
func (s *Server) AuthenticateHeaderForUser(r *rest.Request) (isAuthenticated bool, user models.User, err error) {

	token := r.Header.Get("Authorization")

	isAuthenticated, err = user.Api.APITokenExists(s.Db, token)

	if !isAuthenticated || err != nil {

		log.Println("AuthenticatedHeaderForUser() authorized", isAuthenticated)
		if err == nil {
			err = errors.New(ErrorUnAuthorized)
		}

		return isAuthenticated, user, err
	}

	if err = user.Api.GetByAPIToken(s.Db, token); err != nil {
		return
	}

	if err = user.Get(s.Db, user.Api.UserID); err != nil {
		return
	}

	if !user.IsActive {
		return false, user, errors.New("user is not active")
	}

	if err = user.Bio.Get(s.Db, user.ID); err != nil {
		return
	}

	return
}

func (s *Server) AuthenticateHeaderForAdmin(r *rest.Request) (isAuthenticated bool) {

	token := r.Header.Get("Authorization")
	adminToken := os.Getenv("ADMIN_TOKEN")

	return token == adminToken

}

// validate and return a user
func (s *Server) LoginProcess(response models.BaseResponse, r *rest.Request) (currentUser models.User, err error) {
	isAuthenticated, currentUser, err := s.AuthenticateHeaderForUser(r)

	if !isAuthenticated || err != nil {
		log.Println("LoginProcess() authorized", isAuthenticated)

		if err == nil {
			err = currentUser.Errors(models.ErrorUserDoesNotExist, "token")
		}

		response.SendError(err.Error())
		return
	}

	log.Printf("LoginProcess() user logged in %v", currentUser.ID)

	return
}

// todo: Update Page params request by data structure
//  func (s *Server) ParamsString(r *rest.Request, key string) (param string)
//  func (s *Server) ParamsUint(r *rest.Request, key string) (param uint)
// 	so on and so forth...

// parse for page number in params
func (s *Server) GetPageFromParams(r *rest.Request) (page int) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	pageString := values.Get("page")

	return util.ConvertPageParamsToInt(pageString)
}

// parse for week interval in params
func (s *Server) GetWeekIntervalFromParams(r *rest.Request) (interval int) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	week := values.Get("week")

	return util.ConvertPageParamsToInt(week)
}

// parse for user_id in params
func (s *Server) GetUserIDFromParams(r *rest.Request) (userID uint64, err error) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	pageString := values.Get("user_id")

	return util.ConvertToUint64(pageString)
}

// parse for video_id in params
func (s *Server) GetVideoIDFromParams(r *rest.Request) (userID uint64, err error) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	pageString := values.Get("video_id")

	return util.ConvertToUint64(pageString)
}

// parse for event_id  in params
func (s *Server) GetEventIDFromParams(r *rest.Request) (eventID uint64, err error) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	pageString := values.Get("event_id")

	return util.ConvertToUint64(pageString)

}

// parse query  in params
func (s *Server) GetQueryFromParams(r *rest.Request) (param string) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	param = values.Get("query")

	return
}

func (s *Server) GetQueryTypeFromParams(r *rest.Request) (param string) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	param = values.Get("query_type")

	return
}

func (s *Server) GetCategoryFromParams(r *rest.Request) (param string) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	param = values.Get("categories")

	return
}

func (s *Server) GetAccountTypeFromParams(r *rest.Request) (param int, err error) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	paramString := values.Get("account_type")

	return util.ConvertStringToInt(paramString)
}

func (s *Server) GetRelationshipFromParams(r *rest.Request) (param string, err error) {
	params := r.PathParam("params")
	values, _ := url.ParseQuery(params)

	paramString := values.Get("relationship")

	if paramString == "" {
		err = errors.New("missing relationship (followers or followings)")
	}

	return paramString, err
}
