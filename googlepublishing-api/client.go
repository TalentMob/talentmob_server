package googlepublishing

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/awa/go-iap/playstore"

	"github.com/rathvong/talentmob_server/models"
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

	log.Printf("Transaction: %+v", transaction)

	// You need to prepare a public key for your Android app's in app billing
	// at https://console.developers.google.com.
	jsonKey, err := ioutil.ReadFile("config/google-publishing-api.json")
	if err != nil {
		log.Fatal(err)
	}

	client, err := playstore.New(jsonKey)

	if err != nil {
		return err
	}

	ctx := context.Background()

	resp, err := client.VerifySubscription(ctx, "com.talentmob.talentmob", transaction.ItemID, transaction.PurchaseID)

	if err != nil {
		log.Printf("ValidatePurchase -> response error: ", err)
		return err
	}

	transaction.PurchaseState = int(resp.PurchaseState)

	return nil
}
