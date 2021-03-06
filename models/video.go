package models

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	pq "github.com/lib/pq"
	"github.com/rathvong/talentmob_server/system"
	"github.com/rathvong/talentmob_server/talentmobtranscoding"
)

// main structure for videos model
// videos files will not be uploaded from
// this server but directly uploaded via mobile to
// s3. The server will keep track of the association
// between the videos on the server and what videos the users
// needed to interact with the app
type Video struct {
	BaseModel
	Publisher           ProfileUser `json:"publisher"`
	UserID              uint64      `json:"user_id"`
	Categories          string      `json:"categories"`
	Downvotes           uint64      `json:"downvotes"`
	Upvotes             uint64      `json:"upvotes"`
	Shares              uint64      `json:"shares"`
	Views               uint64      `json:"views"`
	Comments            uint64      `json:"comments"`
	Thumbnail           string      `json:"thumbnail"`
	Key                 string      `json:"key"`
	Title               string      `json:"title"`
	IsActive            bool        `json:"is_active"`
	IsUpvoted           bool        `json:"is_upvoted"`
	IsDownvoted         bool        `json:"is_downvoted"`
	QueryRank           float64     `json:"query_rank"`
	Boost               Boost       `json:"boost"`
	CompetitionEndDate  int64       `json:"competition_end_date"`
	UpVoteTrendingCount uint        `json:"upvote_trending_count"`
	Priority            int
	EventID             uint64 `json:"event_id"`
}

// SQL query to create a row
func (v *Video) queryCreate() (qry string) {
	return `INSERT INTO videos
						(user_id,
						categories,
						downvotes,
						upvotes,
						shares,
						views,
						comments,
						thumbnail,
						key,
						title,
						created_at,
						updated_at,
						is_active)
			VALUES
						($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING 	id`
}

// SQL query to update a row
func (v *Video) queryUpdate() (qry string) {
	return `UPDATE videos SET
						user_id = $2,
						categories = $3,
						downvotes = $4,
						upvotes = $5,
						shares = $6,
						views = $7,
						comments = $8,
						thumbnail = $9,
						key = $10,
						title = $11,
						created_at = $12,
						updated_at = $13,
						is_active = $14,
						upvote_trending_count = $15
			WHERE	id = $1`
}

// SQL query for the users time-line
func (v *Video) queryTimeLine() (qry string) {
	return `  SELECT *
    FROM (
    (SELECT
             1 as priority,
             videos.id,
             videos.user_id,
             videos.categories,
             videos.downvotes,
             videos.upvotes,
             videos.shares,
             videos.views,
             videos.comments,
             videos.thumbnail,
             videos.key,
             videos.title,
             videos.created_at,
             videos.updated_at,
             videos.is_active,
             videos.upvote_trending_count
    FROM videos
    WHERE videos.id NOT IN (select video_id from votes where user_id = $1)
    AND videos.user_id != $1
    AND videos.is_active = true
    AND videos.upvote_trending_count > 4
    and videos.created_at > now()::date - 7
    ORDER BY upvote_trending_count DESC
    LIMIT 4
    ) UNION ALL (
    SELECT
            1 as priority,
            videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
            videos.upvote_trending_count
            FROM boosts
            INNER JOIN videos
            ON videos.id = boosts.video_id
            AND videos.user_id != $1
            AND videos.is_active = true
            WHERE boosts.is_active = true
            AND boosts.end_time >= now()
            AND boosts.video_id NOT IN (SELECT video_id from votes where user_id = $1)
            ORDER BY random()
			LIMIT 3
        ) UNION ALL (

                WITH recent_videos as (
                	SELECT
                	3 as priority, 
					   videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
            videos.upvote_trending_count,
					dense_rank()
						over(partition by user_id order by created_at desc) as the_ranking
					FROM videos
					WHERE videos.id NOT IN (select video_id from votes where user_id = $1)
 					AND videos.user_id != $1
					AND videos.is_active = true
					AND videos.upvote_trending_count <= 4
					OR videos.id NOT IN (select video_id from votes where user_id = $1)
					AND videos.user_id != $1
					AND videos.is_active = true
					AND videos.upvote_trending_count IS NULL
					ORDER BY videos.id DESC
				LIMIT 20
                )

                select
                  	3 as priority,
                   videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
            videos.upvote_trending_count
                from recent_videos videos
                where the_ranking = 1
                order by created_at DESC, upvote_trending_count DESC
        )
		
    ) as feed
    ORDER BY priority ASC
    LIMIT $2
    OFFSET $3
`
}

// SQL query for the users time-line
// has a larget amount of videos limit at 100
func (v *Video) queryDiscoveryTimeLine() (qry string) {
	return `  SELECT *
    FROM (
    (SELECT
             1 as priority,
             videos.id,
             videos.user_id,
             videos.categories,
             videos.downvotes,
             videos.upvotes,
             videos.shares,
             videos.views,
             videos.comments,
             videos.thumbnail,
             videos.key,
             videos.title,
             videos.created_at,
             videos.updated_at,
             videos.is_active,
             videos.upvote_trending_count
    FROM videos
    WHERE videos.id NOT IN (select video_id from votes where user_id = $1)
    AND videos.user_id != $1
    AND videos.is_active = true
    AND videos.upvote_trending_count > 1
    and videos.created_at > now()::date - 7
    ORDER BY upvote_trending_count DESC
    LIMIT 4
    ) UNION ALL (
    SELECT
            1 as priority,
            videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
            videos.upvote_trending_count
            FROM boosts
            INNER JOIN videos
            ON videos.id = boosts.video_id
            AND videos.user_id != $1
            AND videos.is_active = true
            WHERE boosts.is_active = true
            AND boosts.end_time >= now()
            AND boosts.video_id NOT IN (SELECT video_id from votes where user_id = $1)
            ORDER BY random()
			LIMIT 3
        ) UNION ALL (

                WITH recent_videos as (
                	SELECT
                	3 as priority, 
					   videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
            videos.upvote_trending_count,
					dense_rank()
						over(partition by user_id order by created_at desc) as the_ranking
					FROM videos
					WHERE videos.id NOT IN (select video_id from votes where user_id = $1)
 					AND videos.user_id != $1
					AND videos.is_active = true
					AND videos.upvote_trending_count <= 1
					OR videos.id NOT IN (select video_id from votes where user_id = $1)
					AND videos.user_id != $1
					AND videos.is_active = true
					AND videos.upvote_trending_count IS NULL
					ORDER BY videos.id DESC
				LIMIT 100
                )

                select
                  	3 as priority,
                   videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
            videos.upvote_trending_count
                from recent_videos videos
                where the_ranking = 1
                order by created_at DESC, upvote_trending_count DESC
        )
		
    ) as feed
    ORDER BY priority ASC
    LIMIT $2
    OFFSET $3
`
}

