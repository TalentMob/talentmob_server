package googlepublishing

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/rathvong/talentmob_server/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

type GoogleIAP struct {
	Kind               string `json:"kind"`
	PurchaseTimeMillis uint64 `json:"purchaseTimeMillis"`
	PurchaseState      int    `json:"purchaseState"`
	ConsumptionState   bool   `json:"consumptionState"`
	OrderId            string `json:"orderId"`
	DeveloperPayload   string `json:"developerPayload"`
}

func ValidatePurchase(transaction *models.Transaction) error {

	log.Printf("Transaction: %+v", transaction)

	// You need to prepare a public key for your Android app's in app billing
	// at https://console.developers.google.com.
	jsonKey, err := ioutil.ReadFile("config/google-publishing-api.json")
	if err != nil {
		log.Fatal(err)
	}

	conf, err := google.JWTConfigFromJSON(jsonKey, "https://www.googleapis.com/auth/androidpublisher")
	if err != nil {
		log.Fatal("conf", err)
	}

	client := conf.Client(oauth2.NoContext)

	resp, err := client.Get("https://www.googleapis.com/androidpublisher/v2/applications/com.talentmob.talentmob/purchases/" + "products" + "/" + transaction.ItemID + "/tokens/" + transaction.PurchaseID)

	body, err := ioutil.ReadAll(resp.Body)
	log.Println("response: ", string(body))

	appResult := &GoogleIAP{}
	err = json.Unmarshal(body, &appResult)

	if err != nil {
		return err
	}

	transaction.PurchaseState = int(appResult.PurchaseState)
	transaction.PurchaseTimeMilis = appResult.PurchaseTimeMillis

	return nil
}
