package api

import (

	"github.com/ant0ine/go-json-rest/rest"
	"net/http"
	"log"

	"net/url"
	"github.com/rathvong/util"
	"os"
	"errors"
	"github.com/rathvong/talentmob_server/system"
	"github.com/rathvong/talentmob_server/models"
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
const(
	Version                   = "1"
	UrlMakeHandle             = "/"
	UrlPostUserRegistration   = "/api/"+ Version + "/u/registration"
	UrlPostUserLogin          = "/api/" + Version + "/u/login"
	UrlPostUserFacebookLogin  = "/api/" + Version + "/u/facebook"
	UrlPostUserUpdate         = "/api/" + Version + "/u/update"
	UrlGetUserImportedVideos  = "/api/" + Version + "/u/videos/imported/:params"
	UrlGetUserFavouriteVideos = "/api/" + Version + "/u/videos/favourite/:params"
	UrlGetUserProfile 		  = "/api/" + Version + "/u/:params"
	UrlGetTimeLine            = "/api/" + Version + "/time-line/:params"
	UrlGetHistory             = "/api/" + Version + "/history/:params"
	UrlGetLeaderBoard         = "/api/" + Version + "/leaderboard/:params"
	URLGetLeaderBoardHistory  = "/api/" + Version + "/leaderboard/history/:params"
	UrlGetEvents 			  = "/api/" + Version + "/events/:params"
	UrlPostVideo              = "/api/" + Version + "/video"
	UrlGetComments			  = "/api/" + Version + "/comments/:params"
	UrlPostComment			  = "/api/" + Version + "/comments"
	UrlPostPerformTask        = "/api/" + Version + "/tasks"
	UrlGetDiscovery           = "/api/" + Version + "/discovery/:params"
	DefaultAddressPort        = "8080"

)

// Server to handle micro services.
// will hold a reference to database
// for all DB calls
type Server struct {
	Db *system.DB
}

// The address port used to connect to REST service
func (s *Server) getAddressPort() (string){
	port := os.Getenv("PORT")

	if port == ""{
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
		rest.Post(UrlPostUserUpdate, s.PostUpdateUser)	,
		rest.Get(UrlGetUserImportedVideos, s.GetImportedVideos),
		rest.Get(UrlGetUserFavouriteVideos, s.GetFavouriteVideos),
		rest.Get(UrlGetUserProfile, s.GetProfile),
		rest.Get(UrlGetTimeLine, s.GetTimeLine),
		rest.Get(UrlGetHistory, s.GetHistory),
		rest.Get(UrlGetLeaderBoard, s.GetLeaderBoard),
		rest.Get(URLGetLeaderBoardHistory, s.GetLeaderBoardHistory),
		rest.Get(UrlGetEvents, s.GetEvents),
		rest.Post(UrlPostVideo, s.PostVideo),
		rest.Get(UrlGetComments, s.GetComments),
		rest.Post(UrlPostComment, s.PostComment),
		rest.Post(UrlPostPerformTask, s.PostPerformTask),
		rest.Get(UrlGetDiscovery, s.HandleQueries),
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
func (s *Server) AuthenticateHeadersForJWT(r *rest.Request) (isAuthenticated bool, token string){
	token = r.Header.Get("Authorization")

	return system.UserIsAuthenticated(token), token
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

	return
}


// validate and return a user
func (s *Server) LoginProcess(response models.BaseResponse,r *rest.Request) (currentUser models.User, err error){
	isAuthenticated, currentUser, err :=  s.AuthenticateHeaderForUser(r)

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
func (s *Server) GetPageFromParams(r *rest.Request) (page int){
	params := r.PathParam("params")
	values,_ := url.ParseQuery(params)

	pageString := values.Get("page")

	return util.ConvertPageParamsToInt(pageString)
}


// parse for week interval in params
func (s *Server) GetWeekIntervalFromParams(r *rest.Request) (interval int){
	params := r.PathParam("params")
	values,_ := url.ParseQuery(params)

	week := values.Get("week")

	return util.ConvertPageParamsToInt(week)
}


// parse for user_id in params
func (s *Server) GetUserIDFromParams(r *rest.Request) (userID uint64, err error){
	params := r.PathParam("params")
	values,_ := url.ParseQuery(params)

	pageString := values.Get("user_id")

	return util.ConvertToUint64(pageString)
}


// parse for video_id in params
func (s *Server) GetVideoIDFromParams(r *rest.Request) (userID uint64, err error){
	params := r.PathParam("params")
	values,_ := url.ParseQuery(params)

	pageString := values.Get("video_id")

	return util.ConvertToUint64(pageString)
}

// parse for event_id  in params
func (s *Server) GetEventIDFromParams(r *rest.Request) (eventID uint64, err error){
	params := r.PathParam("params")
	values,_ := url.ParseQuery(params)

	pageString := values.Get("event_id")

	return	util.ConvertToUint64(pageString)

}


// parse query  in params
func (s *Server) GetQueryFromParams(r *rest.Request) (param string){
	params := r.PathParam("params")
	values,_ := url.ParseQuery(params)

	param = values.Get("query")

	return
}


func (s *Server) GetQueryTypeFromParams(r *rest.Request) (param string){
	params := r.PathParam("params")
	values,_ := url.ParseQuery(params)

	param = values.Get("query_type")

	return
}

func (s *Server) GetCategoryFromParams(r *rest.Request) (param string){
	params := r.PathParam("params")
	values,_ := url.ParseQuery(params)

	param = values.Get("categories")

	return
}