// SQL query for imported videos
func (v *Video) queryImportedVideos() (qry string) {
	return `SELECT		id,
						user_id,
						categories,
						downvotes,
						upvotes,
						shares,
						views,
						comments,
						thumbnail,
						key,
						title,
						created_at,
						updated_at,
						is_active,
						videos.upvote_trending_count
			FROM videos
			WHERE user_id = $1
			AND is_active = true
			ORDER BY created_at DESC
			LIMIT $2
			OFFSET $3 `
}

//SQL query for favourite videos
func (v *Video) queryFavouriteVideos() (qry string) {
	return `SELECT		videos.id,
						videos.user_id,
						videos.categories,
						videos.downvotes,
						videos.upvotes,
						videos.shares,
						videos.views,
						videos.comments,
						videos.thumbnail,
						videos.key,
						videos.title,
						videos.created_at,
						videos.updated_at,
						videos.is_active,
						videos.upvote_trending_count
			FROM videos
			LEFT JOIN votes
			ON votes.video_id = videos.id
			AND votes.upvote > 0
			WHERE votes.user_id = $1
			AND is_active = true
			ORDER BY votes.created_at DESC
			LIMIT $2
			OFFSET $3`
}

func (v *Video) queryHistory() (qry string) {
	return `SELECT		videos.id,
						videos.user_id,
						videos.categories,
						videos.downvotes,
						videos.upvotes,
						videos.shares,
						videos.views,
						videos.comments,
						videos.thumbnail,
						videos.key,
						videos.title,
						videos.created_at,
						videos.updated_at,
						videos.is_active,
						videos.upvote_trending_count
			FROM videos
			LEFT JOIN votes
			ON votes.video_id = videos.id
			WHERE votes.user_id = $1
			AND is_active = true
			ORDER BY votes.created_at DESC
			LIMIT $2
			OFFSET $3`
}

func (v *Video) queryLeaderBoard() (qry string) {
	return `SELECT		id,
						user_id,
						categories,
						downvotes,
						upvotes,
						shares,
						views,
						comments,
						thumbnail,
						key,
						title,
						created_at,
						updated_at,
						is_active,
						videos.upvote_trending_count
			FROM videos
			WHERE is_active = true
			ORDER BY upvotes DESC, downvotes ASC
			LIMIT $1
			OFFSET $2`
}

//SQL query for a single video
func (v *Video) queryVideoByID() (qry string) {
	return `SELECT 		id,
						user_id,
						categories,
						downvotes,
						upvotes,
						shares,
						views,
						comments,
						thumbnail,
						key,
						title,
						created_at,
						updated_at,
						is_active,
						videos.upvote_trending_count
			FROM videos
			WHERE id = $1`
}

func (v *Video) querySoftDeleteVideo() (qry string) {
	return `UPDATE videos SET
					is_active = false
			WHERE id = $1`
}

// This query will return videos by rank comparing against
// the lexemes found in videos.meta column.
// Only videos the user hasn't voted on will
// return a result.
func (v *Video) queryVideoByTitleAndCategory() (qry string) {
	return `SELECT
						v.id,
						v.user_id,
						v.categories,
						v.downvotes,
						v.upvotes,
						v.shares,
						v.views,
						v.comments,
						v.thumbnail,
						v.key,
						v.title,
						v.created_at,
						v.updated_at,
						v.is_active,
						v.rank,
						v.upvote_trending_count

				FROM (
						SELECT
						id,
						user_id,
						categories,
						downvotes,
						upvotes,
						shares,
						views,
						comments,
						thumbnail,
						key,
						title,
						created_at,
						updated_at,
						is_active,
						ts_rank_cd(meta, to_tsquery('%v'))	as rank,
						videos.upvote_trending_count
						FROM videos
						WHERE is_active = true
						AND user_id != $3
						AND id NOT IN (select video_id from votes where user_id = $3)
						) v
				WHERE v.rank > 0
				ORDER BY v.rank DESC
				LIMIT $1
				OFFSET $2`

}

// Recent videos  registered will
// show up in this query.
func (v *Video) queryRecentVideos() (qry string) {
	return `SELECT	videos.id,
						videos.user_id,
						videos.categories,
						videos.downvotes,
						videos.upvotes,
						videos.shares,
						videos.views,
						videos.comments,
						videos.thumbnail,
						videos.key,
						videos.title,
						videos.created_at,
						videos.updated_at,
						videos.is_active,
						videos.upvote_trending_count

			FROM videos
			WHERE is_active = true
			ORDER BY videos.created_at DESC
			LIMIT $1
			OFFSET $2 `
}

func (v *Video) queryUpvotedUsers() (qry string) {
	return `SELECT users.id,
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
				FROM votes
				INNER JOIN users
				ON users.id = votes.user_id
				WHERE votes.video_id = $1
				AND votes.upvote > 0
				LIMIT $2
				OFFSET $3`
}

// validate all important values needed for videos
func (v *Video) validateError() (err error) {

	if v.Categories == "" {
		return v.Errors(ErrorMissingValue, "categories")
	}

	if v.Title == "" {
		return v.Errors(ErrorMissingValue, "title")
	}

	if v.UserID == 0 {
		return v.Errors(ErrorMissingValue, "user_id")
	}

	if v.Key == "" {
		return v.Errors(ErrorMissingValue, "key")
	}

	return
}

