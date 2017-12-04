package models

import (
	"github.com/rathvong/talentmob_server/system"
	"log"
	"time"
)

// This will handle all possible activities
// a user can participate in to gain points.

type PointActivity int

const (
	POINT_ACTIVITY_VIDEO_WATCHED PointActivity = iota
	POINT_ACTIVITY_VIDEO_VOTED
	POINT_ACTIVITY_FIRST_VOTE
	POINT_ACTIVITY_CORRECT_VOTE
	POINT_ACTIVITY_AD_WATCHED
	POINT_ACTIVITY_REFERRED_USERS
	POINT_ACTIVITY_TWENTY_FOUR_HOUR_BOOST
	POINT_ACTIVITY_THREE_DAYS_BOOST
	POINT_ACTIVITY_SEVEN_DAYS_BOOST
)

// Contains the point value for each activity performed
var activityPoints = []uint64{5, 5, 10, 20, 25, 1000, -1000, -2000, -3000}

// The point value of the activity
func (p *PointActivity) Value() (value uint64){
	return activityPoints[*p]
}




type Point struct {
	BaseModel
	UserID                   uint64 `json:"user_id"`
	VideosWatched            uint64 `json:"videos_watched"`
	VideosVoted              uint64 `json:"videos_voted"`
	FirstVotes               uint64 `json:"first_votes"`
	CorrectVotes             uint64 `json:"correct_votes"`
	AdWatched                uint64 `json:"ad_watched"`
	AdWatchedCount           uint64 `json:"ad_watched_count"`
	LastAdWatchedDate        time.Time `json:"last_ad_watched_date"`
	ReferredUsers            uint64 `json:"referred_users"`
	TwentyFourHourVideoBoost uint64 `json:"twenty_four_hour_video_boost"`
	ThreeDaysVideoBoost      uint64 `json:"three_days_video_boost"`
	SevenDaysVideoBoost      uint64 `json:"seven_days_video_boost"`
	Total                    uint64 `json:"total"`
	IsActive                 bool   `json:"is_active"`

}


func (p * Point) AddPoints(activity PointActivity) {
	switch activity {
	case POINT_ACTIVITY_VIDEO_WATCHED:
		p.VideosWatched = p.VideosWatched + activity.Value()
	case POINT_ACTIVITY_VIDEO_VOTED:
		p.VideosVoted = p.VideosVoted + activity.Value()
	case POINT_ACTIVITY_FIRST_VOTE:
		p.FirstVotes = p.FirstVotes + activity.Value()
	case POINT_ACTIVITY_CORRECT_VOTE:
		p.CorrectVotes = p.CorrectVotes + activity.Value()
	case POINT_ACTIVITY_AD_WATCHED:
		p.AdWatched = p.AdWatched + activity.Value()
	case POINT_ACTIVITY_REFERRED_USERS:
		p.ReferredUsers = p.ReferredUsers + activity.Value()
	case POINT_ACTIVITY_TWENTY_FOUR_HOUR_BOOST:
		p.TwentyFourHourVideoBoost = p.TwentyFourHourVideoBoost + activity.Value()
	case POINT_ACTIVITY_THREE_DAYS_BOOST:
		p.ThreeDaysVideoBoost = p.ThreeDaysVideoBoost + activity.Value()
	case POINT_ACTIVITY_SEVEN_DAYS_BOOST:
		p.SevenDaysVideoBoost = p.SevenDaysVideoBoost + activity.Value()
	}

	p.Total = p.Total + activity.Value()
}

func (p *Point) queryCreate() (qry string){
	return `INSERT INTO points
						(user_id,
						videos_watched,
						videos_voted,
						first_votes,
						correct_votes,
						ad_watched,
						ad_watched_count,
						last_ad_watched_date,
						referred_users,
						twenty_four_hour_video_boost,
						three_days_video_boost,
						seven_days_video_boost,
						total,
						is_active,
						created_at,
						updated_at)
				 VALUES
						($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
				 RETURNING id`
}

func (p *Point) queryUpdate() (qry string){
	return `UPDATE points SET
						videos_watched = $2,
						videos_voted = $3,
						first_votes = $4,
						correct_votes = $5,
						ad_watched = $6,
						ad_watched_count = $7,
						last_ad_watched_date = $8,
						referred_users = $9,
						twenty_four_hour_video_boost = $10,
						three_days_video_boost = $11,
						seven_days_video_boost = $12,
						total = $13,
						is_active = $14,
						updated_at = $15
				WHERE id = $1
					`
}

func (p *Point) queryExistsForUser() (qry string){
	return `SELECT EXISTS(SELECT 1 FROM points WHERE user_id = $1)`
}

func (p *Point) queryGetByUserID() (qry string){
	return `SELECT
						id,
						user_id,
						videos_watched,
						videos_voted,
						first_votes,
						correct_votes,
						ad_watched,
						ad_watched_count,
						last_ad_watched_date,
						referred_users,
						twenty_four_hour_video_boost,
						three_days_video_boost,
						seven_days_video_boost,
						total,
						is_active,
						created_at,
						updated_at

                 FROM	points
				 WHERE id = $1
				 ORDER BY created_at DESC
				 LIMIT 1
`
}

func (p *Point) queryTopUsers() (qry string){
	return `SELECT
					users.id,
				    users.facebook_id,
				    users.avatar,
				    users.name,
				    users.email,
					users.account_type,
					users.minutes_watched,
					users.points,
					users.created_at,
					users.updated_at,
					users.encrypted_password,
					users.favourite_videos_count,
					users.imported_videos_count
				FROM users
				INNER JOIN points
				ON points.user_id = users.id
				WHERE users.id = $1
				ORDER BY points.total DESC
				LIMIT $2,
				OFFSET $3`
}


