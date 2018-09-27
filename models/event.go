package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/rathvong/talentmob_server/leaderboardpayouts"
	"github.com/rathvong/talentmob_server/system"
)

const (
	EventDateLayout   = "01/2/2006"
	EventCreateLayout = "2006-01-02"
)

type Event struct {
	BaseModel
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	EventType        string    `json:"event_type"`
	IsActive         bool      `json:"is_active"`
	CompetitorsCount uint64    `json:"competitors_count"`
	UpvotesCount     uint64    `json:"upvotes_count"`
	DownvotesCount   uint64    `json:"downvotes_count"`
	EndDateUnix      int64     `json:"end_date_unix"`
	PrizePool        uint64    `json:"prize_pool"`
	PrizeList        []uint    `json:"prize_list"`
	ThumbNail        string    `json:"thumb_nail"`
	BuyIn            uint64    `json:"buy_in"`
	IsOpened         bool      `json:"is_open"`
	BuyInFee         uint      `json:"buy_in_fee"`
	UserID           uint64    `json:"user_id"`
}

var EventType = eventType{
	LeaderBoard:   "leaderboard",
	Weekly:        "weekly",
	Daily:         "daily",
	Hourly:        "hourly",
	UserGenerated: "user_generated",
}

type eventType struct {
	LeaderBoard   string `json:"leaderboard"`
	Weekly        string `json:"weekly"`
	Daily         string `json:"daily"`
	Hourly        string `json:"hourly"`
	UserGenerated string `json:"user_generated"`
}

func (e *Event) queryCreate() (qry string) {
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
				updated_at,
				prize_pool,
				thumb_nail,
				buy_in,
				is_open,
				buy_in_fee,
				user_id)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)

			RETURNING id
	`
}

func (e *Event) queryUpdate() (qry string) {
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
				updated_at = $11,
				prize_pool = $12,
				thumb_nail = $13,
				buy_in = $14,
				is_open = $15,
				buy_in_fee = $16,
				user_id = $17
				WHERE id = $1`
}

func (e *Event) querySoftDelete() (qry string) {
	return `UPDATE events set
				is_active = $2
			WHERE id = $1`
}

func (e *Event) queryGetByID() (qry string) {
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
				updated_at,
				prize_pool,
				thumb_nail,
				buy_in,
				is_open,
				buy_in_fee,
				user_id
			FROM events
			WHERE
				id = $1
			`
}

func (e *Event) queryGetByTitleDate() (qry string) {
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
				updated_at,
				prize_pool,
				thumb_nail,
				buy_in,
				is_open,
				buy_in_fee,
				user_id
			FROM events
			WHERE
				event_type = $1
			AND
				title = $2
			`
}

