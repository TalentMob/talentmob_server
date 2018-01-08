package models

import (
	"testing"
	"github.com/rathvong/talentmob_server/system"

	"log"
)

var db = system.Connect("postgres://aa172wwch662fm.cnnnwjq8tvcc.us-east-1.rds.amazonaws.com:5432/talentmob_testing?user=Rath&password=talentmob123")


func TestEvent_GetByTitleDate(t *testing.T) {

	date := "01/1/2018"


	e := Event{}

	if err := e.GetByTitleDate(db, EventType.LeaderBoard, date); err != nil {
		t.Error("GetByTitleDate() -> ", err)
	}


	if date != e.Title {
		t.Error("GetByTitleDate() does not match 01/1/2018 -> ", e.Title)
	}

	log.Printf("GetByTitleDate -> ID -> %v Title: %v ",e.ID, e.Title)

}

func TestEvent_BeginningOfWeekMonday(t *testing.T) {
	date := "2018-01-01 00:00:00 -0800 PST"

	e := Event{}

	monday := e.BeginningOfWeekMonday()

	if date != monday.String() {
		t.Errorf("BeginningOfWeekMonday() date does not match 2018-01-01 00:00:00 -0800 PST != %v", monday.String())
	}

	log.Println("BeginningOfWeekMonday -> Passed" , monday.String())
}

func TestEvent_GetAvailableEvent(t *testing.T) {


	e := Event{}

	date := e.BeginningOfWeekMonday()
	formattedDate := date.Format(EventDateLayout)


	if err := e.GetAvailableEvent(db); err != nil {
		t.Error("GetAvailableEvent() ", err)
	}

	if formattedDate != e.Title {
		t.Errorf("GetAvailableEvent() Date does not match %v != %v", formattedDate, e.Title)

	}

}