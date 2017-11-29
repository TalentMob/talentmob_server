package main

import (
	"os"
	"github.com/rathvong/talentmob_server/api"
	"github.com/rathvong/talentmob_server/system"


)

// Key strings for environment variables
const (
	AWS_ENVIRONMENT_DATABASE_URL = "DATABASE_TALENTMOB_TESTING"
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


	if setDatabaseUrl() == "" {
		panic("database url does not exist in environment")
	}


	db = system.Connect(setDatabaseUrl())


	defer db.Close()

	server := api.Server{Db:db}
	server.Serve()

}

// find and set database for server
func setDatabaseUrl() (url string){
	if herokuDatabaseUrl != "" {
		return herokuDatabaseUrl
	} else if awsDatabaseURL != "" {
		return  awsDatabaseURL
	}

	return
}


