package models

import (
	"time"
	"github.com/rathvong/talentmob_server/system"
	"log"
	"database/sql"
	"github.com/jinzhu/now"
)

//id SERIAL PRIMARY KEY,
//start_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
//end_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
//title CHARACTER VARYING NOT NULL DEFAULT '',
//description CHARACTER VARYING NOT NULL DEFAULT '',
//is_active BOOLEAN DEFAULT TRUE,
//competitors_count INTEGER DEFAULT 0,
//upvotes_count INTEGER DEFAULT 0,
//downvotes_count INTEGER DEFAULT 0,
//created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
//updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL);

const (
	EventDateLayout = "01/2/2006"
)

type Event struct {
	BaseModel
	StartDate time.Time `json:"start_date"`
	EndDate time.Time `json:"end_date"`
	Title string `json:"title"`
	Description string `json:"description"`
	EventType string `json:"event_type"`
	IsActive bool `json:"is_active"`
	CompetitorsCount uint64 `json:"competitors_count"`
	UpvotesCount uint64 `json:"upvotes_count"`
	DownvotesCount uint64 `json:"downvotes_count"`
}

var EventType = eventType {
	LeaderBoard:"leaderboard",
	Weekly: "weekly",
	Daily:"daily",
	Hourly:"hourly"}

type eventType struct {
	LeaderBoard string `json:"leaderboard"`
	Weekly string `json:"weekly"`
	Daily string `json:"daily"`
	Hourly string `json:"hourly"`
}

func (e *Event) queryCreate() (qry string){
	return `INSERT INTO events
				(start_date,
				end_date,
				title,
				description,
				event_type,
				is_active,
				competitors_count,
				upvotes_count,
				downvotes_count,
				created_at,
				updated_at)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)

			RETURNING id
	`
}

func (e *Event) queryUpdate() (qry string){
	return `UPDATE events SET
				start_date = $2,
				end_date = $3,
				title = $4,
				description = $5,
				event_type = $6,
				is_active = $7,
				competitors_count = $8,
				upvotes_count = $9,
				downvotes_count = $10,
				updated_at = $11
				WHERE id = $1`
}

func (e *Event) querySoftDelete() (qry string){
	return `UPDATE events set
				is_active = $2
			WHERE id = $1`
}

func (e *Event) queryGetByID() (qry string){
	return `SELECT
				id,
				start_date,
				end_date,
				title,
				description,
				event_type,
				is_active,
				competitors_count,
				upvotes_count,
				downvotes_count,
				created_at,
				updated_at
			FROM events
			WHERE
				id = $1
			`
}

func (e *Event) queryGetByTitleDate() (qry string){
	return `SELECT
				id,
				start_date,
				end_date,
				title,
				description,
				event_type,
				is_active,
				competitors_count,
				upvotes_count,
				downvotes_count,
				created_at,
				updated_at
			FROM events
			WHERE

				event_type = $1
			AND
				title = $2
			`
}

func (e *Event) queryGetEvents() (qry string){
	return `SELECT
				id,
				start_date,
				end_date,
				title,
				description,
				event_type,
				is_active,
				competitors_count,
				upvotes_count,
				downvotes_count,
				created_at,
				updated_at
			FROM events
			WHERE is_active = true
			ORDER BY start_date DESC
			LIMIT $1
			OFFSET $2 `
}

func (e *Event) queryExist() (qry string){
	return `SELECT EXISTS( select 1 from events where start_date = $1 and event_type = $2 and title = $3)`
}


func (e *Event) validateCreateErrors() (err error){

	if e.StartDate.String() == "" {
		return e.Errors(ErrorMissingValue, "start_date")
	}

	if e.EndDate.String() == "" {
		return e.Errors(ErrorMissingValue, "end_date")
	}

	if e.EventType == "" {
		return e.Errors(ErrorMissingValue, "event_type")
	}

	if e.Title == "" {
		return e.Errors(ErrorMissingValue, "title")
	}

	return
}

