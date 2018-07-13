package googlepublishing

import (
	"fmt"
	"os"
)

const (
	URLGooglePublishingAPI = ""
)

var GoogleAPIToken = os.Getenv("GOOGLE_API_TOKEN")

func ValidatePurchase(token string) error {

	qry := fmt.Sprintf("%s?token=%skey=%s", URLGooglePublishingAPI, token, GoogleAPIToken)

	talenthttp.Request()

	return nil
}
