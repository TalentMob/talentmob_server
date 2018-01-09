package models

import (
	"log"
	"time"
	"database/sql"
	"github.com/rathvong/talentmob_server/system"

	"fmt"

)

// main structure for videos model
// videos files will not be uploaded from
// this server but directly uploaded via mobile to
// s3. The server will keep track of the association
// between the videos on the server and what videos the users
// needed to interact with the app
type Video struct {
	BaseModel
	Publisher   ProfileUser `json:"publisher"`
	UserID      uint64      `json:"user_id"`
	Categories  string      `json:"categories"`
	Downvotes   uint64      `json:"downvotes"`
	Upvotes     uint64      `json:"upvotes"`
	Shares      uint64      `json:"shares"`
	Views       uint64      `json:"views"`
	Comments    uint64      `json:"comments"`
	Thumbnail   string      `json:"thumbnail"`
	Key         string      `json:"key"`
	Title       string      `json:"title"`
	IsActive    bool        `json:"is_active"`
	IsUpvoted   bool        `json:"is_upvoted"`
	IsDownvoted bool        `json:"is_downvoted"`
	QueryRank   float64     `json:"query_rank"`
	Boost       Boost       `json:"boost"`
}

// SQL query to create a row
func (v *Video) queryCreate() (qry string){
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
func (v *Video) queryUpdate() (qry string){
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
						is_active = $14
			WHERE	id = $1`
}

// SQL query for the users time-line
func (v *Video) queryTimeLine() (qry string){
	return `SELECT *
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
            videos.is_active
            FROM boosts
            LEFT JOIN videos
            ON videos.id = boosts.video_id
            AND videos.user_id != 1
            AND videos.is_active = true
            AND videos.id NOT IN (select video_id from votes where user_id = $1)

            WHERE boosts.end_time >= now()
            AND boosts.is_active = true
            ORDER BY end_time DESC
        ) UNION ALL (
        SELECT
            2 as priority,
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
            videos.is_active
         FROM videos
         WHERE videos.id NOT IN (select video_id from votes where user_id = $1)
         AND videos.user_id != $1
         AND videos.is_active = true
         ORDER BY created_at DESC
         LIMIT 100
         OFFSET 0
         )) as feed
    LIMIT $2
    OFFSET $3;`
}

// SQL query for imported videos
func (v *Video) queryImportedVideos() (qry string){
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
						is_active
			FROM videos
			WHERE user_id = $1
			AND is_active = true
			ORDER BY created_at DESC
			LIMIT $2
			OFFSET $3 `
}

