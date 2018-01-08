package models

import (
	"testing"
	"github.com/rathvong/talentmob_server/system"

	"log"
)

func TestEvent_GetByTitleDate(t *testing.T) {

	date := "01/1/2018"

	db := system.Connect("postgres://aa172wwch662fm.cnnnwjq8tvcc.us-east-1.rds.amazonaws.com:5432/talentmob_testing?user=Rath&password=talentmob123")

	e := Event{}

	if err := e.GetByTitleDate(db, EventType.LeaderBoard, date); err != nil {
		t.Error("GetByTitleDate() -> ", err)
	}

	log.Printf("GetByTitleDate -> ID -> %v Title: %v ",e.ID, e.Title)

	if date != e.Title {
		t.Error("GetByTitleDate() does not match 01/1/2018 -> ", e.Title)
	}
}