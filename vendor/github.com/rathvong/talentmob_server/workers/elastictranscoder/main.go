package main

import (
	"github.com/rathvong/talentmob_server/system"

)

func main(){

	db := system.Database()
	defer db.Close()

}
