package models

import (
	"log"
	"time"

	"github.com/rathvong/talentmob_server/system"
)

// This will handle all possible activities
// a user can participate in to gain points.

type PointActivity int

const (
	POINT_ACTIVITY_FIRST_VOTE PointActivity = iota
	POINT_ACTIVITY_CORRECT_VOTE
	POINT_ACTIVITY_AD_WATCHED
	POINT_ACTIVITY_REFERRED_USERS
	POINT_ACTIVITY_TWENTY_FOUR_HOUR_BOOST
	POINT_ACTIVITY_THREE_DAYS_BOOST
	POINT_ACTIVITY_SEVEN_DAYS_BOOST
	POINT_ACTIVITY_INCORRECT_VOTE
	POINT_ACTIVITY_TIE_VOTE
	POINT_ACTIVITY_8_HOUR_BOOST
	POINT_TRANSACTION_2250_STARPOWER
	POINT_TRANSACTION_9500_STARPOWER
	POINT_TRANSACTION_24500_STARPOWER
	POINT_TRANSACTION_100000_STARPOWER
)

// Contains the point value for each activity performed
var activityPoints = []int64{10, 25, 50, 1000, -2500, -5000, -10000, 0, 10, -1000, 2250, 9500, 24500, 100000}

const (
	POINT_ADS         = "ads"
	POINT_VOTE        = "vote"
	POINT_BOOST       = "boost"
	POINT_VIEW        = "view"
	POINT_TRANSACTION = "transaction"
)

// The point value of the activity
func (p *PointActivity) Value() (value int64) {
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
	ReferredUsers            uint64 `json:"referred_users"`
	TwentyFourHourVideoBoost int64  `json:"twenty_four_hour_video_boost"`
	ThreeDaysVideoBoost      int64  `json:"three_days_video_boost"`
	SevenDaysVideoBoost      int64  `json:"seven_days_video_boost"`
	Total                    int64  `json:"total"`
	TotalLifetime            int64  `json:"total_lifetime"`
	TotalMob                 int64  `json:"total_mob"`
	IsActive                 bool   `json:"is_active"`
}

func (p *Point) isAbleToAddPointsForAds() (permission bool, err error) {
	return
}

func (p *Point) AddPoints(activity PointActivity) {
	switch activity {

	case POINT_ACTIVITY_FIRST_VOTE:
		p.FirstVotes = p.FirstVotes + uint64(activity.Value())
		p.TotalMob = p.TotalMob + activity.Value()
		p.TotalLifetime = p.TotalLifetime + activity.Value()

	case POINT_ACTIVITY_CORRECT_VOTE:
		p.CorrectVotes = p.CorrectVotes + uint64(activity.Value())
		p.TotalMob = p.TotalMob + activity.Value()
		p.TotalLifetime = p.TotalLifetime + activity.Value()

	case POINT_ACTIVITY_INCORRECT_VOTE:
		p.CorrectVotes = p.CorrectVotes + uint64(activity.Value())
		p.TotalMob = p.TotalMob + activity.Value()
		p.TotalLifetime = p.TotalLifetime + activity.Value()

	case POINT_ACTIVITY_TIE_VOTE:
		p.CorrectVotes = p.CorrectVotes + uint64(activity.Value())
		p.TotalMob = p.TotalMob + activity.Value()
		p.TotalLifetime = p.TotalLifetime + activity.Value()

	case POINT_ACTIVITY_AD_WATCHED:
		p.AdWatched = p.AdWatched + uint64(activity.Value())
		p.TotalLifetime = p.TotalLifetime + activity.Value()

	case POINT_ACTIVITY_REFERRED_USERS:
		p.ReferredUsers = p.ReferredUsers + uint64(activity.Value())
		p.TotalLifetime = p.TotalLifetime + activity.Value()

	case POINT_ACTIVITY_TWENTY_FOUR_HOUR_BOOST:
		p.TwentyFourHourVideoBoost = p.TwentyFourHourVideoBoost + activity.Value()

	case POINT_ACTIVITY_THREE_DAYS_BOOST:
		p.ThreeDaysVideoBoost = p.ThreeDaysVideoBoost + activity.Value()

	case POINT_ACTIVITY_SEVEN_DAYS_BOOST:
		p.SevenDaysVideoBoost = p.SevenDaysVideoBoost + activity.Value()

	case POINT_TRANSACTION_2250_STARPOWER, POINT_TRANSACTION_9500_STARPOWER, POINT_TRANSACTION_24500_STARPOWER, POINT_TRANSACTION_100000_STARPOWER:
		p.TotalLifetime = p.TotalLifetime + activity.Value()
	}

	p.Total = p.Total + activity.Value()

	return
}