// CreateForWeeklyEvents a new video
func (v *Video) CreateForWeeklyEvents(db *system.DB) (err error) {

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

		// Register video in this weeks competition
		compete := Competitor{}
		if err = compete.RegisterForWeeklyEvent(db, *v); err != nil {
			log.Println("competitor.Register() error: ", err)

		}

		// Create new categories
		category := Category{}
		category.CreateNewCategoriesFromTags(db, v.Categories, *v)

		if err := talentmobtranscoding.Transcode(v.ID); err != nil {
			log.Println("transcode err: ", err)
		}

		if err := talentmobtranscoding.TranscodeWithWatermark(v.ID); err != nil {
			log.Println("transcode with watermark err: ", err)
		}

	}()

	if err != nil {
		log.Println("Video.Create() Begin -> ", err)
		return
	}

	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
	v.IsActive = true

	err = tx.QueryRow(v.queryCreate(),
		v.UserID,
		v.Categories,
		v.Downvotes,
		v.Upvotes,
		v.Shares,
		v.Views,
		v.Comments,
		v.Thumbnail,
		v.Key,
		v.Title,
		v.CreatedAt,
		v.UpdatedAt,
		v.IsActive).Scan(&v.ID)

	if err != nil {
		log.Printf("Video.Create() QueryRow() -> %v Error -> %v", v.queryCreate(), err)
	}

	// Register video into competition
	return
}

func (v *Video) Create(db *system.DB) (err error) {

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

		// Register video in this weeks competition
		compete := Competitor{}
		if err = compete.RegisterForWeeklyEvent(db, *v); err != nil {
			log.Println("competitor.Register() error: ", err)
		}

		if v.EventID != 0 {
			compete := Competitor{}
			compete.VideoID = v.ID
			compete.UserID = v.UserID
			compete.EventID = v.EventID
			if err = compete.Create(db); err != nil {
				log.Println("competitor.Register() error: ", err)
			}

		}

		// Create new categories
		category := Category{}
		category.CreateNewCategoriesFromTags(db, v.Categories, *v)

		if err := talentmobtranscoding.Transcode(v.ID); err != nil {
			log.Println("transcode err: ", err)
		}

		if err := talentmobtranscoding.TranscodeWithWatermark(v.ID); err != nil {
			log.Println("transcode with watermark err: ", err)
		}

	}()

	if err != nil {
		log.Println("Video.Create() Begin -> ", err)
		return
	}

	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
	v.IsActive = true

	err = tx.QueryRow(v.queryCreate(),
		v.UserID,
		v.Categories,
		v.Downvotes,
		v.Upvotes,
		v.Shares,
		v.Views,
		v.Comments,
		v.Thumbnail,
		v.Key,
		v.Title,
		v.CreatedAt,
		v.UpdatedAt,
		v.IsActive).Scan(&v.ID)

	if err != nil {
		log.Printf("Video.Create() QueryRow() -> %v Error -> %v", v.queryCreate(), err)
	}

	// Register video into competition
	return
}

func (v *Video) SoftDelete(db *system.DB) (err error) {
	if v.ID == 0 {
		return v.Errors(ErrorMissingID, "id")
	}

	_, err = db.Exec(v.querySoftDeleteVideo(), v.ID)

	if err != nil {
		log.Printf("Video.SoftDelete() id -> %v Exec() -> %v Error -> %v", v.ID, v.querySoftDeleteVideo(), err)
		return
	}

	log.Println("Video.SoftDelete() video -> ", v.ID)
	return
}

func (v *Video) Update(db *system.DB) (err error) {

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
		log.Println("Video.Update() Begin -> ", err)
		return
	}

	_, err = tx.Exec(v.queryUpdate(),
		v.ID,
		v.UserID,
		v.Categories,
		v.Downvotes,
		v.Upvotes,
		v.Shares,
		v.Views,
		v.Comments,
		v.Thumbnail,
		v.Key,
		v.Title,
		v.CreatedAt,
		v.UpdatedAt,
		v.IsActive,
		v.UpVoteTrendingCount,
	)

	if err != nil {
		log.Printf("Video.Update() ID -> %v Exec() -> %v Error -> %v", v.ID, v.queryUpdate(), err)
		return
	}

	log.Print("Video.Update() Video successfully updated, id ->", v.ID)
	return
}

//Get users timeline
func (v *Video) GetTimeLine(db *system.DB, userID uint64, page int) (videos []Video, err error) {
	if userID == 0 {
		err = v.Errors(ErrorMissingValue, "userID")
		return
	}

	rows, err := db.Query(
		v.queryTimeLine(),
		userID,
		LimitQueryPerRequest,
		OffSet(page),
	)

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetTimeLine() userID -> %v Query -> %v Error -> %v", userID, v.queryTimeLine(), err)
	}

	return v.parseTimeLineRows(db, rows, userID, 0)
}

