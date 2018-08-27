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

	err := r.DecodeJsonPayload(&transaction)

	if err != nil {
		response.SendError(err.Error())
		return
	}

	isValid, err := googlepublishing.ValidatePurchase(transaction.ItemID, transaction.PurchaseID)

	if !isValid || err != nil {

		if err != nil {
			response.SendError(err.Error())
			return
		}

		response.SendError("purchase is not valid")
		return
	}

	transaction.PurchaseState = models.PurchaseStatePurchase

	if err = transaction.Create(s.Db); err != nil {
		response.SendError(err.Error())
		return
	}

	response.SendSuccess(transaction)
}

func (s *Server) GetTransactions(w rest.ResponseWriter, r *rest.Request) {

}
