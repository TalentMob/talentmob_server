package models

import (
	"database/sql"
	"log"
	"time"

	pq "github.com/lib/pq"

	"github.com/rathvong/talentmob_server/system"
)

// Competitor table keeps track of the weekly competitor
// for the videos that the users participate in
// the rules for this competition are ->
//
//1. user uploads video at any time of the week.
//2. video goes on a 7 day / 168 hour timer to gather votes.
//3. once the 7 days are up, voting for the weekly leaderboard for that video stops.
//4. once the last video added has reached its 7 days of votes,
//	the leaderboard is released for "The week of..." and displays the ranking.
//5. when the leaderboard is released, it notifies all participants via notifications,
//	with a deep link that takes them to the leaderboard
//6. the competition will end 12am midnight on Sunday
type Competitor struct {
	BaseModel
	UserID      uint64    `json:"user_id"`
	VideoID     uint64    `json:"video_id"`
	EventID     uint64    `json:"event_id"`
	Upvotes     uint64    `json:"up_votes"`
	Downvotes   uint64    `json:"down_votes"`
	VoteEndDate time.Time `json:"vote_end_date"`
	IsActive    bool      `json:"is_active"`
	IsUpvoted   bool      `json:"is_upvoted"`
	isDownvoted bool      `json:"is_downvoted"`
}

func (c *Competitor) queryCreate() (qry string) {
	return `INSERT INTO competitors
			(user_id,
			video_id,
			event_id,
			up_votes,
			down_votes,
			vote_end_date,
			is_active,
			created_at,
			updated_at)
			VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`
}

func (c *Competitor) queryGetByVideoID() (qry string) {
	return `
			SELECT
				id,
				user_id,
				video_id,
				event_id,
				up_votes,
				down_votes,
				vote_end_date,
				is_active,
				created_at,
				updated_at
			FROM competitors
			WHERE
				video_id = $1

	`
}

func (c *Competitor) queryGetByID() (qry string) {
	return `
			SELECT
				id,
				user_id,
				video_id,
				event_id,
				up_votes,
				down_votes,
				vote_end_date,
				is_active,
				created_at,
				updated_at
			FROM competitors
			WHERE
				id = $1

	`
}

func (c *Competitor) queryGetHistoryByCompetitionDate() (qry string) {
	return `SELECT
				id,
				user_id,
				video_id,
				event_id,
				up_votes,
				down_votes,
				vote_end_date,
				is_active,
				created_at,
				updated_at
			FROM competitors
			WHERE
				competition_date = $1
			ORDER BY up_votes DESC, down_votes ASC
			LIMIT $2,
			OFFSET $3
		`
}

func (c *Competitor) queryGetVideosByCompetitionDate() (qry string) {
	return `SELECT		videos.id,
						videos.user_id,
						videos.categories,
						competitors.down_votes,
						competitors.up_votes,
						videos.shares,
						videos.views,
						videos.comments,
						videos.thumbnail,
						videos.key,
						videos.title,
						videos.created_at,
						videos.updated_at,
						videos.is_active,
						competitors.vote_end_date,
						(SELECT EXISTS(select 1 from votes where user_id = $2 and video_id = videos.id and upvote > 0)),
						(SELECT EXISTS(select 1 from votes where user_id = $2 and video_id = videos.id and downvote > 0))

			FROM videos
			INNER JOIN competitors
			ON competitors.video_id = videos.id
			WHERE
				videos.is_active = true
			AND	competitors.is_active = true
			AND competitors.event_id = $1

			ORDER BY competitors.event_id, competitors.up_votes DESC, competitors.down_votes ASC
			LIMIT $3
			OFFSET $4`
}

func (c *Competitor) queryUpdate() (qry string) {
	return `
			UPDATE competitors SET
				user_id = $2,
				video_id = $3,
				event_id = $4,
				up_votes = $5,
				down_votes = $6,
				vote_end_date = $7,
				is_active = $8,
				updated_at = $9
			WHERE id = $1
			`
}

