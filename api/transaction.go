package api

import (
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

	err = googlepublishing.ValidatePurchase(&transaction)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	if err = transaction.Create(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	if transaction.PurchaseState == models.PurchaseStatePurchase {

	}

	response.SendSuccess(transaction)
}

func (s *Server) GetTransactions(w rest.ResponseWriter, r *rest.Request) {

}