func (p *Point) AddPayout(payout int64) {
	p.Total = p.Total + payout
	p.TotalLifetime = p.TotalLifetime + payout
	return
}

func (p *Point) queryCreate() (qry string) {
	return `INSERT INTO points
						(user_id,
						videos_watched,
						videos_voted,
						first_votes,
						correct_votes,
						ad_watched,
						referred_users,
						twenty_four_hour_video_boost,
						three_days_video_boost,
						seven_days_video_boost,
						total,
						total_lifetime,
						total_mob,
						is_active,
						created_at,
						updated_at)
				 VALUES
						($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16 )
				 RETURNING id`
}

func (p *Point) queryUpdate() (qry string) {
	return `UPDATE points SET
						videos_watched = $2,
						videos_voted = $3,
						first_votes = $4,
						correct_votes = $5,
						ad_watched = $6,
						referred_users = $7,
						twenty_four_hour_video_boost = $8,
						three_days_video_boost = $9,
						seven_days_video_boost = $10,
						total = $11,
						total_lifetime = $12,
						total_mob = $13,
						is_active = $14,
						updated_at = $15
				WHERE id = $1
					`
}

func (p *Point) queryExistsForUser() (qry string) {
	return `SELECT EXISTS(SELECT 1 FROM points WHERE user_id = $1)`
}

func (p *Point) queryGetByUserID() (qry string) {
	return `SELECT
						id,
						user_id,
						videos_watched,
						videos_voted,
						first_votes,
						correct_votes,
						ad_watched,
						referred_users,
						twenty_four_hour_video_boost,
						three_days_video_boost,
						seven_days_video_boost,
						total,
						total_lifetime,
						total_mob,
						is_active,
						created_at,
						updated_at

                 FROM	points
				 WHERE user_id = $1
				 ORDER BY created_at DESC
				 LIMIT 1
`
}

func (p *Point) queryTopUsers() (qry string) {
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
				WHERE users.is_active = true
				ORDER BY points.total DESC
				LIMIT $1
				OFFSET $2`
}

func (p *Point) queryTopMob() (qry string) {
	return `SELECT
					users.id,
				    users.facebook_id,
				    users.avatar,
				    users.name,
				    users.email,
					users.account_type,
					users.minutes_watched,
					points.total_mob,
					users.created_at,
					users.updated_at,
					users.encrypted_password,
					users.favourite_videos_count,
					users.imported_videos_count
				FROM users
				INNER JOIN points
				ON points.user_id = users.id
				WHERE points.total_mob > 0
				ORDER BY points.total_mob DESC
				LIMIT $1
				OFFSET $2`
}

func (p *Point) queryTopTalent() (qry string) {
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
					users.imported_videos_count,
           			(SELECT
               			COUNT(*)
					FROM votes
            		INNER JOIN videos
            		ON videos.id = votes.video_id
            		AND videos.user_id = users.id
            		WHERE upvote > 0)
             		as votes
				FROM  users
				WHERE users.id != 8
				AND users.id != 11
				AND users.id != 10
				ORDER BY votes DESC
				LIMIT $1
				OFFSET $2`
}

func (p *Point) validateCreateErrors() (err error) {
	if p.UserID == 0 {
		return p.Errors(ErrorMissingValue, "user_id")
	}
	return
}

func (p *Point) Create(db *system.DB) (err error) {

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
		p.ReferredUsers,
		p.TwentyFourHourVideoBoost,
		p.ThreeDaysVideoBoost,
		p.SevenDaysVideoBoost,
		p.Total,
		p.TotalLifetime,
		p.TotalMob,
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

func (p *Point) validateUpdateErrors() (err error) {

	if p.ID == 0 {
		return p.Errors(ErrorMissingValue, "id")
	}

	return p.validateCreateErrors()
}

func (p *Point) Update(db *system.DB) (err error) {

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
		p.ReferredUsers,
		p.TwentyFourHourVideoBoost,
		p.ThreeDaysVideoBoost,
		p.SevenDaysVideoBoost,
		p.Total,
		p.TotalLifetime,
		p.TotalMob,
		p.IsActive,
		p.UpdatedAt,
	)

	if err != nil {
		log.Printf("Point.Update() id -> %v Exec() -> %v \n Error -> %v", p.ID, p.queryUpdate(), err)
		return
	}

	log.Println("Point.Update() user_id -> ", p.UserID)

	return
}

