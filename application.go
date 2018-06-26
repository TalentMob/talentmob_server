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

// Initialized database url set in environment
var (
	//AWS DB URL
	awsDatabaseURL = os.Getenv(AWS_ENVIRONMENT_DATABASE_URL)

)


var AWS_CONFIG = awsDatabaseURL + "&sslmode=verify-full&sslrootcert=config/rds-combined-ca-bundle.pem"


func main() {

	db := system.Connect(AWS_CONFIG)
	defer db.Close()

	server := api.Server{Db: db}
	server.Serve()

}


