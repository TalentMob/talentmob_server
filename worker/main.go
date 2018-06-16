package main

import (
	"github.com/rathvong/talentmob_server/system"

	"github.com/rathvong/talentmob_server/elastictranscoderapi"
	"os"
)

const (
	AWS_ENVIRONMENT_DATABASE_URL    = "DATABASE_AWS"
	HEROKU_ENVIRONMENT_DATABASE_URL = "DATABASE_URL"
)


// Initialized database url set in environment
var (
	//AWS DB URL
	awsDatabaseURL = os.Getenv(AWS_ENVIRONMENT_DATABASE_URL)
)

var AWS_CONFIG = awsDatabaseURL

func main(){

	db := system.Connect("postgres://aa172wwch662fm.cnnnwjq8tvcc.us-east-1.rds.amazonaws.com:5432/talentmob_testing?user=Rath&password=talentmob123")
	defer db.Close()

	service := elastictranscoderapi.Service{}
	service.Serve(8010, db)
}
