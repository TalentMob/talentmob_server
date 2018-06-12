package main

import (
	"github.com/rathvong/talentmob_server/system"
	"github.com/rathvong/talentmob_server/api/elastictranscoderapi"
)

func main(){

	db := system.Database()
	defer db.Close()

	service := elastictranscoderapi.Service{}

	service.Serve(8080, db)
}