func (v *Video) GetTimeLine2(db *system.DB, userID uint64, page int) (videos []Video, err error) {
	if userID == 0 {
		err = v.Errors(ErrorMissingValue, "userID")
		return
	}

	qry := ` SELECT *
    FROM (
    (SELECT
             1 as priority,
             videos.id,
             videos.user_id,
             videos.categories,
             videos.downvotes,
             videos.upvotes,
             videos.shares,
             videos.views,
             videos.comments,
             videos.thumbnail,
             videos.key,
             videos.title,
             videos.created_at,
             videos.updated_at,
             videos.is_active,
			 videos.upvote_trending_count,
			 competitors.vote_end_date,
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and upvote > 0)),
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and downvote > 0)),
			users.id,
			users.name,
			users.avatar,
			users.account_type,
			users.created_at,
			users.updated_at,
			(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = competitors.user_id AND follower_id = $1 AND is_active = true)),
			boosts.id,
			boosts.user_id,
			boosts.video_id,
			boosts.start_time,
			boosts.end_time,
			boosts.is_active,
			boosts.created_at,
			boosts.updated_at
	FROM videos
	LEFT JOIN competitors
	ON competitors.video_id = videos.id
	LEFT JOIN users
	ON users.id = videos.user_id
	LEFT JOIN boosts
	ON boosts.video_id = videos.id
	AND boosts.is_active = true
	AND boosts.end_time > now()
    WHERE videos.id NOT IN (select video_id from votes where user_id = $1)
    AND videos.user_id != $1
    AND videos.is_active = true
    AND videos.upvote_trending_count > 4
    and videos.created_at > now()::date - 7
    ORDER BY upvote_trending_count DESC
    LIMIT 4
    ) UNION ALL (
    SELECT
            1 as priority,
            videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
			videos.upvote_trending_count,
			competitors.vote_end_date,
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and upvote > 0)),
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and downvote > 0)),
			users.id,
			users.name,
			users.avatar,
			users.account_type,
			users.created_at,
			users.updated_at,
			(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = competitors.user_id AND follower_id = $1 AND is_active = true)),
			boosts.id,
			boosts.user_id,
			boosts.video_id,
			boosts.start_time,
			boosts.end_time,
			boosts.is_active,
			boosts.created_at,
			boosts.updated_at
            FROM boosts
            INNER JOIN videos
            ON videos.id = boosts.video_id
            AND videos.user_id != $1
			AND videos.is_active = true
			LEFT JOIN competitors
			ON competitors.video_id = videos.id
			LEFT JOIN users
			ON users.id = videos.user_id
			WHERE boosts.is_active = true
            AND boosts.end_time >= now()
            AND boosts.video_id NOT IN (SELECT video_id from votes where user_id = $1)
            ORDER BY random()
			LIMIT 3
        ) UNION ALL (

                WITH recent_videos as (
                	SELECT
                	3 as priority, 
					   videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
			videos.upvote_trending_count,
					dense_rank()
						over(partition by user_id order by created_at desc) as the_ranking
					FROM videos
					WHERE videos.id NOT IN (select video_id from votes where user_id = $1)
 					AND videos.user_id != $1
					AND videos.is_active = true
					AND videos.upvote_trending_count <= 4
					OR videos.id NOT IN (select video_id from votes where user_id = $1)
					AND videos.user_id != $1
					AND videos.is_active = true
					AND videos.upvote_trending_count IS NULL
					ORDER BY videos.id DESC
				LIMIT 20
                )

                select
                  	3 as priority,
                   videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
			videos.upvote_trending_count,
			competitors.vote_end_date,
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and upvote > 0)),
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and downvote > 0)),
			users.id,
			users.name,
			users.avatar,
			users.account_type,
			users.created_at,
			users.updated_at,
			(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = competitors.user_id AND follower_id = $1 AND is_active = true)),
			boosts.id,
			boosts.user_id,
			boosts.video_id,
			boosts.start_time,
			boosts.end_time,
			boosts.is_active,
			boosts.created_at,
			boosts.updated_at
				from recent_videos videos
				LEFT JOIN competitors
				ON competitors.video_id = videos.id
				LEFT JOIN users
				ON users.id = videos.user_id
				LEFT JOIN boosts
				ON boosts.video_id = videos.id
				AND boosts.is_active = true
				AND boosts.end_time > now()
				where the_ranking = 1
                order by videos.created_at DESC, videos.upvote_trending_count DESC
        )
		
    ) as feed
    ORDER BY priority ASC
    LIMIT $2
    OFFSET $3`

	rows, err := db.Query(
		qry,
		userID,
		LimitQueryPerRequest,
		OffSet(page),
	)

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetTimeLine() userID -> %v Query -> %v Error -> %v", userID, qry, err)
	}

	return v.parseTimeLineRows2(db, rows, userID, 0)
}

func (v *Video) GetDiscoveryTimeLine(db *system.DB, userID uint64, page int) (videos []Video, err error) {
	if userID == 0 {
		err = v.Errors(ErrorMissingValue, "userID")
		return
	}

	rows, err := db.Query(
		v.queryDiscoveryTimeLine(),
		userID,
		LimitQueryPerRequest,
		OffSet(page),
	)

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetTimeLine() userID -> %v Query -> %v Error -> %v", userID, v.queryTimeLine(), err)
	}

	return v.parseTimeLineRows(db, rows, userID, 0)
}

func (v *Video) GetDiscoveryTimeLine2(db *system.DB, userID uint64, page int) (videos []Video, err error) {
	if userID == 0 {
		err = v.Errors(ErrorMissingValue, "userID")
		return
	}

	qry := ` SELECT *
    FROM (
    (SELECT
             1 as priority,
             videos.id,
             videos.user_id,
             videos.categories,
             videos.downvotes,
             videos.upvotes,
             videos.shares,
             videos.views,
             videos.comments,
             videos.thumbnail,
             videos.key,
             videos.title,
             videos.created_at,
             videos.updated_at,
             videos.is_active,
			 videos.upvote_trending_count,
			 competitors.vote_end_date,
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and upvote > 0)),
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and downvote > 0)),
			users.id,
			users.name,
			users.avatar,
			users.account_type,
			users.created_at,
			users.updated_at,
			(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = competitors.user_id AND follower_id = $1 AND is_active = true)),
			boosts.id,
			boosts.user_id,
			boosts.video_id,
			boosts.start_time,
			boosts.end_time,
			boosts.is_active,
			boosts.created_at,
			boosts.updated_at
	FROM videos
	LEFT JOIN competitors
	ON competitors.video_id = videos.id
	LEFT JOIN users
	ON users.id = videos.user_id
	LEFT JOIN boosts
	ON boosts.video_id = videos.id
	AND boosts.is_active = true
	AND boosts.end_time > now()
    WHERE videos.id NOT IN (select video_id from votes where user_id = $1)
    AND videos.user_id != $1
    AND videos.is_active = true
    AND videos.upvote_trending_count > 4
    and videos.created_at > now()::date - 7
    ORDER BY upvote_trending_count DESC
    LIMIT 4
    ) UNION ALL (
    SELECT
            1 as priority,
            videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
			videos.upvote_trending_count,
			competitors.vote_end_date,
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and upvote > 0)),
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and downvote > 0)),
			users.id,
			users.name,
			users.avatar,
			users.account_type,
			users.created_at,
			users.updated_at,
			(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = competitors.user_id AND follower_id = $1 AND is_active = true)),
			boosts.id,
			boosts.user_id,
			boosts.video_id,
			boosts.start_time,
			boosts.end_time,
			boosts.is_active,
			boosts.created_at,
			boosts.updated_at
            FROM boosts
            INNER JOIN videos
            ON videos.id = boosts.video_id
            AND videos.user_id != $1
			AND videos.is_active = true
			LEFT JOIN competitors
			ON competitors.video_id = videos.id
			LEFT JOIN users
			ON users.id = videos.user_id
            WHERE boosts.is_active = true
            AND boosts.end_time >= now()
            AND boosts.video_id NOT IN (SELECT video_id from votes where user_id = $1)
            ORDER BY random()
			LIMIT 3
        ) UNION ALL (

                WITH recent_videos as (
                	SELECT
                	3 as priority, 
					   videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
			videos.upvote_trending_count,
					dense_rank()
						over(partition by user_id order by created_at desc) as the_ranking
					FROM videos
					WHERE videos.id NOT IN (select video_id from votes where user_id = $1)
 					AND videos.user_id != $1
					AND videos.is_active = true
					AND videos.upvote_trending_count <= 4
					OR videos.id NOT IN (select video_id from votes where user_id = $1)
					AND videos.user_id != $1
					AND videos.is_active = true
					AND videos.upvote_trending_count IS NULL
					ORDER BY videos.id DESC
				LIMIT 100
                )

                select
                  	3 as priority,
                   videos.id,
            videos.user_id,
            videos.categories,
            videos.downvotes,
            videos.upvotes,
            videos.shares,
            videos.views,
            videos.comments,
            videos.thumbnail,
            videos.key,
            videos.title,
            videos.created_at,
            videos.updated_at,
            videos.is_active,
			videos.upvote_trending_count,
			competitors.vote_end_date,
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and upvote > 0)),
			(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and downvote > 0)),
			users.id,
			users.name,
			users.avatar,
			users.account_type,
			users.created_at,
			users.updated_at,
			(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = competitors.user_id AND follower_id = $1 AND is_active = true)),
			boosts.id,
			boosts.user_id,
			boosts.video_id,
			boosts.start_time,
			boosts.end_time,
			boosts.is_active,
			boosts.created_at,
			boosts.updated_at
				from recent_videos videos
				LEFT JOIN competitors
				ON competitors.video_id = videos.id
				LEFT JOIN users
				ON users.id = videos.user_id
				LEFT JOIN boosts
				ON boosts.video_id = videos.id
				AND boosts.is_active = true
				AND boosts.end_time > now()
                where the_ranking = 1
                order by videos.created_at DESC, videos.upvote_trending_count DESC
        )
		
    ) as feed
    ORDER BY priority ASC
    LIMIT $2
    OFFSET $3`

	rows, err := db.Query(
		qry,
		userID,
		LimitQueryPerRequest,
		OffSet(page),
	)

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetTimeLine() userID -> %v Query -> %v Error -> %v", userID, qry, err)
	}

	return v.parseTimeLineRows2(db, rows, userID, 0)
}