func (c *Competitor) querySoftDeleteByID() (qry string) {
	return `UPDATE competitors SET
				is_active = $2
			WHERE id = $1`
}

func (c *Competitor) validateCreateErrors() (err error) {
	if c.UserID == 0 {
		return c.Errors(ErrorMissingValue, "user_id")
	}

	if c.VideoID == 0 {
		return c.Errors(ErrorMissingValue, "video_id")
	}

	if c.EventID == 0 {
		return c.Errors(ErrorMissingValue, "event_id")

	}

	return
}

func (c *Competitor) validateUpdateErrors() (err error) {

	if c.ID == 0 {
		return c.Errors(ErrorMissingValue, "id")
	}

	return c.validateCreateErrors()
}

func (c *Competitor) addToEvent(db *system.DB) (err error) {
	event := Event{}
	if err = event.GetAvailableEvent(db); err != nil {
		return
	}

	c.EventID = event.ID

	return
}

func (c *Competitor) Register(db *system.DB, video Video) (err error) {
	c.UserID = video.UserID
	c.VideoID = video.ID

	return c.Create(db)
}

func (c *Competitor) Create(db *system.DB) (err error) {

	if err = c.addToEvent(db); err != nil {
		return
	}

	if err := c.validateCreateErrors(); err != nil {
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
			log.Println("Competitor.Create() Commit() - ", err)
			return
		}

	}()

	if err != nil {
		log.Println("Competitor.Create() Begin() - ", err)
		return
	}

	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	c.IsActive = true
	c.VoteEndDate = c.CreatedAt.Add(time.Hour * time.Duration(168))

	err = tx.QueryRow(c.queryCreate(),
		c.UserID,
		c.VideoID,
		c.EventID,
		c.Upvotes,
		c.Downvotes,
		c.VoteEndDate,
		c.IsActive,
		c.CreatedAt,
		c.UpdatedAt).Scan(&c.ID)

	if err != nil {
		log.Printf("Competitor.Create() UserID -> %v VideoID -> %v QueryRow() -> %v Error -> %v", c.UserID, c.VideoID, c.queryCreate(), err)
		return
	}

	return
}

//Validate if the vote is valid and updateable by the end date the video was created at
func (c *Competitor) IsVoteUpdateable() (isValid bool) {
	if c.ID == 0 {
		return false
	}

	currentTime := time.Now()

	if currentTime.Unix() <= c.VoteEndDate.Unix() {
		return true
	}

	return false
}

func (c *Competitor) Update(db *system.DB) (err error) {

	if err := c.validateUpdateErrors(); err != nil {
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
			log.Println("Competitor.Update() Commit() - ", err)
			return
		}

	}()

	if err != nil {
		log.Println("Competitor.Update() Begin() - ", err)
		return
	}

	c.UpdatedAt = time.Now()

	_, err = tx.Exec(c.queryUpdate(), c.ID,
		c.UserID,
		c.VideoID,
		c.EventID,
		c.Upvotes,
		c.Downvotes,
		c.VoteEndDate,
		c.IsActive,
		c.UpdatedAt)

	if err != nil {
		log.Printf("Competitor.Update() UserID -> %v VideoID -> %v QueryRow() -> %v Error -> %v", c.UserID, c.VideoID, c.queryUpdate(), err)
		return
	}

	return
}

func (c *Competitor) GetHistory(db *system.DB, eventID uint64, userID uint64, limit int, offset int) (videos []Video, err error) {
	if eventID == 0 {
		return videos, c.Errors(ErrorMissingValue, "event_id")
	}

	rows, err := db.Query(c.queryGetVideosByCompetitionDate(), eventID, userID, limit, offset)

	defer rows.Close()

	if err != nil {
		log.Printf("event_id -> %v Query() -> %v Error -> %v", eventID, c.queryGetVideosByCompetitionDate(), err)
		return
	}

	return c.parseRows(db, userID, rows)
}

