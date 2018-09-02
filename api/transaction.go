package api

import (
	"log"

	"github.com/ant0ine/go-json-rest/rest"
	googlepublishing "github.com/rathvong/talentmob_server/googlepublishing-api"
	"github.com/rathvong/talentmob_server/models"
)

func (s *Server) PostTransaction(w rest.ResponseWriter, r *rest.Request) {

	var response models.BaseResponse
	var transaction models.Transaction

	response.Init(w)

	currentUser, err := s.LoginProcess(response, r)

	if err != nil {
		return
	}

	err = r.DecodeJsonPayload(&transaction)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	transaction.PurchaseState = models.PurchaseStateNotValidated

	err = googlepublishing.ValidatePurchase(&transaction)

	if err != nil {
		log.Println("ValidatePurchase: ", err)
		response.SendError(err.Error())
		return
	}

	if err = transaction.Create(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	if transaction.PurchaseState == models.PurchaseStatePurchase {
		switch transaction.ItemID {
		case "2250_star_power":
			currentUser.AddStarPower(s.Db, models.POINT_TRANSACTION_2250_STARPOWER)
		case "9500_star_power":
			currentUser.AddStarPower(s.Db, models.POINT_TRANSACTION_9500_STARPOWER)
		case "24500_star_power":
			currentUser.AddStarPower(s.Db, models.POINT_TRANSACTION_24500_STARPOWER)
		case "100k_star_power":
			currentUser.AddStarPower(s.Db, models.POINT_TRANSACTION_100000_STARPOWER)
		}
	}

	response.SendSuccess(transaction)
}

func (s *Server) GetTransactions(w rest.ResponseWriter, r *rest.Request) {

}