func (e *Event) validateUpdateErrors() (err error){
	if e.ID == 0 {
		return e.Errors(ErrorMissingValue, "id")
	}


	return e.validateCreateErrors()
}

func (e *Event) Create(db *system.DB)(err error){

	if err = e.validateCreateErrors(); err != nil {
		return err
	}

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}()

	if err != nil {
		log.Println("Event.Create() Begin()", err)
		return
	}

	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	e.IsActive = true


	err = tx.QueryRow(e.queryCreate(),
			e.StartDate,
			e.EndDate,
			e.Title,
			e.Description,
			e.EventType,
			e.IsActive,
			e.CompetitorsCount,
			e.UpvotesCount,
			e.DownvotesCount,
			e.CreatedAt,
			e.UpdatedAt).Scan(&e.ID)

	if err != nil {
		log.Printf("startDate -> %v title -> %v eventType -> %v QueryRow() -> %v Error -> %v", e.StartDate.String(), e.Title, e.EventType, e.queryCreate(),err )
		return
	}

	log.Println("Event.create() Event Created -> ", e.ID)

	return
}

func (e *Event) Update(db *system.DB)(err error){

	if err := e.validateUpdateErrors(); err != nil {
		return err
	}

	tx, err := db.Begin()

	defer func(){
		if err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			log.Println("Event.Update() Commit() - ", err)
			return
		}

	}()


	if err != nil {
		log.Println("Event.Update() Begin() - ", err)
		return
	}

	e.UpdatedAt = time.Now()

	_, err = tx.Exec(e.queryUpdate(),
		e.ID,
		e.StartDate,
		e.EndDate,
		e.Title,
		e.Description,
		e.EventType,
		e.IsActive,
		e.CompetitorsCount,
		e.UpvotesCount,
		e.DownvotesCount,
		e.UpdatedAt)

	if err != nil {
		log.Printf("Event.Update() id -> %v Exec() -> %v Error -> %v", e.ID, e.queryUpdate(), err)
		return
	}

	return
}

func (e *Event) SoftDelete(db *system.DB)(err error){
	if e.ID == 0 {
		return e.Errors(ErrorMissingID, "id")
	}

	_, err = db.Exec(e.querySoftDelete(), e.ID)

	if err != nil {
		log.Printf("Event.SoftDelete() id -> %v Exec() -> %v Error -> %v", e.ID, e.querySoftDelete(), err)
		return
	}

	return
}

func (e *Event) Get(db *system.DB, eventID uint64)(err error){
	if eventID == 0 {
		return e.Errors(ErrorMissingValue, "event_id")
	}

	err = db.QueryRow(e.queryGetByID(), eventID).Scan(
		&e.ID,
		&e.StartDate,
		&e.EndDate,
		&e.Title,
		&e.Description,
		&e.EventType,
		&e.IsActive,
		&e.CompetitorsCount,
		&e.UpvotesCount,
		&e.DownvotesCount,
		&e.CreatedAt,
		&e.UpdatedAt)

	if err != nil {
		log.Printf("Event.Get() id -> %v QueryRow() -> %v Error -> %v", e.ID, e.queryGetByID(), err)
		return
	}


	return
}

func (e *Event) GetByTitleDate(db *system.DB, et string, title string)(err error){

	if et == "" {
		return e.Errors(ErrorMissingValue, "event_type")
	}

	if title == "" {
		return e.Errors(ErrorMissingValue, "title")
	}

	err = db.QueryRow(e.queryGetByTitleDate(), et, title).Scan(
		&e.ID,
		&e.StartDate,
		&e.EndDate,
		&e.Title,
		&e.Description,
		&e.EventType,
		&e.IsActive,
		&e.CompetitorsCount,
		&e.UpvotesCount,
		&e.DownvotesCount,
		&e.CreatedAt,
		&e.UpdatedAt)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Event.GetByTitleDate() id -> %v QueryRow() -> %v Error -> %v", e.ID, e.queryGetByTitleDate(), err)
		return
	}

	if e.ID == 0 {
		log.Println("Event not found. -> ", title)
	}

	return
}

