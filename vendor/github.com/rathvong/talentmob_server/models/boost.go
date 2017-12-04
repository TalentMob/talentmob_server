package models

import (
	"time"
	"github.com/rathvong/talentmob_server/system"
	"log"
	"database/sql"
)


type Boost struct {
	BaseModel
	UserID    uint64    `json:"user_id"`
	VideoID   uint64    `json:"video_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsActive  bool      `json:"is_active"`
}


func (b *Boost) queryCreate() (qry string){
	return `INSERT INTO boosts
						(user_id,
						video_id,
						start_time,
						end_time,
						is_active,
						created_at,
						updated_at)
				VALUES
						($1, $2, $3, $4, $5, $6)
				RETURNING id`
}

func (b *Boost) queryUpdate() (qry string){
	return `UPDATE boosts SET
						user_id = $2,
						video_id = $3,
						start_time = $4,
						end_time = $5,
						is_active = $6,
						updated_at = $7
				WHERE id = $1`
}

func (b *Boost) queryGetByVideoID() (qry string){
	return `SELECT
						id,
						user_id,
						video_id,
						start_time,
						end_time,
						is_active,
						created_at,
						updated_at
				FROM boosts
				WHERE video_id = $1
				AND is_active = true
				ORDER BY created_at DESC
				LIMIT 1`
}

func (b *Boost) queryGetByUserID() (qry string){
	return `SELECT
						id,
						user_id,
						video_id,
						start_time,
						end_time,
						is_active,
						created_at,
						updated_at
				FROM boosts
				WHERE user_id = $1
				AND is_active = true
				ORDER BY created_at DESC
				LIMIT $2
				OFFSET $3
				`
}

func (b *Boost) queryExistsForVideo() (qry string){
	return `SELECT EXISTS(SELECT 1 FROM boosts WHERE video_id = $1 AND end_time > $2 AND is_active = true)`
}

func (b *Boost) validateCreateErrors() (err error){
	if b.UserID == 0 {
		return b.Errors(ErrorMissingValue, "user_id")
	}

	if b.VideoID == 0 {
		return b.Errors(ErrorMissingValue, "video_id")
	}

	if b.StartTime.String() == "" {
		return b.Errors(ErrorMissingValue, "start_time")
	}

	if b.EndTime.String() == "" {
		return b.Errors(ErrorMissingValue, "end_time")
	}

	return
}

func (b *Boost) validateUpdateErrors() (err error){

	if b.ID == 0 {
		return b.Errors(ErrorMissingValue, "id")
	}

	return b.validateCreateErrors()
}

func (b *Boost) Create(db *system.DB) (err error){
	if err = b.validateCreateErrors(); err != nil {
		log.Println("Boost.Create() Error -> ", err)
		return
	}

	if exists, err := b.ExistsForVideo(db, b.VideoID); exists || err != nil {
		if err != nil {
			return
		}

		err = b.Errors(ErrorExists, "video_id" )
		log.Println("Boost.Create() A current boost is already active -> ", err)

		return
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
		log.Println("Boost.Create() Begin() Error -> ", err)
		return
	}

	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()

	err = tx.QueryRow(
			b.queryCreate(),
			b.UserID,
			b.VideoID,
			b.StartTime,
			b.EndTime,
			b.IsActive,
			b.CreatedAt,
			b.UpdatedAt,

	).Scan(&b.ID)

	if err != nil {
		log.Printf("Boost.Create() QueryRow() -> %v Error -> %v", b.queryCreate(), err)
		return
	}

	return
}

func (b *Boost) Update(db *system.DB) (err error){
	if err = b.validateUpdateErrors(); err != nil {
		log.Println("Boost.Update() Error -> ", err)
		return
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
		log.Println("Boost.Update() Begin() Error -> ", err)
		return
	}

	b.UpdatedAt = time.Now()

	_, err = tx.Exec(
		b.queryUpdate(),
		b.ID,
		b.UserID,
		b.VideoID,
		b.StartTime,
		b.EndTime,
		b.IsActive,
		b.UpdatedAt,

	)

	if err != nil {
		log.Printf("Boost.Update() Exec() -> %v Error -> %v", b.queryUpdate(), err)
		return
	}

	return
}

func (b *Boost) ExistsForVideo(db *system.DB, videoID uint64) (exists bool, err error){

	if videoID == 0 {
		err = b.Errors(ErrorMissingValue, "video_id")
		log.Println("Boost.GetByVideoID() Error -> ", err)
		return
	}

	err = db.QueryRow(b.queryGetByVideoID(), videoID).Scan(&exists)

	if err != nil {
		log.Printf("Boost.ExistsForVideo() videoID -> %v QueryRow() -> %v Error -> %v", videoID, b.queryExistsForVideo(), err)
		return
	}

	return
}

func (b *Boost) GetByVideoID(db *system.DB, videoID uint64) (err error){

	if videoID == 0 {
		err = b.Errors(ErrorMissingValue, "video_id")
		log.Println("Boost.GetByVideoID() Error -> ", err)
		return
	}

	err = db.QueryRow(b.queryGetByVideoID(), videoID).Scan(
		&b.ID,
		&b.UserID,
		&b.VideoID,
		&b.StartTime,
		&b.EndTime,
		&b.IsActive,
		&b.CreatedAt,
		&b.UpdatedAt,)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Boost.GetByVideoID() videoID -> %v QueryRow() -> %v Error -> %v", videoID, b.queryGetByVideoID(), err)
		return
	}

	return
}

func (b *Boost) GetByUserID(db *system.DB, userID uint64, page int) (boosts []Boost, err error){
	if userID == 0 {
		err = b.Errors(ErrorMissingValue, "user_id")
		log.Println("Boost.GetByUserID() Error -> ", err)
		return
	}

	rows, err := db.Query(b.queryGetByUserID(), userID, LimitQueryPerRequest, offSet(page))

	defer rows.Close()

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Boost.GetByUserID() videoID -> %v QueryRow() -> %v Error -> %v",userID, b.queryGetByUserID(), err)

		return
	}

	return b.parseRows(rows)
}

func (b *Boost) parseRows(rows *sql.Rows) (boosts []Boost, err error){

	for rows.Next(){
		boost := Boost{}
		err = rows.Scan(
		    &boost.ID,
			&boost.UserID,
			&boost.VideoID,
			&boost.StartTime,
			&boost.EndTime,
			&boost.IsActive,
			&boost.CreatedAt,
			&boost.UpdatedAt,
			)

		if err != nil {
			log.Println("Boost.parseRows() Error -> ", err)
			return
		}

		boosts = append(boosts, boost)
	}

	return
}


