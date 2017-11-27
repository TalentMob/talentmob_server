package main

import (
	"os"
	"github.com/rathvong/talentmob_server/api"
	"github.com/rathvong/talentmob_server/system"


)

// initialise Database
var db *system.DB

var (
	//AWS DB URL
	databaseURL = os.Getenv("DATABASE_TALENTMOB_TESTING")


	// Heroku DB URL
	herokuDatabaseUrl = os.Getenv("DATABASE_URL")
)


func main() {


	if herokuDatabaseUrl == "" {
		panic("DATABASE_URL does not exist")
	}

	db = system.Connect(herokuDatabaseUrl)


	defer db.Close()

	server := api.Server{Db:db}
	server.Serve()

}