//Get users imported videos
func (v *Video) GetImportedVideos(db *system.DB, userID uint64, page int) (videos []Video, err error) {
	if userID == 0 {
		err = v.Errors(ErrorMissingValue, "userID")
		return
	}

	rows, err := db.Query(v.queryImportedVideos(), userID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetImportedVideos() userID -> %v Query() -> %v Error -> %v", userID, v.queryImportedVideos(), err)
		return
	}

	return v.parseRows(db, rows, userID, 0)
}

func (v *Video) GetImportedVideos2(db *system.DB, userID uint64, currentUserID uint64, page int) (videos []Video, err error) {
	if userID == 0 {
		err = v.Errors(ErrorMissingValue, "userID")
		return
	}

	if currentUserID == 0 {
		err = v.Errors(ErrorMissingValue, "currentUserID")
		return
	}

	qry := `SELECT		
				videos.id,
				videos.user_id,
				videos.categories,
				videos.downvotes,
				videos.upvotes,
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
			FROM videos
			LEFT JOIN competitors
			ON competitors.video_id = videos.id
			LEFT JOIN users
			ON users.id = videos.user_id
			LEFT JOIN boosts
			ON boosts.video_id = videos.id
			AND boosts.is_active = true
			AND boosts.end_time > now()
			WHERE videos.user_id = $1
			AND videos.is_active = true
			ORDER BY videos.created_at DESC
			LIMIT $3
			OFFSET $4 `

	rows, err := db.Query(qry, userID, currentUserID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetImportedVideos() userID -> %v Query() -> %v Error -> %v", userID, qry, err)
		return
	}

	return v.parseRows2(db, rows, userID, 0)
}

//Get users favourite videos
func (v *Video) GetFavouriteVideos(db *system.DB, userID uint64, page int) (videos []Video, err error) {
	if userID == 0 {
		err = v.Errors(ErrorMissingValue, "userID")
		return
	}

	rows, err := db.Query(v.queryFavouriteVideos(), userID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetFavouriteVideos() userID -> %v Query() -> %v Error -> %v", userID, v.queryFavouriteVideos(), err)
		return
	}

	return v.parseRows(db, rows, userID, 0)
}

func (v *Video) GetFavouriteVideos2(db *system.DB, userID uint64, currentUserID uint64, page int) (videos []Video, err error) {
	if userID == 0 {
		err = v.Errors(ErrorMissingValue, "userID")
		return
	}

	if currentUserID == 0 {
		err = v.Errors(ErrorMissingValue, "currentUserID")
		return
	}

	qry := `SELECT		
				videos.id,
				videos.user_id,
				videos.categories,
				videos.downvotes,
				videos.upvotes,
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
			FROM videos
			LEFT JOIN competitors
			ON competitors.video_id = videos.id
			LEFT JOIN users
			ON users.id = videos.user_id
			LEFT JOIN boosts
			ON boosts.video_id = videos.id
			AND boosts.is_active = true
			AND boosts.end_time > now()
			LEFT JOIN votes
			ON votes.video_id = videos.id
			AND votes.upvote > 0
			WHERE votes.user_id = $1
			AND videos.is_active = true
			ORDER BY votes.created_at DESC
			LIMIT $3
			OFFSET $4`

	rows, err := db.Query(qry, userID, currentUserID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetFavouriteVideos() userID -> %v Query() -> %v Error -> %v", userID, qry, err)
		return
	}

	return v.parseRows2(db, rows, userID, 0)
}

//Get users vote history
func (v *Video) GetHistory(db *system.DB, userID uint64, page int) (videos []Video, err error) {
	if userID == 0 {
		err = v.Errors(ErrorMissingValue, "userID")
		return
	}

	rows, err := db.Query(v.queryHistory(), userID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetHistory() userID -> %v Query() -> %v Error -> %v", userID, v.queryHistory(), err)
		return
	}

	return v.parseRows(db, rows, userID, 0)
}

//Get Leader board list
func (v *Video) GetLeaderBoard(db *system.DB, page int, userID uint64) (videos []Video, err error) {

	rows, err := db.Query(v.queryLeaderBoard(), LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetFavouriteVideos() Query() -> %v Error -> %v", v.queryLeaderBoard(), err)
		return
	}

	return v.parseRows(db, rows, userID, 0)
}

