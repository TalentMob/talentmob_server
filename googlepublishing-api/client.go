package googlepublishing

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rathvong/talentmob_server/talent/http"
)

var GoogleAPIToken = os.Getenv("FCM_SERVER_KEY")

func ValidatePurchase(productID, token string) (bool, error) {

	qry := fmt.Sprintf("https://www.googleapis.com/androidpublisher/v3/applications/%s/purchases/products/%s/tokens/%s?key=%s", "com.talentmob.talentmob", productID, token, GoogleAPIToken)

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

	return false, err
}
