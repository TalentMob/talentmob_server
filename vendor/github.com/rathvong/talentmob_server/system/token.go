package system

import (
	"os"
	"github.com/dgrijalva/jwt-go"
	"log"
)

func UserIsAuthenticated(t string) bool {
	token, err := jwt.Parse(t, getKey)

	if err != nil {
		log.Println("ParseJWTToken -> ", err)
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Println(claims["sub"])
		log.Println("ParseJWTToken ->  valid")

		return true

	} else {
		log.Println(claims["sub"])
		log.Println("ParseJWTToken -> not valid")


		return false
	}

}

func getKey(token *jwt.Token) (interface{}, error) {
	key := os.Getenv("TALENTMOB_API_KEY")

	return []byte(key), nil
}
