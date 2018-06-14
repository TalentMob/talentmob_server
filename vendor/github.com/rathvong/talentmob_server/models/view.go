package models

import (
	"github.com/rathvong/talentmob_server/system"
	"log"
	"time"
)

// The view struct is to keep track of how many views
// a video has accumulated.
type View struct {
	BaseModel
	UserID  uint64 `json:"user_id"`
	VideoID uint64 `json:"video_id"`
}

// SQL query to create a new row
func (v *View) queryCreate() (qry string) {
	return `INSERT INTO views
				(user_id,
				video_id,
				created_at,
				updated_at)
			VALUES
				($1, $2, $3, $4)
			RETURNING id`
}

// SQL query to check if row exists
func (v *View) queryExists() (qry string) {
	return `SELECT EXISTS(select 1 from views where user_id = $1 and video_id = $2)`
}

// Ensure correct fields are entered
func (v *View) validateError() (err error) {
	if v.UserID == 0 {
		return v.Errors(ErrorMissingValue, "userID")
	}

	if v.VideoID == 0 {
		return v.Errors(ErrorMissingValue, "videoID")
	}

	return
}

// Create a new view
func (v *View) Create(db *system.DB) (err error) {

	if err = v.validateError(); err != nil {
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
		log.Println("View.Create() Begin()", err)
		return
	}

	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()

	err = tx.QueryRow(v.queryCreate(),
		v.UserID,
		v.VideoID,
		v.CreatedAt,
		v.UpdatedAt).Scan(&v.ID)

	if err != nil {
		log.Printf("View.Create() QueryRow() -> %v Error -> %v", v.queryCreate(), err)
		return
	}

	log.Println("View.Create() View created, id -> ", v.ID)
	return
}

// Check if a view exists
func (v *View) Exists(db *system.DB, userID uint64, videoID uint64) (exists bool, err error) {
	if userID == 0 {
		return false, v.Errors(ErrorMissingValue, "userID")
	}

	if videoID == 0 {
		return false, v.Errors(ErrorMissingValue, "videoID")
	}

	err = db.QueryRow(v.queryExists(), userID, videoID).Scan(&exists)

	if err != nil {
		log.Printf("View.Exists() userID -> %v videoID -> %v QueryRow() -> %v Error -> %v", userID, videoID, v.queryExists(), err)
		return
	}

	log.Println("View.Exists() Exists -> ", exists)
	return
}