func (v *Video) GetLeaderBoard2(db *system.DB, page int, userID uint64) (videos []Video, err error) {

	qry := `SELECT		
	videos.id,
	videos.user_id,
	videos.categories,
	videos.downvotes,
	videos.upvotes,
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
	(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and upvote > 0)),
	(SELECT EXISTS(select 1 from votes where user_id = $1 and video_id = videos.id and downvote > 0)),
	users.id,
	users.name,
	users.avatar,
	users.account_type,
	users.created_at,
	users.updated_at,
	(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = competitors.user_id AND follower_id = $1 AND is_active = true)),
	boosts.id,
	boosts.user_id,
	boosts.video_id,
	boosts.start_time,
	boosts.end_time,
	boosts.is_active,
	boosts.created_at,
	boosts.updated_at
FROM videos
LEFT JOIN competitors
ON competitors.video_id = videos.id
LEFT JOIN users
ON users.id = videos.user_id
LEFT JOIN boosts
ON boosts.video_id = videos.id
AND boosts.is_active = true
AND boosts.end_time > now()
WHERE videos.is_active = true
ORDER BY videos.upvotes DESC, videos.downvotes ASC
LIMIT $2
OFFSET $3`

	rows, err := db.Query(qry, userID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Video.GetFavouriteVideos() Query() -> %v Error -> %v", qry, err)
		return
	}

	return v.parseRows2(db, rows, userID, 0)
}

func (v *Video) GetVideoByID(db *system.DB, id uint64) (err error) {

	if id == 0 {
		return v.Errors(ErrorMissingID, "id")
	}

	var trending sql.NullInt64
	err = db.QueryRow(v.queryVideoByID(), id).Scan(&v.ID,
		&v.UserID,
		&v.Categories,
		&v.Downvotes,
		&v.Upvotes,
		&v.Shares,
		&v.Views,
		&v.Comments,
		&v.Thumbnail,
		&v.Key,
		&v.Title,
		&v.CreatedAt,
		&v.UpdatedAt,
		&v.IsActive,
		&trending)

	if err != nil {
		log.Printf("Video.GetVideoByID() id -> %v QueryRow() -> %v Error -> %v", id, v.queryVideoByID(), err)
	}

	if trending.Valid {
		v.UpVoteTrendingCount = uint(trending.Int64)
	}

	return
}

func (v *Video) GetVideoByID2(db *system.DB, id uint64) (err error) {

	if id == 0 {
		return v.Errors(ErrorMissingID, "id")
	}

	qry := `SELECT 		videos.id,
						videos.user_id,
						videos.categories,
						videos.downvotes,
						videos.upvotes,
						videos.shares,
						videos.views,
						videos.comments,
						videos.thumbnail,
						videos.key,
						videos.title,
						videos.created_at,
						videos.updated_at,
						videos.is_active,
						videos.upvote_trending_count,
						users.id,
						users.avatar,
						users.name,
						users.account_type,
						users.created_at,
						users.updated_at,
						boosts.id,
						boosts.user_id,
						boosts.video_id,
						boosts.start_time,
						boosts.end_time,
						boosts.is_active,
						boosts.created_at,
						boosts.updated_at,
						competitors.vote_end_date					
		FROM videos
		LEFT JOIN users
		ON users.id = videos.user_id
		AND users.is_active = true
		LEFT JOIN boosts
		ON boosts.video_id = videos.id
		AND boosts.is_active = true
		AND boosts.end_time > now()	
		LEFT JOIN competitors
		ON competitors.video_id = videos.id
		WHERE videos.id = $1
		AND videos.is_active = true
		`

	var trending sql.NullInt64
	var boostID sql.NullInt64
	var boostUserID sql.NullInt64
	var boostVideoID sql.NullInt64
	var boostIsActive sql.NullBool
	var boostStartTime pq.NullTime
	var boostEndTime pq.NullTime
	var boostCreatedAt pq.NullTime
	var boostUpdatedAt pq.NullTime
	var endDate pq.NullTime

	err = db.QueryRow(qry, id).Scan(&v.ID,
		&v.UserID,
		&v.Categories,
		&v.Downvotes,
		&v.Upvotes,
		&v.Shares,
		&v.Views,
		&v.Comments,
		&v.Thumbnail,
		&v.Key,
		&v.Title,
		&v.CreatedAt,
		&v.UpdatedAt,
		&v.IsActive,
		&trending,
		&v.Publisher.ID,
		&v.Publisher.Avatar,
		&v.Publisher.Name,
		&v.Publisher.AccountType,
		&v.Publisher.CreatedAt,
		&v.Publisher.UpdatedAt,
		&boostID,
		&boostUserID,
		&boostVideoID,
		&boostStartTime,
		&boostEndTime,
		&boostIsActive,
		&boostCreatedAt,
		&boostUpdatedAt,
		&endDate,
	)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Video.GetVideoByID() id -> %v QueryRow() -> %v Error -> %v", id, qry, err)
		return
	}

	if trending.Valid {
		v.UpVoteTrendingCount = uint(trending.Int64)
	}

	if boostID.Valid {
		v.Boost.ID = uint64(boostID.Int64)
	}

	if boostUserID.Valid {
		v.Boost.UserID = uint64(boostUserID.Int64)
	}

	if boostVideoID.Valid {
		v.Boost.VideoID = uint64(boostVideoID.Int64)
	}

	if boostIsActive.Valid {
		v.Boost.IsActive = boostIsActive.Bool
	}

	if boostStartTime.Valid {
		v.Boost.StartTime = boostStartTime.Time
		v.Boost.StartTimeUnix = v.Boost.StartTime.UnixNano() / 1000000
	}

	if boostEndTime.Valid {
		v.Boost.EndTime = boostEndTime.Time
		v.Boost.EndTimeUnix = v.Boost.EndTime.UnixNano() / 1000000
	}

	if boostCreatedAt.Valid {
		v.Boost.CreatedAt = boostCreatedAt.Time
	}

	if boostUpdatedAt.Valid {
		v.Boost.UpdatedAt = boostUpdatedAt.Time
	}

	if endDate.Valid {
		v.CompetitionEndDate = endDate.Time.UnixNano() / 1000000
	}

	return
}