//SQL query for favourite videos
func (v *Video) queryFavouriteVideos() (qry string){
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
						videos.is_active
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

func (v *Video) queryHistory() (qry string){
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
						videos.is_active
			FROM videos
			LEFT JOIN votes
			ON votes.video_id = videos.id
			WHERE votes.user_id = $1
			AND is_active = true
			ORDER BY votes.created_at DESC
			LIMIT $2
			OFFSET $3`
}

func (v *Video) queryLeaderBoard() (qry string){
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
						is_active
			FROM videos
			WHERE is_active = true
			ORDER BY upvotes DESC, downvotes ASC
			LIMIT $1
			OFFSET $2`
}

//SQL query for a single video
func (v *Video) queryVideoByID() (qry string){
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
						is_active
			FROM videos
			WHERE id = $1`
}

func (v *Video) querySoftDeleteVideo() (qry string){
	return `UPDATE videos SET
					is_active = false
			WHERE id = $1`
}

// This query will return videos by rank comparing against
// the lexemes found in videos.meta column.
// Only videos the user hasn't voted on will
// return a result.
func (v *Video) queryVideoByTitleAndCategory() (qry string){
	return `SELECT
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
						rank

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
							ts_rank_cd(meta, to_tsquery('%v'))	as rank
						FROM videos
						WHERE is_active = true
						AND user_id != $3
						AND id NOT IN (select video_id from votes where user_id = $3)
						) v
				WHERE rank > 0
				ORDER BY rank DESC
				LIMIT $1
				OFFSET $2`

	}


// Recent videos  registered will
// show up in this query.
func (v *Video) queryRecentVideos() (qry string){
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

			FROM videos
			WHERE is_active = true
			ORDER BY videos.created_at DESC
			LIMIT $1
			OFFSET $2 `
}

	// validate all important values needed for videos
	func (v *Video) validateError() (err error){

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


	// Create a new video
	func (v *Video) Create(db *system.DB) (err error){

		if err = v.validateError(); err != nil {
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
				return
			}

			// Register video in this weeks competition
			compete := Competitor{}
			if err = compete.Register(db, *v); err != nil {
				tx.Rollback()
				return
			}

			// Create new categories
			category := Category{}
			category.CreateNewCategoriesFromTags(db, v.Categories, *v)

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
			v.IsActive)

		if err != nil {
			log.Printf("Video.Update() ID -> %v Exec() -> %v Error -> %v", v.ID, v.queryUpdate(), err)
			return
		}

		log.Print("Video.Update() Video successfully updated, id ->", v.ID)
		return
	}



	//Get users timeline
	func (v *Video) GetTimeLine(db *system.DB, userID uint64, page int) (videos []Video, err error){
		if userID == 0 {
			err = v.Errors(ErrorMissingValue, "userID")
			return
		}

		rows, err := db.Query(v.queryTimeLine(), userID, LimitQueryPerRequest, offSet(page))

		defer rows.Close()

		if err != nil {
			log.Printf("Video.GetTimeLine() userID -> %v Query -> %v Error -> %v", userID, v.queryTimeLine(), err)
		}

		return v.parseTimelineRows(db, rows, userID, 0)
	}


	//Get users imported videos
	func (v *Video) GetImportedVideos(db *system.DB, userID uint64, page int) (videos []Video, err error){
		if userID == 0 {
			err = v.Errors(ErrorMissingValue, "userID")
			return
		}

		rows, err := db.Query(v.queryImportedVideos(), userID, LimitQueryPerRequest,  offSet(page))

		defer rows.Close()

		if err != nil {
			log.Printf("Video.GetImportedVideos() userID -> %v Query() -> %v Error -> %v", userID,  v.queryImportedVideos(), err)
			return
		}

		return v.parseRows(db, rows, userID, 0)
	}


	//Get users favourite videos
	func (v *Video) GetFavouriteVideos(db *system.DB, userID uint64, page int) (videos []Video, err error){
		if userID == 0 {
			err = v.Errors(ErrorMissingValue, "userID")
			return
		}

		rows, err := db.Query(v.queryFavouriteVideos(), userID, LimitQueryPerRequest, offSet(page))

		defer rows.Close()

		if err != nil {
			log.Printf("Video.GetFavouriteVideos() userID -> %v Query() -> %v Error -> %v", userID, v.queryFavouriteVideos(), err)
			return
		}

		return v.parseRows(db, rows, userID, 0)
	}


	//Get users vote history
	func (v *Video) GetHistory(db *system.DB, userID uint64, page int) (videos []Video, err error){
		if userID == 0 {
			err = v.Errors(ErrorMissingValue, "userID")
			return
		}

		rows, err := db.Query(v.queryHistory(), userID, LimitQueryPerRequest, offSet(page))

		defer rows.Close()

		if err != nil {
			log.Printf("Video.GetHistory() userID -> %v Query() -> %v Error -> %v", userID, v.queryHistory(), err)
			return
		}

		return v.parseRows(db, rows, userID, 0)
	}


	//Get Leader board list
	func (v *Video) GetLeaderBoard(db *system.DB, page int, userID uint64) (videos []Video, err error){


		rows, err := db.Query(v.queryLeaderBoard(), LimitQueryPerRequest, offSet(page))

		defer rows.Close()

		if err != nil {
			log.Printf("Video.GetFavouriteVideos() Query() -> %v Error -> %v", v.queryLeaderBoard(), err)
			return
		}

		return v.parseRows(db, rows, userID, 0)
	}


	func (v *Video) GetVideoByID(db *system.DB, id uint64) (err error){

		if id == 0 {
			return v.Errors(ErrorMissingID, "id")
		}

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
			&v.IsActive)

		if err != nil {
			log.Printf("Video.GetVideoByID() id -> %v QueryRow() -> %v Error -> %v", id, v.queryVideoByID(), err )
		}


		return
	}


	func (v *Video) Find(db *system.DB,  qry string, page int, userID uint64, weekInterval int) (video []Video, err error){



		 log.Println("Video.Find() Query String -> ", qry)

		 rows, err := db.Query(fmt.Sprintf(v.queryVideoByTitleAndCategory(), qry),  LimitQueryPerRequest, offSet(page), userID)

		defer rows.Close()

		if err != nil {
			log.Printf("Video.Find() qry -> %v page -> %v -> userID ->%v Query() -> %v Error -> %v", qry,  page, userID,fmt.Sprintf(v.queryVideoByTitleAndCategory(),qry), err)
			return
		}

		return v.parseQueryRows(db, rows, userID, weekInterval)
	}


	func (v *Video) Recent(db *system.DB, userID uint64, page int, weeklyInterval int) (videos []Video, err error){

		rows, err := db.Query(v.queryRecentVideos(), LimitQueryPerRequest, offSet(page))

		defer rows.Close()

		if err != nil  {
			log.Printf("Video.Recent() Query() -> %v Error -> %v", v.queryRecentVideos(), err)
			return
		}

		return v.parseRows(db, rows, userID, weeklyInterval)
	}



	//Parse rows for video queries
	func (v *Video) parseRows(db *system.DB, rows *sql.Rows, userID uint64, weekInterval int) (videos []Video, err error){

		for rows.Next() {
			video := Video{}

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
					)

			if err != nil {
				log.Println("Video.parseRows() Error -> ", err)
				return
			}

			vote := Vote{}
			user := ProfileUser{}
			boost := Boost{}

			if video.IsUpvoted, err = vote.HasUpVoted(db, userID, video.ID, weekInterval); err != nil {
				return videos, err
			}

			if video.IsDownvoted, err = vote.HasDownVoted(db, userID, video.ID, weekInterval); err != nil {
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

//Parse rows for video queries
func (v *Video) parseTimelineRows(db *system.DB, rows *sql.Rows, userID uint64, weekInterval int) (videos []Video, err error){

	for rows.Next() {
		video := Video{}

		var priority int

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
		)

		if err != nil {
			log.Println("Video.parseRows() Error -> ", err)
			return
		}

		vote := Vote{}
		user := ProfileUser{}
		boost := Boost{}

		if video.IsUpvoted, err = vote.HasUpVoted(db, userID, video.ID, weekInterval); err != nil {
			return videos, err
		}

		if video.IsDownvoted, err = vote.HasDownVoted(db, userID, video.ID, weekInterval); err != nil {
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

//Parse rows for video queries
func (v *Video) parseQueryRows(db *system.DB, rows *sql.Rows, userID uint64, weekInterval int) (videos []Video, err error) {

	for rows.Next() {
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
		)

		if err != nil {
			log.Println("Video.parseQueryRows() Error -> ", err)
			return
		}

		vote := Vote{}
		user := ProfileUser{}
		boost := Boost{}

		if video.IsUpvoted, err = vote.HasUpVoted(db, userID, video.ID, weekInterval); err != nil {
			return videos, err
		}

		if video.IsDownvoted, err = vote.HasDownVoted(db, userID, video.ID, weekInterval); err != nil {
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