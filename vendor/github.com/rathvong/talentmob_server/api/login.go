package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"context"
	 "firebase.google.com/go"
	_ "firebase.google.com/go/auth"
	 "google.golang.org/api/option"
	"github.com/rathvong/talentmob_server/models"
	"fmt"
	"github.com/ahmdrz/goinsta"
	"log"
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

	if err := user.Api.RemoveOLDAPIs(s.Db, user.DeviceID); err != nil {
		log.Println("Facebook Login -> Error: ", err)
		
	}

	user.Api.GenerateAccessToken()

	if exists, err := user.FacebookIDExists(s.Db, user.FacebookID); !exists || err != nil {
		if err != nil {
			response.SendError(err.Error() + " FacebookIDExists()")
			return
		}

		user.GeneratePassword()

		if exists, err  := user.NameExists(s.Db, user.Name); exists || err != nil{
			if err != nil {
				response.SendError(err.Error())
				return
			}

			user.GenerateUserName()
		}

		user.AccountType = models.ACCOUNT_TYPE_MOB

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

	if err := s.updateApi(user, &currentUser); err != nil {
		response.SendError(err.Error() + " updateApi()")
		return
	}


	if err := s.Login(&currentUser); err != nil {
		response.SendError(err.Error() + " Login()")
		return
	}

	currentUser.IsReturning = true

	response.SendSuccess(currentUser)
}



func (s *Server) LoginWithInstagram(w rest.ResponseWriter, r *rest.Request) {
	response := models.BaseResponse{}
	response.Init(w)
	isAuthenticated, _ := s.AuthenticateHeadersForJWT(r)

	if !isAuthenticated {
		response.SendError(models.ErrorUnauthorized + " AuthenticatedHeaderForJWT()")
		return
	}

	user := models.User{}

	r.DecodeJsonPayload(&user)


	insta := goinsta.New(user.Name, user.Password)
	defer insta.Logout()

	if err := insta.Login(); err != nil {
		response.SendError(err.Error())
		return
	}


	user, err := s.createLoginForInstagram(insta.Informations)

	if err != nil {
		response.SendError(err.Error())
		return
	}


	response.SendSuccess(user)


}


func (s *Server) createLoginForInstagram(userInfo goinsta.Informations) (user models.User, err error){
	ci := models.ContactInformation{}

	user.Api.GenerateAccessToken()

	if exists := ci.ExistsInstagram(s.Db, userInfo.UUID); exists {

		if err = ci.GetInstagram(s.Db, userInfo.UUID); err != nil {
			return user, err
		}

		if err = user.Get(s.Db, ci.UserID); err != nil {

			return user, err
		}



		if err = s.Login(&user); err != nil {
			return user, err
		}

		user.IsReturning = true
		return
	}

	user.Name = userInfo.Username
	user.Email = userInfo.UUID
	user.AccountType = models.ACCOUNT_TYPE_MOB
	user.Avatar = "https://d2akrl70m8vory.cloudfront.net/default_profile_medium"
	user.GeneratePassword()

	if exists, err  := user.NameExists(s.Db, user.Name); exists || err != nil{
		if err != nil {
			return user, err
		}

		user.GenerateUserName()
	}


	if err  = user.Create(s.Db); err != nil {
		return user, err
	}

	if err = user.Bio.Get(s.Db, user.ID); err != nil {
		return user, err
	}


	ci.UserID = user.ID
	ci.PhoneNumber = userInfo.UUID
	ci.InstagramID = userInfo.UUID


	if err = ci.Create(s.Db); err != nil {
		return user, err
	}

	return
}


func (s *Server) createLoginForEmail(email string) (user models.User, err error){
	if exists, err := user.EmailExists(s.Db, email); exists || err != nil {

		if err != nil {
			return user, err
		}

		if err = user.GetByEmail(s.Db, email); err != nil {
			return user, err
		}


		user.Api.GenerateAccessToken()

		if err = s.Login(&user); err != nil {
			return user, err
		}


		user.IsReturning = true

		return user, err
	}



	user.GenerateUserName()
	user.Email = email
	user.AccountType = models.ACCOUNT_TYPE_MOB
	user.Avatar = "https://d2akrl70m8vory.cloudfront.net/default_profile_medium"
	user.GeneratePassword()
	user.Api.GenerateAccessToken()


	if err  = user.Create(s.Db); err != nil {
		return user, err
	}

	if err = user.Bio.Get(s.Db, user.ID); err != nil {
		return user, err
	}

	ci := models.ContactInformation{}

	ci.UserID = user.ID
	ci.PhoneNumber = email
	ci.InstagramID = email


	if err = ci.Create(s.Db); err != nil {
		return user, err
	}

	return user, err
}

func (s *Server) createLoginForPhone(phone string) (user models.User, err error) {

	ci := models.ContactInformation{}

	if exists := ci.ExistsPhone(s.Db, phone); exists {

		if err = ci.GetPhone(s.Db, phone); err != nil {
			return user, err
		}


		if err = user.Get(s.Db, ci.UserID); err != nil {
			return user, err
		}

		user.Api.GenerateAccessToken()

		if err = s.Login(&user); err != nil {
			return user, err
		}

		user.IsReturning = true


		return
	}



	user.GenerateUserName()
	user.Email = phone
	user.AccountType = models.ACCOUNT_TYPE_MOB
	user.Avatar = "https://d2akrl70m8vory.cloudfront.net/default_profile_medium"
	user.GeneratePassword()
	user.Api.GenerateAccessToken()


	if err  = user.Create(s.Db); err != nil {
		return user, err
	}

	if err = user.Bio.Get(s.Db, user.ID); err != nil {
		return user, err
	}

	ci.UserID = user.ID
	ci.PhoneNumber = phone
	ci.InstagramID = phone


	if err = ci.Create(s.Db); err != nil {
		return user, err
	}


	return
}

type SocialLogin struct {
	Verification string `json:"verification"`
	DeviceID string `json:"device_id"`
}


func (s *Server) UserFirebaseLogin(w rest.ResponseWriter, r *rest.Request){
	response := models.BaseResponse{}
	response.Init(w)

	verification := SocialLogin{}
	r.DecodeJsonPayload(&verification)


	idToken, err := s.AuthenticateHeaderForIDToken(r)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	opt := option.WithCredentialsFile("config/google-services.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		response.SendError("error initializing app")
		return
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		response.SendError("error getting Auth client:")
		return
	}

	token, err := client.VerifyIDToken(idToken)

	if err != nil {
		err = fmt.Errorf("error verifying ID token: %v", idToken)
		response.SendError(err.Error())
		return
	}



	u, err := client.GetUser(context.Background(), token.UID)

	if err != nil {
		response.SendError("FireBase.GetUser() Error ")
		return
	}

	var user models.User

	if err = user.Api.RemoveOLDAPIs(s.Db, verification.DeviceID); err != nil {
		log.Println("FireBaseLogin() -> Error: ", err)

	}

	switch verification.Verification {
	case "phone":
		user, err = s.createLoginForPhone(u.PhoneNumber)

	case "gmail":
		user, err = s.createLoginForEmail(u.Email)

	default:
		response.SendError(ErrorActionIsNotSupported)
		return
	}


	if err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(user)


}

func (s *Server) updateApi(user models.User, currentUser *models.User)(err error){

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