func (v *Video) Find(db *system.DB, qry string, page int, userID uint64, weekInterval int) (video []Video, err error) {

	log.Println("Video.Find() Query String -> ", qry)

	rows, err := db.Query(fmt.Sprintf(v.queryVideoByTitleAndCategory(), qry), LimitQueryPerRequest, OffSet(page), userID)

	defer rows.Close()

	if err != nil {
		log.Printf("Video.Find() qry -> %v page -> %v -> userID ->%v Query() -> %v Error -> %v", qry, page, userID, fmt.Sprintf(v.queryVideoByTitleAndCategory(), qry), err)
		return
	}

	return v.parseQueryRows(db, rows, userID, weekInterval)
}

func (v *Video) Find2(db *system.DB, q string, page int, userID uint64, weekInterval int) (video []Video, err error) {

	log.Println("Video.Find() Query String -> ", q)

	qry := `SELECT
	v.id,
	v.user_id,
	v.categories,
	v.downvotes,
	v.upvotes,
	v.shares,
	v.views,
	v.comments,
	v.thumbnail,
	v.key,
	v.title,
	v.created_at,
	v.updated_at,
	v.is_active,
	v.upvote_trending_count,
	competitors.vote_end_date,
   (SELECT EXISTS(select 1 from votes where user_id = $3 and video_id = v.id and upvote > 0)),
   (SELECT EXISTS(select 1 from votes where user_id = $3 and video_id = v.id and downvote > 0)),
   users.id,
   users.name,
   users.avatar,
   users.account_type,
   users.created_at,
   users.updated_at,
   (SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = competitors.user_id AND follower_id = $3 AND is_active = true)),
   boosts.id,
   boosts.user_id,
   boosts.video_id,
   boosts.start_time,
   boosts.end_time,
   boosts.is_active,
   boosts.created_at,
   boosts.updated_at


FROM (
	SELECT
	id,
	user_id,
	categories,
	downvotes,
	upvotes,
	shares,
	views,
	comments,
	thumbnail,
	key,
	title,
	created_at,
	updated_at,
	is_active,
	ts_rank_cd(meta, to_tsquery('%v'))	as rank,
	videos.upvote_trending_count
	FROM videos
	WHERE is_active = true
	AND user_id != $3
	AND id NOT IN (select video_id from votes where user_id = $3)
	) v

	LEFT JOIN competitors
ON competitors.video_id = v.id
LEFT JOIN users
ON users.id = v.user_id
LEFT JOIN boosts
ON boosts.video_id = v.id
AND boosts.is_active = true
AND boosts.end_time > now()
WHERE v.rank > 0
ORDER BY v.rank DESC, v.created_at DESC
LIMIT $1
OFFSET $2`

	rows, err := db.Query(fmt.Sprintf(qry, q), LimitQueryPerRequest, OffSet(page), userID)

	defer rows.Close()

	if err != nil {
		log.Printf("Video.Find() qry -> %v page -> %v -> userID ->%v Query() -> %v Error -> %v", q, page, userID, fmt.Sprintf(qry, q), err)
		return
	}

	return v.parseQueryRows2(db, rows, userID, weekInterval)
}

func (v *Video) Recent(db *system.DB, userID uint64, page int, weeklyInterval int) (videos []Video, err error) {

	rows, err := db.Query(v.queryRecentVideos(), LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Video.Recent() Query() -> %v Error -> %v", v.queryRecentVideos(), err)
		return
	}

	return v.parseRows(db, rows, userID, weeklyInterval)
}

//Parse rows for video queries
func (v *Video) parseRows(db *system.DB, rows *sql.Rows, userID uint64, weekInterval int) (videos []Video, err error) {

	for rows.Next() {
		video := Video{}

		var trending sql.NullInt64

		err = rows.Scan(
			&video.ID,
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
			&trending,
		)

		if err != nil {
			log.Println("Video.parseRows() Error -> ", err)
			return
		}

		vote := Vote{}
		user := ProfileUser{}
		boost := Boost{}

		if trending.Valid {
			video.UpVoteTrendingCount = uint(trending.Int64)
		}

		if video.IsUpvoted, err = vote.HasUpVoted(db, userID, video.ID); err != nil {
			return videos, err
		}

		if video.IsDownvoted, err = vote.HasDownVoted(db, userID, video.ID); err != nil {
			return videos, err
		}

		if err = user.GetUser(db, video.UserID); err != nil {
			return videos, err
		}

		compete := Competitor{}

		compete.GetByVideoID(db, video.ID)

		video.CompetitionEndDate = compete.VoteEndDate.UnixNano() / 1000000

		boost.GetByVideoID(db, video.ID)

		video.Boost = boost

		video.Publisher = user

		videos = append(videos, video)
	}

	return
}

func (v *Video) parseRows2(db *system.DB, rows *sql.Rows, userID uint64, weekInterval int) (videos []Video, err error) {

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

		var endDate pq.NullTime
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

		if endDate.Valid {
			video.CompetitionEndDate = endDate.Time.UnixNano() / 1000000
		}

		videos = append(videos, video)
	}

	return
}

//Parse rows for video queries
func (v *Video) parseTimeLineRows(db *system.DB, rows *sql.Rows, userID uint64, weekInterval int) (videos []Video, err error) {

	for rows.Next() {
		video := Video{}

		var trending sql.NullInt64
		var priority sql.NullInt64

		err = rows.Scan(
			&priority,
			&video.ID,
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
			&trending,
		)

		if err != nil {
			log.Println("Video.parseTimeLineRows() Error -> ", err)
			return
		}

		vote := Vote{}
		user := ProfileUser{}
		boost := Boost{}

		if trending.Valid {
			video.UpVoteTrendingCount = uint(trending.Int64)
		}

		if priority.Valid {
			video.Priority = int(priority.Int64)
		}

		if video.IsUpvoted, err = vote.HasUpVoted(db, userID, video.ID); err != nil {
			return videos, err
		}

		if video.IsDownvoted, err = vote.HasDownVoted(db, userID, video.ID); err != nil {
			return videos, err
		}

		if err = user.GetUser(db, video.UserID); err != nil {
			return videos, err
		}

		r := new(Relationship)
		user.IsFollowing, _ = r.IsFollowing(db, user.ID, userID)

		boost.GetByVideoID(db, video.ID)

		video.Boost = boost

		video.Publisher = user

		videos = append(videos, video)
	}

	return
}

