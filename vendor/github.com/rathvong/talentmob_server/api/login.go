package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"context"
	"log"

	 "firebase.google.com/go"
	_ "firebase.google.com/go/auth"
	 "google.golang.org/api/option"
	"github.com/rathvong/talentmob_server/models"
)


// HTTP POST - handle for all user registrations request with email
func (s *Server) UserRegistrations(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	response.SendSuccess("testing")
}


// HTTP POST - handle login request with email
func (s *Server) UserLogin(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)


	response.SendSuccess("testing")
}


// HTTP POST - handle login request with facebook
// * The facebook login is setup for testing only
// it is not ready for production because to properly
// secure facebook users it is recommended to include a graph
// api to validate facebook user keys
//
//	FacebookID        string `json:"facebook_id"`
//	Avatar            string `json:"avatar"`
//	Name              string `json:"name"`
//	Email             string `json:"email"`
//	AccountType       int    `json:"account_type"`
//	MinutesWatched    uint64 `json:"minutes_watched"`
//	Points            uint64 `json:"points"`
//	Password          string `json:"password, omitempty"`
//
//
func (s *Server) UserFacebookLogin(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)
	isAuthenticated, _ := s.AuthenticateHeadersForJWT(r)

	if !isAuthenticated {
		response.SendError(models.ErrorUnauthorized + " AuthenticatedHeaderForJWT()")
		return
	}

	user := models.User{}

	r.DecodeJsonPayload(&user)

	currentUser := user

	user.Api.GenerateAccessToken()

	if exists, err := user.FacebookIDExists(s.Db, user.FacebookID); !exists || err != nil {
		if err != nil {
			response.SendError(err.Error() + " FacebookIDExists()")
			return
		}

		user.GeneratePassword()

		if err = user.Create(s.Db); err != nil {
			response.SendError(err.Error() + " user.Create()")
			return
		}


		if err = user.Bio.Get(s.Db, user.ID); err != nil {
			response.SendError(err.Error() + " Bio.Get()")
			return
		}

		response.SendSuccess(user)
		return
	}

	if err := currentUser.GetByFacebookID(s.Db, user.FacebookID); err != nil {
		response.SendError(err.Error() + " getByFacebookID()")
		return
	}

	if err := s.updateAvatar(user, &currentUser); err != nil {
		response.SendError(err.Error() + " updateAvatar()")
		return
	}


	if err := s.Login(&currentUser); err != nil {
		response.SendError(err.Error() + " Login()")
		return
	}

	response.SendSuccess(currentUser)
}


func (s *Server) UserPhoneNumberLogin(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	idToken, err := s.AuthenticateHeaderForIDToken(r)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	opt := option.WithCredentialsFile("config/google-services.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	token, err := client.VerifyIDToken(idToken)
	if err != nil {
		log.Fatalf("error verifying ID token: %v\n", err)
	}

	log.Printf("Verified ID token: %v\n", token)

	u, err := client.GetUser(context.Background(), token.UID)

	if err != nil {
		log.Fatalf("FireBase.GetUser() Error -> %v", err)
	}

	ci := models.ContactInformation{}

	if exists := ci.ExistsPhone(s.Db, u.PhoneNumber); exists {
		user := models.User{}

		if err = s.Login(&user); err != nil {
			response.SendError(err.Error())
			return
		}

		response.SendSuccess(user)

		return
	}


	user := models.User{}
	user.GenerateUserName()
	user.Email = u.PhoneNumber
	user.AccountType = models.ACCOUNT_TYPE_MOB
	user.Avatar = "https://d2akrl70m8vory.cloudfront.net/default_profile_medium"
	user.GeneratePassword()
	user.Api.GenerateAccessToken()


	if err  = user.Create(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	if err = user.Bio.Get(s.Db, user.ID); err != nil {
		response.SendError(err.Error())
		return
	}

	ci.UserID = user.ID
	ci.PhoneNumber = u.PhoneNumber


	if err = ci.Create(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(user)



}

func (s *Server) updateAvatar(user models.User, currentUser *models.User)(err error){
	currentUser.Avatar = user.Avatar
	currentUser.Api = user.Api
	return currentUser.Update(s.Db)
}

// save a new api for the user to use for access
func (s *Server) Login(user *models.User) (err error){
	user.Api.UserID = user.ID

	if err = user.Bio.Get(s.Db, user.ID); err != nil {
		return
	}

	return user.Api.Create(s.Db)
}


func (s *Server) GetLastWeeksWinner(w rest.ResponseWriter, r *rest.Request){

	response := models.BaseResponse{}
	response.Init(w)

	isAuthenticated, _ := s.AuthenticateHeadersForJWT(r)

	if !isAuthenticated {
		response.SendError(models.ErrorUnauthorized + " AuthenticatedHeaderForJWT()")
		return
	}

	tp := TaskParams{}

	var user models.User

	user.ID = 999999999

	tp.Init(&response, &user, s.Db)
	tp.db = s.Db



	tp.HandleGetWinnerLastClosedEvent()
}
