package models

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/rathvong/talentmob_server/system"
	"github.com/stretchr/testify/assert"
)

var db *system.DB

func init() {
	db = system.Connect(fmt.Sprintf("client_encoding=UTF8 host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", "0.0.0.0", 5432, "root", "password", "talent", "disable"))

	exists, err := checkIfTableExists()

	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		if err := createTable(); err != nil {
			log.Fatal(err)
		}
	}
}

func createTable() error {
	_, err := db.Exec(`CREATE TABLE events (
		id SERIAL PRIMARY KEY,
		start_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
		end_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
		title CHARACTER VARYING NOT NULL DEFAULT '',
		description CHARACTER VARYING NOT NULL DEFAULT '',
		event_type CHARACTER VARYING NOT NULL DEFAULT '',
		is_active BOOLEAN DEFAULT TRUE,
		competitors_count INTEGER DEFAULT 0,
		upvotes_count INTEGER DEFAULT 0,
		downvotes_count INTEGER DEFAULT 0,
		created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
		prize_pool INTEGER DEFAULT 0);`)

	if err != nil {
		return err
	}

	return nil
}

func checkIfTableExists() (bool, error) {

	var exists bool
	err := db.QueryRow(`SELECT EXISTS (
		SELECT 1
		FROM   information_schema.tables 
		WHERE  table_schema = 'public'
		AND    table_name = 'events'
		);`).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func TestEvent_GetByTitleDate(t *testing.T) {

	date := "01/1/2018"

	e := Event{}

	if err := e.GetByTitleDate(db, EventType.LeaderBoard, date); err != nil && sql.ErrNoRows != err {
		t.Error("GetByTitleDate() -> ", err)
	}

	assert.Equal(t, date, e.Title)
	if date != e.Title {
		t.Error("GetByTitleDate() does not match 01/1/2018 -> ", e.Title)
	}

	log.Printf("GetByTitleDate -> ID -> %v Title: %v ", e.ID, e.Title)

}

func TestEvent_BeginningOfWeekMonday(t *testing.T) {
	date := "2018-01-08 00:00:00 -0800 PST"

	e := Event{}

	monday := e.BeginningOfWeekMonday()

	if date != monday.String() {
		t.Errorf("BeginningOfWeekMonday() date does not match 2018-01-01 00:00:00 -0800 PST != %v", monday.String())
	}

	log.Println("BeginningOfWeekMonday -> Passed", monday.String())
}

func TestEvent_GetAvailableEvent(t *testing.T) {

	e := Event{}

	date := e.BeginningOfWeekMonday()
	formattedDate := date.Format(EventDateLayout)

	if err := e.GetAvailableWeeklyEvent(db); err != nil {
		t.Error("GetAvailableEvent() ", err)
	}

	if formattedDate != e.Title {
		t.Errorf("GetAvailableEvent() Date does not match %v != %v", formattedDate, e.Title)
	}

	t.Logf("%+v", e)

}