func (c *Competitor) GetHistory2(db *system.DB, eventID uint64, userID uint64, limit int, offset int) (videos []Video, err error) {
	if eventID == 0 {
		return videos, c.Errors(ErrorMissingValue, "event_id")
	}

	qry := `SELECT		
	videos.id,
	videos.user_id,
	videos.categories,
	competitors.down_votes,
	competitors.up_votes,
	videos.shares,
	videos.views,
	videos.comments,
	videos.thumbnail,
	videos.key,
	videos.title,
	videos.created_at,
	videos.updated_at,
	videos.is_active,
	competitors.vote_end_date,
	(SELECT EXISTS(select 1 from votes where user_id = $2 and video_id = videos.id and upvote > 0)),
	(SELECT EXISTS(select 1 from votes where user_id = $2 and video_id = videos.id and downvote > 0)),
	users.id,
	users.name,
	users.avatar,
	users.account_type,
	users.created_at,
	users.updated_at,
	(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = competitors.user_id AND follower_id = $2 AND is_active = true)),
	boosts.id,
	boosts.user_id,
	boosts.video_id,
	boosts.start_time,
	boosts.end_time,
	boosts.is_active,
	boosts.created_at,
	boosts.updated_at	

FROM competitors
INNER JOIN videos
ON videos.id = competitors.video_id
LEFT JOIN users
ON users.id = competitors.user_id
LEFT JOIN boosts
ON boosts.video_id = competitors.video_id
AND boosts.is_active = true
AND boosts.end_time > now()
WHERE
videos.is_active = true
AND	competitors.is_active = true
AND competitors.event_id = $1

ORDER BY competitors.event_id, competitors.up_votes DESC, competitors.down_votes ASC
LIMIT $3
OFFSET $4`

	rows, err := db.Query(qry, eventID, userID, limit, offset)

	defer rows.Close()

	if err != nil {
		log.Printf("event_id -> %v Query() -> %v Error -> %v", eventID, qry, err)
		return
	}

	return c.parseRows2(db, userID, rows)
}

func (c *Competitor) parseRows2(db *system.DB, userID uint64, rows *sql.Rows) (videos []Video, err error) {

	for rows.Next() {
		video := Video{}

		var boostID sql.NullInt64
		var boostUserID sql.NullInt64
		var boostVideoID sql.NullInt64
		var boostIsActive sql.NullBool
		var boostStartTime pq.NullTime
		var boostEndTime pq.NullTime
		var boostCreatedAt pq.NullTime
		var boostUpdatedAt pq.NullTime

		var endDate time.Time
		err = rows.Scan(&video.ID,
			&video.UserID,
			&video.Categories,
			&video.Downvotes,
			&video.Upvotes,
			&video.Shares,
			&video.Views,
			&video.Comments,
			&video.Thumbnail,
			&video.Key,
			&video.Title,
			&video.CreatedAt,
			&video.UpdatedAt,
			&video.IsActive,
			&endDate,
			&video.IsUpvoted,
			&video.IsDownvoted,
			&video.Publisher.ID,
			&video.Publisher.Name,
			&video.Publisher.Avatar,
			&video.Publisher.AccountType,
			&video.Publisher.CreatedAt,
			&video.Publisher.UpdatedAt,
			&video.Publisher.IsFollowing,
			&boostID,
			&boostUserID,
			&boostVideoID,
			&boostStartTime,
			&boostEndTime,
			&boostIsActive,
			&boostCreatedAt,
			&boostUpdatedAt,
		)

		if err != nil {
			log.Println("Video.parseRows() Error -> ", err)
			return
		}

		if boostID.Valid {
			video.Boost.ID = uint64(boostID.Int64)
		}

		if boostUserID.Valid {
			video.Boost.UserID = uint64(boostUserID.Int64)
		}

		if boostVideoID.Valid {
			video.Boost.VideoID = uint64(boostVideoID.Int64)
		}

		if boostIsActive.Valid {
			video.Boost.IsActive = boostIsActive.Bool
		}

		if boostStartTime.Valid {
			video.Boost.StartTime = boostStartTime.Time
			video.Boost.StartTimeUnix = video.Boost.StartTime.UnixNano() / 1000000
		}

		if boostEndTime.Valid {
			video.Boost.EndTime = boostEndTime.Time
			video.Boost.EndTimeUnix = video.Boost.EndTime.UnixNano() / 1000000

		}

		if boostCreatedAt.Valid {
			video.Boost.CreatedAt = boostCreatedAt.Time
		}

		if boostUpdatedAt.Valid {
			video.Boost.UpdatedAt = boostUpdatedAt.Time
		}

		video.CompetitionEndDate = endDate.UnixNano() / 1000000

		videos = append(videos, video)
	}

	return
}