func (v *Video) parseTimeLineRows2(db *system.DB, rows *sql.Rows, userID uint64, weekInterval int) (videos []Video, err error) {

	for rows.Next() {
		video := Video{}

		var trending sql.NullInt64
		var priority sql.NullInt64

		var boostID sql.NullInt64
		var boostUserID sql.NullInt64
		var boostVideoID sql.NullInt64
		var boostIsActive sql.NullBool
		var boostStartTime pq.NullTime
		var boostEndTime pq.NullTime
		var boostCreatedAt pq.NullTime
		var boostUpdatedAt pq.NullTime

		var endDate pq.NullTime
		err = rows.Scan(
			&priority,
			&video.ID,
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
			&trending,
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

		if trending.Valid {
			video.UpVoteTrendingCount = uint(trending.Int64)
		}

		if priority.Valid {
			video.Priority = int(priority.Int64)
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

		if endDate.Valid {
			video.CompetitionEndDate = endDate.Time.UnixNano() / 1000000
		}

		videos = append(videos, video)
	}

	return
}

//Parse rows for video queries
func (v *Video) parseQueryRows(db *system.DB, rows *sql.Rows, userID uint64, weekInterval int) (videos []Video, err error) {

	for rows.Next() {

		var trending sql.NullInt64

		video := Video{}

		err = rows.Scan(
			&video.ID,
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
			&video.QueryRank,
			&trending,
		)

		if err != nil {
			log.Println("Video.parseQueryRows() Error -> ", err)
			return
		}

		vote := Vote{}
		user := ProfileUser{}
		boost := Boost{}

		if trending.Valid {
			video.UpVoteTrendingCount = uint(trending.Int64)
		}

		if video.IsUpvoted, err = vote.HasUpVoted(db, userID, video.ID); err != nil {
			return videos, err
		}

		if video.IsDownvoted, err = vote.HasDownVoted(db, userID, video.ID); err != nil {
			return videos, err
		}

		if err = user.GetUser(db, video.UserID); err != nil {
			return videos, err
		}

		boost.GetByVideoID(db, video.ID)

		video.Boost = boost
		video.Publisher = user

		videos = append(videos, video)
	}

	return
}

func (v *Video) parseQueryRows2(db *system.DB, rows *sql.Rows, userID uint64, weekInterval int) (videos []Video, err error) {

	for rows.Next() {

		video := Video{}

		var trending sql.NullInt64
		var priority sql.NullInt64

		var boostID sql.NullInt64
		var boostUserID sql.NullInt64
		var boostVideoID sql.NullInt64
		var boostIsActive sql.NullBool
		var boostStartTime pq.NullTime
		var boostEndTime pq.NullTime
		var boostCreatedAt pq.NullTime
		var boostUpdatedAt pq.NullTime

		var endDate pq.NullTime
		err = rows.Scan(
			&video.ID,
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
			&trending,
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

		if trending.Valid {
			video.UpVoteTrendingCount = uint(trending.Int64)
		}

		if priority.Valid {
			video.Priority = int(priority.Int64)
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

		if endDate.Valid {
			video.CompetitionEndDate = endDate.Time.UnixNano() / 1000000
		}

		videos = append(videos, video)
	}

	return
}

func (v *Video) UpVotedUsers(db *system.DB, videoID uint64, page int) (users []User, err error) {
	if videoID == 0 {
		return users, v.Errors(ErrorMissingValue, "video.UpVotedUsers() videoID = 0")
	}

	rows, err := db.Query(v.queryUpvotedUsers(), videoID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("UpVotedUsers() videoID -> %v query() -> %v error -> %v", videoID, v.queryUpvotedUsers(), err)
		return
	}

	return v.ParseUserRows(db, rows)
}

func (v *Video) UpVotedUsers2(db *system.DB, videoID uint64, userID uint64, page int) (users []User, err error) {
	if videoID == 0 {
		return users, v.Errors(ErrorMissingValue, "video.UpVotedUsers() videoID = 0")
	}

	qry := `SELECT 	users.id,
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
					(SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = users.id AND follower_id = $2 AND is_active = true))

			FROM votes
			INNER JOIN users
			ON users.id = votes.user_id
			WHERE votes.video_id = $1
			AND votes.upvote > 0
			LIMIT $3
			OFFSET $4`

	rows, err := db.Query(qry, videoID, userID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("UpVotedUsers() videoID -> %v query() -> %v error -> %v", videoID, qry, err)
		return
	}

	return v.ParseUserRows2(db, rows)
}

/**
Parse data rows retrieve by followers and following query
*/
func (v *Video) ParseUserRows(db *system.DB, rows *sql.Rows) (users []User, err error) {

	for rows.Next() {
		user := User{}

		err = rows.Scan(
			&user.ID,
			&user.FacebookID,
			&user.Avatar,
			&user.Name,
			&user.Email,
			&user.AccountType,
			&user.MinutesWatched,
			&user.Points,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.EncryptedPassword,
			&user.FavouriteVideosCount,
			&user.ImportedVideosCount,
		)

		if err != nil {
			log.Println("Video.ParseRows()", err)
			return
		}

		users = append(users, user)
	}

	return
}

func (v *Video) ParseUserRows2(db *system.DB, rows *sql.Rows) (users []User, err error) {

	for rows.Next() {
		user := User{}

		err = rows.Scan(
			&user.ID,
			&user.FacebookID,
			&user.Avatar,
			&user.Name,
			&user.Email,
			&user.AccountType,
			&user.MinutesWatched,
			&user.Points,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.EncryptedPassword,
			&user.FavouriteVideosCount,
			&user.ImportedVideosCount,
			&user.IsFollowing,
		)

		if err != nil {
			log.Println("Video.ParseRows()", err)
			return
		}

		users = append(users, user)
	}

	return
}

func (v *Video) HasPriority(videos []Video) bool {
	for _, v := range videos {
		if v.Priority < 3 {
			return true
		}
	}

	return false
}

func (v *Video) Shuffle(input []Video) (outputArray []Video) {
	log.Println("Shuffling Videos")

	inputLength := len(input)
	// add these lines here to create a local slice []int
	inputArray := make([]Video, inputLength)
	copy(inputArray, input)

	for i := 0; i < inputLength; i++ {
		randomNum := generateRandom(inputArray)
		outputArray = append(outputArray, inputArray[randomNum])
		inputArray = append(inputArray[:randomNum], inputArray[(randomNum+1):]...)
	}

	return outputArray
}

func generateRandom(input []Video) int {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	return random.Intn(len(input))
}