func (e *Event) queryGetEvents() (qry string) {
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
				updated_at,
				prize_pool,
				thumb_nail,
				buy_in,
				is_open,
				buy_in_fee,
				user_id
			FROM events
			WHERE is_active = true
			AND event_type = 'leaderboard'
			ORDER BY start_date DESC
			LIMIT $1
			OFFSET $2 `
}

func (e *Event) queryExist() (qry string) {
	return `SELECT EXISTS( select 1 from events where start_date = $1 and event_type = $2 and title = $3)`
}

func (e *Event) validateCreateErrors() (err error) {

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

func (e *Event) validateUpdateErrors() (err error) {
	if e.ID == 0 {
		return e.Errors(ErrorMissingValue, "id")
	}

	return e.validateCreateErrors()
}

func (e *Event) Create(db *system.DB) (err error) {

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

	e.Title = strings.Replace(e.Title, " ", "", -1)

	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	e.IsActive = true
	e.IsOpened = true

	//startDate = "('"+ e.BeginningOfWeekMonday().Format(EventCreateLayout) +"' AT TIME ZONE 'UTC') AT TIME ZONE 'America/Los_Angeles'"

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
		e.UpdatedAt,
		e.PrizePool,
		e.ThumbNail,
		e.BuyIn,
		e.IsOpened,
		e.BuyInFee,
		e.UserID,
	).Scan(&e.ID)

	if err != nil {
		log.Printf("startDate -> %v title -> %v eventType -> %v QueryRow() -> %v Error -> %v", e.StartDate.String(), e.Title, e.EventType, e.queryCreate(), err)
		return
	}

	log.Println("Event.create() Event Created -> ", e.ID)

	return
}

func (e *Event) Update(db *system.DB) (err error) {

	if err := e.validateUpdateErrors(); err != nil {
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
		e.UpdatedAt,
		e.PrizePool,
		e.ThumbNail,
		e.BuyIn,
		e.IsOpened,
		e.BuyInFee,
		e.UserID,
	)

	if err != nil {
		log.Printf("Event.Update() id -> %v Exec() -> %v Error -> %v", e.ID, e.queryUpdate(), err)
		return
	}

	return
}

func (e *Event) SoftDelete(db *system.DB) (err error) {
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

func (e *Event) Get(db *system.DB, eventID uint64) (err error) {
	if eventID == 0 {
		return e.Errors(ErrorMissingValue, "event_id")
	}

	var userID sql.NullInt64

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
		&e.UpdatedAt,
		&e.PrizePool,
		&e.ThumbNail,
		&e.BuyIn,
		&e.IsOpened,
		&e.BuyInFee,
		&userID)

	if err != nil {
		log.Printf("Event.Get() id -> %v QueryRow() -> %v Error -> %v", e.ID, e.queryGetByID(), err)
		return
	}

	if userID.Valid {
		e.UserID = uint64(userID.Int64)
	}

	e.EndDateUnix = e.EndDate.UnixNano() / 1000000

	return
}

/**
 	We used the event title as the identifier KEY to retrieve a specific event.
	All events are organized by upcoming Monday.
*/
func (e *Event) GetByTitleDate(db *system.DB, et string, title string) (err error) {

	if et == "" {
		return e.Errors(ErrorMissingValue, "event_type")
	}

	if title == "" {
		return e.Errors(ErrorMissingValue, "title")
	}

	var userID sql.NullInt64

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
		&e.UpdatedAt,
		&e.PrizePool,
		&e.ThumbNail,
		&e.BuyIn,
		&e.IsOpened,
		&e.BuyInFee,
		&userID,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Event.GetByTitleDate() id -> %v QueryRow() -> %v Error -> %v", e.ID, e.queryGetByTitleDate(), err)
		return
	}

	if e.ID == 0 {
		log.Println("Event not found. -> ", title)
	}

	if userID.Valid {
		e.UserID = uint64(userID.Int64)
	}

	return
}

func (e *Event) Exists(db *system.DB, startDate time.Time, et string, title string) (exists bool, err error) {

	if startDate.String() == "" {
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

func (e *Event) GetAllEvents(db *system.DB, limit int, offset int) (events []Event, err error) {

	rows, err := db.Query(e.queryGetEvents(), limit, offset)

	defer rows.Close()

	if err != nil {
		log.Printf("Event.GetAllEvents() Query() -> %v Error -> %v", e.queryGetEvents(), err)
		return
	}

	return e.parseRows(db, rows)
}

func (e *Event) parseRows(db *system.DB, rows *sql.Rows) (events []Event, err error) {

	for rows.Next() {
		event := Event{}

		var userID sql.NullInt64

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
			&event.UpdatedAt,
			&event.PrizePool,
			&event.ThumbNail,
			&event.BuyIn,
			&event.IsOpened,
			&event.BuyInFee,
			&userID,
		)

		if err != nil {
			log.Println("Event.parseRows() Error -> ", e)
			return
		}

		event.EndDateUnix = event.StartDate.Add(time.Hour*time.Duration(168)).UnixNano() / 1000000

		if event.PrizePool > 0 {
			rank, _ := leaderboardpayouts.BuildRankingPayout()
			event.PrizeList = rank.GetValuesForEntireRanking(rank.DisplayForRanking(event.PrizePool, int(event.CompetitorsCount)))
		}

		if userID.Valid {
			event.UserID = uint64(userID.Int64)
		}

		events = append(events, event)
	}

	return
}

func (e *Event) GetAllEvents2(db *system.DB, limit int, offset int) (events []Event, err error) {

	qry := `SELECT
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
				updated_at,
				prize_pool,
				thumb_nail,
				buy_in,
				is_open,
				buy_in_fee,
				user_id
			FROM events
			WHERE is_active = true
			AND event_type = 'leaderboard'
			ORDER BY start_date DESC
			LIMIT $1
			OFFSET $2 `

	rows, err := db.Query(qry, limit, offset)

	defer rows.Close()

	if err != nil {
		log.Printf("Event.GetAllEvents2() Query() -> %v Error -> %v", qry, err)
		return
	}

	return e.parseRows2(db, rows)
}

func (e *Event) GetAllEventsByRunning(db *system.DB, isOpened bool) ([]Event, error) {

	qry := `SELECT
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
					updated_at,
					prize_pool,
					thumb_nail,
					buy_in,
					is_open,
					buy_in_fee,
					user_id
			FROM events
			WHERE is_open = $1
			AND is_active = true
			ORDER BY start_date DESC
			 `

	rows, err := db.Query(qry, isOpened)

	defer rows.Close()

	if err != nil {
		log.Printf("Event.GetAllOpenedEvents() Query() -> %v Error -> %v", qry, err)
		return nil, err
	}

	return e.parseRows2(db, rows)
}

func (e *Event) parseRows2(db *system.DB, rows *sql.Rows) (events []Event, err error) {

	for rows.Next() {
		event := Event{}

		var userID sql.NullInt64

		err = rows.Scan(
			&event.ID,
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
			&event.UpdatedAt,
			&event.PrizePool,
			&event.ThumbNail,
			&event.BuyIn,
			&event.IsOpened,
			&event.BuyInFee,
			&userID,
		)

		if err != nil {
			log.Println("Event.parseRows2() Error -> ", e)
			return
		}

		if userID.Valid {
			event.UserID = uint64(userID.Int64)
		}

		event.EndDateUnix = event.StartDate.Add(time.Hour*time.Duration(168)).UnixNano() / 1000000

		events = append(events, event)
	}

	return
}

// Create a new leaderboard event
func (e *Event) createNextLeaderBoardEvent(db *system.DB) (err error) {
	loc, _ := time.LoadLocation("America/Los_Angeles")

	e.StartDate = e.BeginningOfWeekMonday().In(loc)
	e.EndDate = e.StartDate.Add(time.Hour * time.Duration(168))
	e.EventType = EventType.LeaderBoard
	e.Title = e.StartDate.Format(EventDateLayout)
	e.Description = "Weekly Leader Board"

	return e.Create(db)
}

// Look in DB for any events coming up at Sunday at 12am
// If there is no such event, it will create a new one.
func (e *Event) GetAvailableWeeklyEvent(db *system.DB) (err error) {

	date := e.BeginningOfWeekMonday()

	formattedDate := date.Format(EventDateLayout)

	log.Println("Event Date -> ", formattedDate)

	if err = e.GetByTitleDate(db, EventType.LeaderBoard, formattedDate); err != nil && err != sql.ErrNoRows {
		log.Println("GetByTitleDate() -> ", err)
		return
	}

	if e.ID == 0 {

		if err = e.createNextLeaderBoardEvent(db); err != nil {
			log.Println("createNextleaderBoardEvent() -> ", err)
			return err
		}

		if err = e.updateEventDateToProperTime(db); err != nil {
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

func (e *Event) updateEventDateToProperTime(db *system.DB) error {

	if e.ID == 0 {
		return e.Errors(ErrorMissingID, "Event.updateEventDateToPropertime() id")
	}

	qry := fmt.Sprintf("UPDATE events SET start_date = ('%s' AT TIME ZONE 'UTC') AT TIME ZONE 'America/Los_Angeles' where id = %d;", e.StartDate.Format(EventCreateLayout), e.ID)

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
		return err
	}

	_, err = tx.Exec(qry)

	if err != nil {
		log.Printf("Query: %s ", qry)
		return err
	}

	return nil
}

func (e *Event) BeginningOfWeekMonday() time.Time {

	loc, _ := time.LoadLocation("America/Los_Angeles")

	t := e.BeginningOfDay().In(loc)

	weekday := int(t.Weekday())

	if weekday == 0 {
		weekday = 7
	}
	weekday = weekday - 1

	d := time.Duration(-weekday) * 24 * time.Hour
	return t.Add(d)
}

func (e *Event) LastClosedEvent() {

}

type EventRanking struct {
	BaseModel
	EventID        uint64 `json:"event_id"`
	CompetitorID   uint64 `json:"competitor_id"`
	UserID         uint64 `json:"user_id"`
	Ranking        uint   `jons:"ranking"`
	PayOut         uint   `json:"pay_out"`
	TotalVotes     uint   `json:"total_upvotes"`
	VideoID        uint64 `json:"video_id"`
	VideoTitle     string `json:"video_title"`
	VideoThumbnail string `json:"video_thumbnail"`
	IsPaid         bool   `json:"is_paid"`
	IsActive       bool   `json:"is_active"`
	EventTitle     string `json:"event_title"`
}

func (e *EventRanking) validateCreate() error {
	if e.CompetitorID == 0 {
		return e.Errors(ErrorMissingValue, "EventRanking: competitor_id")
	}

	if e.UserID == 0 {
		return e.Errors(ErrorMissingValue, "EventRanking: user_id")
	}

	if e.EventID == 0 {
		return e.Errors(ErrorMissingValue, "EventRanking: event_id")
	}

	if e.VideoID == 0 {
		return e.Errors(ErrorMissingValue, "EventRanking: video_id")
	}

	if e.EventTitle == "" {
		return e.Errors(ErrorMissingValue, "EventRanking: event_title")
	}

	return nil
}

func (e *EventRanking) Create(db *system.DB) error {

	if e.validateCreate() != nil {
		return e.validateCreate()
	}

	sql := `INSERT INTO event_rankings (
				event_id,
				competitor_id,
				user_id,
				ranking,
				pay_out,
				total_upvotes,
				video_title,
				video_thumbnail,
				is_active,
				created_at,
				updated_at,
				is_paid, 
				video_id,
				event_title
			) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
			) RETURNING id`

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}()

	if err != nil {
		return err
	}

	e.IsActive = true
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()

	err = tx.QueryRow(sql,
		e.EventID,
		e.CompetitorID,
		e.UserID,
		e.Ranking,
		e.PayOut,
		e.TotalVotes,
		e.VideoTitle,
		e.VideoThumbnail,
		e.IsActive,
		e.CreatedAt,
		e.UpdatedAt,
		e.IsPaid,
		e.VideoID,
		e.EventTitle,
	).Scan(&e.ID)

	if err != nil {
		log.Printf("EventRanking.Create() Sql -> %v, Error: %v", sql, err)
		return err
	}

	return nil
}

func (e *EventRanking) validateUpdate() error {
	if e.ID == 0 {
		return e.Errors(ErrorMissingID, "EventRanking: id")
	}

	return e.validateCreate()
}

func (e *EventRanking) Update(db *system.DB) error {

	if e.validateUpdate() != nil {
		return e.validateUpdate()
	}

	sql := `UPDATE event_rankings SET
				event_id = $2,
				competitor_id $3,
				user_id = $4,
				ranking = $5,
				pay_out = $6,
				total_upvotes = $7,
				video_title = $8,
				video_thumbnail = $9,
				is_active = $10,
				created_at = $11,
				updated_at = $12,
				is_paid = $13,
				video_id = $14,
				event_title = $15
			WHERE id = $1
	`

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}()

	if err != nil {
		return err
	}

	_, err = tx.Exec(sql,
		e.ID,
		e.EventID,
		e.CompetitorID,
		e.UserID,
		e.Ranking,
		e.PayOut,
		e.TotalVotes,
		e.VideoTitle,
		e.VideoThumbnail,
		e.IsActive,
		e.CreatedAt,
		e.UpdatedAt,
		e.IsPaid,
		e.VideoID,
		e.EventTitle,
	)

	if err != nil {
		log.Printf("EventRanking.Update() Sql -> %v, Error: %v", sql, err)
		return err
	}

	return nil
}

func (e *EventRanking) Get(db *system.DB, competitorID uint64) error {

	if competitorID == 0 {
		return e.Errors(ErrorMissingValue, "EventRanking.Get: competitorID")
	}

	sql := `SELECT	
				id,
				event_id,
				competitor_id,
				user_id,
				ranking,
				pay_out,
				total_upvotes,
				video_title,
				video_thumbnail,
				is_active,
				created_at,
				updated_at,
				is_paid, 
				video_id,
				event_title
			FROM event_rankings
			WHERE competitor = $1	
			`

	err := db.QueryRow(sql, competitorID).Scan(
		&e.ID,
		&e.EventID,
		&e.CompetitorID,
		&e.UserID,
		&e.Ranking,
		&e.PayOut,
		&e.TotalVotes,
		&e.VideoTitle,
		&e.VideoThumbnail,
		&e.IsActive,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.IsPaid,
		&e.VideoID,
		&e.EventTitle,
	)

	if err != nil {
		log.Printf("EventRanking.Get() id:%v, sql: %v, error: %v", competitorID, sql, err)
		return err
	}

	return nil
}