func (e *Event) Exists(db *system.DB, startDate time.Time, et string, title string) (exists bool, err error){

	if startDate.String() == ""{
		return false, e.Errors(ErrorMissingValue, "start_date")
	}

	if et == "" {
		return false, e.Errors(ErrorMissingValue, "event_type")
	}

	if title == "" {
		return false, e.Errors(ErrorMissingValue, "title")
	}

	err = db.QueryRow(e.queryExist(), startDate, et, title).Scan(&exists)

	if err != nil {
		log.Printf("startDate -> %v eventType -> %v title -> %v QueryRow() -> %v Error -> %v", startDate.String(), et, title, e.queryExist(), err)
		return
	}

	return
}

func (e *Event) GetAllEvents(db *system.DB, page int)(events []Event, err error){

	rows, err := db.Query(e.queryGetEvents(), LimitQueryPerRequest, offSet(page))

	defer  rows.Close()

	if err != nil {
		log.Printf("Event.GetAllEvents() Query() -> %v Error -> %v", e.queryGetEvents(), err)
		return
	}

	return e.parseRows(db, rows)
}


func (e *Event) parseRows(db *system.DB, rows *sql.Rows) (events []Event, err error){

	for rows.Next() {
		event := Event{}

		err = rows.Scan(&event.ID,
			&event.StartDate,
			&event.EndDate,
			&event.Title,
			&event.Description,
			&event.EventType,
			&event.IsActive,
			&event.CompetitorsCount,
			&event.UpvotesCount,
			&event.DownvotesCount,
			&event.CreatedAt,
			&event.UpdatedAt)

		if err != nil {
			log.Println("Event.parseRows() Error -> ", e)
			return
		}

		events = append(events, event)
	}

	return
}

// Create a new leaderboard event
func (e *Event)createNextLeaderBoardEvent(db *system.DB) (err error){

	e.StartDate = e.BeginningOfWeekMonday()
	e.EndDate = e.StartDate.Add(time.Hour * time.Duration(168))
	e.EventType = EventType.LeaderBoard
	e.Title = e.StartDate.Format(EventDateLayout)
	e.Description = "Weekly Leader Board"

	return e.Create(db)
}

// Look in DB for any events coming up at Sunday at 12am
// If there is no such event, it will create a new one.
func (e *Event) GetAvailableEvent(db *system.DB) (err error){


	date := e.BeginningOfWeekMonday()

	formattedDate := date.Format(EventDateLayout)

	log.Println("Event Date -> ", formattedDate)


	if err = e.GetByTitleDate(db,  EventType.LeaderBoard, formattedDate); err != nil && err != sql.ErrNoRows {
		log.Println("GetByTitleDate() -> ", err)
		return
	}


	if e.ID == 0 {

		if err = e.createNextLeaderBoardEvent(db); err != nil {
			log.Println("createNextleaderBoardEvent() -> ", err)
			return err
		}

	}

	log.Println("Event id -> ", e.ID)

	return
}


func (e *Event) BeginningOfDay() time.Time {
	hour := time.Time{}

	d := time.Duration(-hour.Hour()) * time.Hour
	return now.BeginningOfHour().Add(d)
}

func (e *Event) BeginningOfWeekMonday() time.Time {


	loc, _ := time.LoadLocation("America/Los_Angeles")

	t := e.BeginningOfDay().In(loc)

	weekday := int(t.Weekday())

		if weekday == 0 {
			weekday = 7
		}
		weekday = weekday - 1


	d := time.Duration(-weekday) * 24 * time.Hour - (23 * time.Hour)
	return t.Add(d)
}




