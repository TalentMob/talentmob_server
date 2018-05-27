package main

import (
	"github.com/rathvong/talentmob_server/api"
	"github.com/rathvong/talentmob_server/system"
	"os"
)

// Key strings for environment variables
const (
	AWS_ENVIRONMENT_DATABASE_URL    = "DATABASE_AWS"
	HEROKU_ENVIRONMENT_DATABASE_URL = "DATABASE_URL"
)

// initialise Database
var db *system.DB

// Initialized database url set in environment
var (
	//AWS DB URL
	awsDatabaseURL = os.Getenv(AWS_ENVIRONMENT_DATABASE_URL)
	// Heroku DB URL
	herokuDatabaseUrl = os.Getenv(HEROKU_ENVIRONMENT_DATABASE_URL)
)

func main() {

	db = system.Connect(awsDatabaseURL + "&sslmode=verify-full&sslrootcert=config/rds-combined-ca-bundle.pem")

	defer db.Close()

	server := api.Server{Db: db}
	server.Serve()

}

// find and set database for server
func setDatabaseUrl() (url string) {
	if herokuDatabaseUrl != "" {
		return herokuDatabaseUrl
	} else if awsDatabaseURL != "" {
		return awsDatabaseURL
	}

	return
}
