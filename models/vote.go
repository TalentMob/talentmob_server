package models

import (
	"log"
	"time"

	"github.com/rathvong/talentmob_server/system"
)

// Main structure for vote models
// votes can be limited to 1 per user per video
type Vote struct {
	BaseModel
	Upvote   int    `json:"upvote"`
	Downvote int    `json:"downvote"`
	UserID   uint64 `json:"user_id"`
	VideoID  uint64 `json:"video_id"`
}

//SQL query to create a row
func (v *Vote) queryCreate() (qry string) {
	return `INSERT INTO votes
				(upvote,
				downvote,
				user_id,
				video_id,
				created_at,
				updated_at)
			VALUES
				($1, $2, $3, $4, $5, $6)
			RETURNING id`
}

//SQL query if vote exists
func (v *Vote) queryExists() (qry string) {
	return `SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = $2)`
}

//SQL query to retrieve vote by user_id and video_id
func (v *Vote) queryGet() (qry string) {
	return `SELECT 	id,
					upvote,
					downvote,
					user_id,
					video_id,
					created_at,
					updated_at
			FROM 	votes
			WHERE	user_id = $1
			AND 	video_id = $2
			ORDER BY created_at DESC
			LIMIT 1`
}

//SQL query check if user has upvoted with in weekly interval
func (v *Vote) queryHasUpvoted() (qry string) {
	return `SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = $2 and upvote > 0)`
}

//SQL query check if user has downvoted with in weekly interval
func (v *Vote) queryHasDownvoted() (qry string) {
	return `SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = $2 and downvote > 0)`
}

func (v *Vote) queryVoteCount() (qry string) {
	return `SELECT COUNT(*) FROM votes WHERE video_id = $1`
}

func (v *Vote) Count(db *system.DB, videoID uint64) (count uint64, err error) {

	if videoID == 0 {

		err = v.Errors(ErrorMissingValue, "video_id")
		return
	}

	err = db.QueryRow(v.queryVoteCount(), videoID).Scan(&count)

	if err != nil {
		log.Printf("Vote.count() videoID -> %v QueryRow() -> %v Error -> %v", videoID, v.queryVoteCount(), err)
		return
	}

	return
}

// ensure correct fields are entered
func (v *Vote) validateErrors() (err error) {
	if v.UserID == 0 {
		return v.Errors(ErrorMissingValue, "userID")
	}

	if v.VideoID == 0 {
		return v.Errors(ErrorMissingValue, "videoID")
	}

	return
}

func (v *Vote) UpdatePoints(db *system.DB) (err error) {

	var count uint64
	var video Video

	video.GetVideoByID(db, v.VideoID)

	p := Point{}

	if err = p.GetByUserID(db, v.UserID); err != nil {
		panic(err)
	}

	count, err = v.Count(db, v.VideoID)

	if err != nil {
		return err
	}

	if (count - 1) == 0 {
		p.AddPoints(POINT_ACTIVITY_FIRST_VOTE)
	}

	return p.Update(db)
}

// create a new vote
func (v *Vote) Create(db *system.DB) (err error) {
	if err = v.validateErrors(); err != nil {
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

		v.UpdatePoints(db)

	}()

	if err != nil {
		log.Println("Vote.Create() Begin() ", err)

		return
	}

	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()

	err = db.QueryRow(v.queryCreate(),
		v.Upvote,
		v.Downvote,
		v.UserID,
		v.VideoID,
		v.CreatedAt,
		v.UpdatedAt).Scan(&v.ID)

	if err != nil {
		log.Printf("Vote.Create() QueryRow() -> %v Error -> %v", v.queryCreate(), err)
		return
	}

	log.Println("Vote.Create() Vote created, id -> ", v.ID)
	return
}

// retrieve a vote
func (v *Vote) Get(db *system.DB, userID uint64, videoID uint64) (err error) {

	if userID == 0 {
		return v.Errors(ErrorMissingValue, "userID")
	}

	if videoID == 0 {
		return v.Errors(ErrorMissingValue, "videoID")
	}

	err = db.QueryRow(v.queryGet(), userID, videoID).Scan(&v.ID,
		&v.Upvote,
		&v.Downvote,
		&v.UserID,
		&v.VideoID,
		&v.CreatedAt,
		&v.UpdatedAt)

	if err != nil {
		log.Printf("Vote.Get() userID -> %v videoID -> %v QueryRow -> %v Error -> %v", userID, videoID, v.queryGet(), err)
	}

	return
}

// validate if a vote exists
func (v *Vote) Exists(db *system.DB, userID uint64, videoID uint64) (exists bool, err error) {
	if userID == 0 {
		return false, v.Errors(ErrorMissingValue, "userID")
	}

	if videoID == 0 {
		return false, v.Errors(ErrorMissingValue, "videoID")
	}

	err = db.QueryRow(v.queryExists(), userID, videoID).Scan(&exists)

	if err != nil {
		log.Printf("Vote.Exists() userID -> %v videoID -> %v QueryRow() -> %v Error -> %v", userID, videoID, v.queryExists(), err)
		return
	}

	return
}

// check for last upvote
func (v *Vote) RecentUpvote(db *system.DB, userID uint64, videoID uint64) (voted bool, err error) {
	if exists, err := v.Exists(db, userID, videoID); !exists || err != nil {
		return false, err
	}

	v.Get(db, userID, videoID)

	if v.Upvote > 0 {
		return true, nil
	}

	return false, nil
}

// check for last downvote
func (v *Vote) RecentDownvote(db *system.DB, userID uint64, videoID uint64) (voted bool, err error) {
	if exists, err := v.Exists(db, userID, videoID); !exists || err != nil {
		return false, err
	}

	v.Get(db, userID, videoID)

	if v.Downvote > 0 {
		return true, nil
	}

	return false, nil
}

// validate if a user has upvoted
func (v *Vote) HasUpVoted(db *system.DB, userID uint64, videoID uint64) (voted bool, err error) {
	if userID == 0 {
		return false, v.Errors(ErrorMissingValue, "userID")
	}

	if videoID == 0 {
		return false, v.Errors(ErrorMissingValue, "videoID")
	}

	err = db.QueryRow(v.queryHasUpvoted(), userID, videoID).Scan(&voted)

	if err != nil {
		log.Printf("Vote.HasUpVoted() userID -> %v videoID -> %v QueryRow() -> %v Error -> %v", userID, videoID, v.queryHasUpvoted(), err)
		return
	}

	return
}

// validate if a user has downvoted
func (v *Vote) HasDownVoted(db *system.DB, userID uint64, videoID uint64) (voted bool, err error) {
	if userID == 0 {
		return false, v.Errors(ErrorMissingValue, "userID")
	}

	if videoID == 0 {
		return false, v.Errors(ErrorMissingValue, "videoID")
	}

	err = db.QueryRow(v.queryHasDownvoted(), userID, videoID).Scan(&voted)

	if err != nil {
		log.Printf("Vote.HasDownVoted() userID -> %v videoID -> %v QueryRow() -> %v Error -> %v", userID, videoID, v.queryHasDownvoted(), err)
		return
	}

	return
}
