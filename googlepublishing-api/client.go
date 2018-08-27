package googlepublishing

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rathvong/talentmob_server/models"
	"github.com/rathvong/talentmob_server/talent/http"
)

var GoogleAPIToken = os.Getenv("FCM_SERVER_KEY")

// consumptionState	integer	The consumption state of the inapp product. Possible values are:
// Yet to be consumed
// Consumed
// developerPayload	string	A developer-specified string that contains supplemental information about an order.
// kind	string	This kind represents an inappPurchase object in the androidpublisher service.
// orderId	string	The order id associated with the purchase of the inapp product.
// purchaseState	integer	The purchase state of the order. Possible values are:
// Purchased
// Canceled
// purchaseTimeMillis	long	The time the product was purchased, in milliseconds since the epoch (Jan 1, 1970).
// purchaseType	integer	The type of purchase of the inapp product. This field is only set if this purchase was not made using the standard in-app billing flow. Possible values are:
// Test (i.e. purchased from a license testing account)
// Promo (i.e. purchased using a promo code)

func ValidatePurchase(transaction *models.Transaction) error {

	qry := fmt.Sprintf("https://www.googleapis.com/androidpublisher/v3/applications/%s/purchases/products/%s/tokens/%s?key=%s", "com.talentmob.talentmob", transaction.ItemID, transaction.PurchaseID, GoogleAPIToken)

	type Response struct {
		Kind               string `json:"kind"`
		PurchaseTimeMillis uint   `json:"purchaseTimeMillis"`
		PurchaseState      int    `json:"purchaseState"`
		ConsumptionState   int    `json:"consumptionState"`
		DeveloperPayload   string `json:"developerPayload"`
		OrderID            string `json:"orderId"`
		PurchaseType       int    `json:"purchaseType"`
	}

	var response Response
	err := talenthttp.Request(http.MethodGet, qry, nil, &response)

	transaction.PurchaseState = response.PurchaseState

	return err
}