func (p *Point) validateCreateErrors() (err error){
	if p.UserID == 0 {
		return p.Errors(ErrorMissingValue, "user_id")
	}
	return
}


func (p *Point) Create(db *system.DB) (err error){

	if err = p.validateCreateErrors(); err != nil {
		log.Println("Point.Create() Error -> ", err)
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
		log.Println("Point.Create() Begin() Error -> ", err)
		return
	}

	p.IsActive = true
	p.UpdatedAt = time.Now()
	p.CreatedAt = time.Now()


	err = tx.QueryRow(
		p.queryCreate(),
		p.UserID,
		p.VideosWatched,
		p.VideosVoted,
		p.FirstVotes,
		p.CorrectVotes,
		p.AdWatched,
		p.AdWatchedCount,
		p.LastAdWatchedDate,
		p.ReferredUsers,
		p.TwentyFourHourVideoBoost,
		p.ThreeDaysVideoBoost,
		p.SevenDaysVideoBoost,
		p.Total,
		p.IsActive,
		p.CreatedAt,
		p.UpdatedAt,

		).Scan(&p.ID)

	if err != nil {
		log.Printf("Point.Create() QueryRow() -> %v \n Error -> %v", p.queryCreate(), err)
		return
	}

	return
}

func (p *Point) validateUpdateErrors() (err error){

	if p.ID == 0 {
		return p.Errors(ErrorMissingValue, "id")
	}

	return p.validateCreateErrors()
}

func (p *Point) InitPointSystem(db *system.DB) (err error){
	exists, err := p.ExistsForUser(db, p.UserID)

	if err != nil {
		log.Println("Point.Update() Error -> ", err)
		return
	}

	if !exists {
		if err = p.Create(db); err != nil {
			return
		}
	}

	return

}

func (p *Point) Update(db *system.DB) (err error){


	if err = p.validateUpdateErrors(); err != nil {
		log.Println("Point.Update() Error -> ", err)

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
		log.Println("Point.Update Begin() Error -> ", err)
		return
	}

	p.UpdatedAt = time.Now()

	_, err = tx.Exec(
		p.queryUpdate(),
		p.ID,
		p.VideosWatched,
		p.VideosVoted,
		p.FirstVotes,
		p.CorrectVotes,
		p.AdWatched,
		p.AdWatchedCount,
		p.LastAdWatchedDate,
		p.ReferredUsers,
		p.TwentyFourHourVideoBoost,
		p.ThreeDaysVideoBoost,
		p.SevenDaysVideoBoost,
		p.Total,
		p.IsActive,
		p.UpdatedAt,

	)

	if err != nil {
		log.Printf("Point.Update() id -> %v Exec() -> %v \n Error -> %v", p.ID, p.queryUpdate, err)
		return
	}


	return
}


func (p *Point) ExistsForUser(db *system.DB, userID uint64) (exists bool, err error){

	if userID == 0 {
		err = p.Errors(ErrorMissingValue, "user_id")
		log.Println("Point.ExistsForUser() Error -> ", err)
		return
	}

	err = db.QueryRow(p.queryExistsForUser(), userID).Scan(&exists)

	if err != nil {
		log.Printf("Point.ExistsForUser() userID -> %v QueryRow() -> %v Error -> %v", userID, p.queryExistsForUser(), err)
		return
	}


	return
}


// Will create points table for all users that don't have
// points setup
func (p *Point) AddToUsers(db *system.DB) ( err error){

	users, err := User{}.GetAllUsers(db)

	if err != nil {
		return
	}

	for _, user := range users {

		exists, err := p.ExistsForUser(db, user.ID)

		if err != nil {
			log.Printf("Point.AddToUsers() userID -> %v Error -> %v", user.ID, err)
			return
		}

		if exists {
			continue
		}

		point := Point{}
		point.UserID = user.ID

		if err = point.Create(db); err != nil {
			log.Println("Point.AddToUsers() Point.Create() Error -> ", err)
			return
		}

	}

	return
}

func (p *Point) GetByUserID(db *system.DB, userID uint64) (err error){

	if userID == 0 {
		err = p.Errors(ErrorMissingValue, "user_id")
		log.Println("Point.GetByUserID() Error -> ", err)
		return
	}

	err = db.QueryRow(p.queryGetByUserID(), userID).Scan(
		&p.ID,
		&p.UserID,
		&p.VideosWatched,
		&p.VideosVoted,
		&p.FirstVotes,
		&p.CorrectVotes,
		&p.AdWatched,
		&p.AdWatchedCount,
		&p.LastAdWatchedDate,
		&p.ReferredUsers,
		&p.TwentyFourHourVideoBoost,
		&p.ThreeDaysVideoBoost,
		&p.SevenDaysVideoBoost,
		&p.Total,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt)


	if err != nil {
		log.Printf("Point.GetByUserID() userID -> %v QueryRow() -> %v Error -> %v", userID, p.queryGetByUserID(), err)
		return
	}

	return
}

func (p *Point) GetTopUsers(db *system.DB) (users []User, err error){

	rows, err := db.Query(p.queryTopUsers())

	defer rows.Close()

	if err != nil {
		log.Printf("Point.GetTopUsers() Query() -> %v Error -> %v", p.queryTopUsers(), err)

		return
	}


	return User{}.parseRows(rows)
}