func (c *Competitor) parseRows(db *system.DB, userID uint64, rows *sql.Rows) (videos []Video, err error) {

	for rows.Next() {
		video := Video{}
		var endDate time.Time
		err = rows.Scan(&video.ID,
			&video.UserID,
			&video.Categories,
			&video.Downvotes,
			&video.Upvotes,
			&video.Shares,
			&video.Views,
			&video.Comments,
			&video.Thumbnail,
			&video.Key,
			&video.Title,
			&video.CreatedAt,
			&video.UpdatedAt,
			&video.IsActive,
			&endDate,
			&video.IsUpvoted,
			&video.IsDownvoted)

		if err != nil {
			log.Println("Video.parseRows() Error -> ", err)
			return
		}

		user := ProfileUser{}
		boost := Boost{}

		if err = user.GetUser(db, video.UserID); err != nil {
			return videos, err
		}

		video.CompetitionEndDate = endDate.UnixNano() / 1000000

		boost.GetByVideoID(db, video.ID)

		video.Boost = boost

		video.Publisher = user

		videos = append(videos, video)
	}

	return
}

func (c *Competitor) SoftDelete(db *system.DB, competitionID uint64) (err error) {

	if c.ID == 0 {
		return c.Errors(ErrorMissingID, "id")
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
		log.Println("Competitor.SoftDelete()", err)
		return
	}

	_, err = tx.Exec(c.querySoftDeleteByID(), c.ID)

	if err != nil {
		log.Printf("Video.SoftDelete() id -> %v Exec() -> %v Error -> %v", c.ID, c.querySoftDeleteByID(), err)
		return
	}

	return

}

func (c *Competitor) GetByVideoID(db *system.DB, videoID uint64) (err error) {

	if videoID == 0 {
		return c.Errors(ErrorMissingValue, "video_id")
	}

	err = db.QueryRow(c.queryGetByVideoID(), videoID).Scan(
		&c.ID,
		&c.UserID,
		&c.VideoID,
		&c.EventID,
		&c.Upvotes,
		&c.Downvotes,
		&c.VoteEndDate,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Competitor.GetByVideoID() videoID -> %v QueryRow() -> %v Error -> %v", videoID, c.queryGetByVideoID(), err)
		return
	}

	return
}

func (c *Competitor) Get(db *system.DB, competitorID uint64) (err error) {

	if competitorID == 0 {
		return c.Errors(ErrorMissingValue, "video_id")
	}

	err = db.QueryRow(c.queryGetByID(), competitorID).Scan(
		&c.ID,
		&c.UserID,
		&c.VideoID,
		&c.EventID,
		&c.Upvotes,
		&c.Downvotes,
		&c.VoteEndDate,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Competitor.Get() videoID -> %v QueryRow() -> %v Error -> %v", competitorID, c.queryGetByID(), err)
		return
	}

	return
}

// Add downvote for the competitor
func (c *Competitor) AddUpvote(db *system.DB) (err error) {
	c.Upvotes++

	if err = c.Update(db); err != nil {
		return
	}

	event := Event{}

	if err = event.Get(db, c.EventID); err != nil {
		return
	}

	event.UpvotesCount++

	return event.Update(db)
}

// Add upvote for the competitor
func (c *Competitor) AddDownvote(db *system.DB) (err error) {
	c.Downvotes++
	if err = c.Update(db); err != nil {
		return
	}

	event := Event{}

	if err = event.Get(db, c.EventID); err != nil {
		return
	}

	event.DownvotesCount++

	return event.Update(db)
}