func (p *Point) ExistsForUser(db *system.DB, userID uint64) (exists bool, err error) {

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
func (p *Point) AddToUsers(db *system.DB) (err error) {

	u := User{}
	users, err := u.GetAllUsers(db)

	if err != nil {
		return
	}

	for _, user := range users {

		exists, err := p.ExistsForUser(db, user.ID)

		if err != nil {
			log.Printf("Point.AddToUsers() userID -> %v Error -> %v", user.ID, err)
			return err
		}

		if exists {
			continue
		}

		point := Point{}
		point.UserID = user.ID

		if err = point.Create(db); err != nil {
			log.Println("Point.AddToUsers() Point.Create() Error -> ", err)
			return err
		}

	}

	return
}

func (p *Point) GetByUserID(db *system.DB, userID uint64) (err error) {

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
		&p.ReferredUsers,
		&p.TwentyFourHourVideoBoost,
		&p.ThreeDaysVideoBoost,
		&p.SevenDaysVideoBoost,
		&p.Total,
		&p.TotalLifetime,
		&p.TotalMob,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt)

	if err != nil {
		log.Printf("Point.GetByUserID() userID -> %v QueryRow() -> %v Error -> %v", userID, p.queryGetByUserID(), err)
		return
	}

	log.Printf("Point.GetByUserID() Point retrieved for user_id -> %v params -> %v", p.UserID, userID)
	return
}

func (p *Point) GetTopUsers(db *system.DB, page int) (users []User, err error) {

	rows, err := db.Query(p.queryTopUsers(), LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Point.GetTopUsers() Query() -> %v Error -> %v", p.queryTopUsers(), err)

		return
	}

	u := User{}

	return u.parseRows(rows)
}

func (p *Point) GetTopMob(db *system.DB, page int) (users []User, err error) {

	rows, err := db.Query(p.queryTopMob(), LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Point.GetTopMob() Query() -> %v Error -> %v", p.queryTopMob(), err)

		return
	}

	u := User{}

	return u.parseRows(rows)
}

func (p *Point) GetTopMob2(db *system.DB, userID uint64, page int) (users []User, err error) {

	qry := `SELECT
				users.id,
				users.facebook_id,
				users.avatar,
				users.name,
				users.email,
				users.account_type,
				users.minutes_watched,
				points.total_mob,
				users.created_at,
				users.updated_at,
				users.encrypted_password,
				users.favourite_videos_count,
				users.imported_videos_count,
				(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = users.id AND follower_id = $1 AND is_active = true))

			FROM users
			INNER JOIN points
			ON points.user_id = users.id
			WHERE points.total_mob > 0
			ORDER BY points.total_mob DESC
			LIMIT $2
			OFFSET $3`

	rows, err := db.Query(qry, userID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Point.GetTopMob() Query() -> %v Error -> %v", qry, err)
		return
	}

	u := User{}

	return u.parseRows2(rows)
}

func (p *Point) GetTopTalent(db *system.DB, page int) (users []User, err error) {
	rows, err := db.Query(p.queryTopTalent(), LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Point.GetTopTalent() Query() -> %v Error -> %v", p.queryTopTalent(), err)

		return
	}

	u := User{}

	return u.parseTalentRows(rows)
}

func (p *Point) GetTopTalent2(db *system.DB, userID uint64, page int) (users []User, err error) {

	qry := `SELECT
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
  					users.imported_videos_count,
	 				(SELECT
		 					COUNT(*)
  					FROM votes
  					INNER JOIN videos
  					ON videos.id = votes.video_id
  					AND videos.user_id = users.id
 					WHERE upvote > 0)
					as votes,
					(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = users.id AND follower_id = $1 AND is_active = true))

					  
			FROM  users
			WHERE users.id != 8
			AND users.id != 11
			AND users.id != 10
			ORDER BY votes DESC
			LIMIT $2
			OFFSET $3`

	rows, err := db.Query(qry, userID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Point.GetTopTalent() Query() -> %v Error -> %v", qry, err)

		return
	}

	u := User{}

	return u.parseTalentRows2(rows)
}
