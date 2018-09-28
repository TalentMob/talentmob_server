package models

import (
	"database/sql"

	"log"
	"time"

	"github.com/rathvong/talentmob_server/system"
)

type Comment struct {
	Publisher ProfileUser `json:"publisher"`
	BaseModel
	UserID     uint64      `json:"user_id"`
	VideoID    uint64      `json:"video_id"`
	Title      string      `json:"title"`
	Content    string      `json:"content"`
	IsActive   bool        `json:"is_active"`
	Object     interface{} `json:"object"`
	ObjectType string      `json:"object_type"`
}

func (c *Comment) queryCreate() (qry string) {
	return `INSERT INTO comments
					(user_id, video_id, title, content, is_active, created_at, updated_at)

			VALUES
					($1, $2, $3, $4, $5, $6, $7)

			RETURNING id`
}

func (c *Comment) queryUpdate() (qry string) {
	return `UPDATE comments set
						title = $2,
						content = $3,
						is_active = $4,
						updated_at = $5
			WHERE id = $1`
}

func (c *Comment) queryGetByID() (qry string) {
	return `SELECT
				id,
				user_id,
				video_id,
				title,
				content,
				is_active,
				created_at,
				updated_at
			FROM comments
			WHERE id = $1`
}

func (c *Comment) queryGetByVideo() (qry string) {
	return `SELECT
				id,
				user_id,
				video_id,
				title,
				content,
				is_active,
				created_at,
				updated_at
			FROM comments
			WHERE video_id = $1
			AND is_active = true
			ORDER BY created_at DESC
			LIMIT $2
			OFFSET $3`
}

func (c *Comment) Create(db *system.DB) (err error) {

	if err = c.validateErrors(); err != nil {
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

		video := Video{}
		if err = video.GetVideoByID(db, c.VideoID); err != nil {
			panic(err)
			return
		}

		if video.UserID != c.UserID {
			Notify(db, c.UserID, video.UserID, VERB_COMMENTED, c.ID, OBJECT_COMMENT)
		}

	}()

	if err != nil {
		log.Println("Comment.Create() Begin() ", err)
		return
	}

	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	c.IsActive = true

	err = tx.QueryRow(c.queryCreate(),
		c.UserID,
		c.VideoID,
		c.Title,
		c.Content,
		c.IsActive,
		c.CreatedAt,
		c.UpdatedAt).Scan(&c.ID)

	if err != nil {
		log.Printf("Comment.Create() user_id -> %v video_id -> %v QueryRow() -> %v Error -> %v", c.UserID, c.VideoID, c.queryCreate(), err)
		return
	}

	return
}

func (c *Comment) Update(db *system.DB) (err error) {
	if err = c.validateErrors(); err != nil {
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
		log.Println("Comment.Update() Begin() ", err)
		return
	}

	c.UpdatedAt = time.Now()

	_, err = tx.Exec(c.queryUpdate(),
		c.ID,
		c.Title,
		c.Content,
		c.IsActive,
		c.UpdatedAt)

	if err != nil {
		log.Printf("Comment.Update() id -> %v QueryRow() -> %v Error -> %v", c.ID, c.queryUpdate(), err)
		return
	}

	return
}

func (c *Comment) Get(db *system.DB, commentID uint64) (err error) {
	if commentID == 0 {
		return c.Errors(ErrorMissingID, "id")
	}

	err = db.QueryRow(c.queryGetByID(), commentID).Scan(
		&c.ID,
		&c.UserID,
		&c.VideoID,
		&c.Title,
		&c.Content,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt)

	if err != nil {
		log.Printf("Comment.Get() commentID -> %v QueryRow() -> %v Error -> %v", commentID, c.queryGetByID(), err)
		return
	}

	return
}

func (c *Comment) GetForVideo(db *system.DB, videoID uint64, page int) (comments []Comment, err error) {
	if videoID == 0 {
		return comments, c.Errors(ErrorMissingValue, "videoID")
	}

	rows, err := db.Query(c.queryGetByVideo(), videoID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("Comment.GetForVideo() videoID -> %v Query() -> %v Error -> %v", videoID, c.queryGetByVideo(), err)
		return
	}

	return c.parseRows(db, rows)
}

func (c *Comment) GetForVideo2(db *system.DB, videoID uint64, page int) (comments []Comment, err error) {
	if videoID == 0 {
		return comments, c.Errors(ErrorMissingValue, "videoID")
	}

	qry := `SELECT
				comments.id,
				comments.user_id,
				comments.video_id,
				comments.title,
				comments.content,
				comments.is_active,
				comments.created_at,
				comments.updated_at,
				users.id,
				users.avatar,
				users.name,
				users.account_type,
				users.created_at,
				users.updated_at
			FROM comments
			INNER JOIN users
			ON users.id = comments.user_id
			AND users.is_active = true
			WHERE comments.video_id = $1
			AND comments.is_active = true
			ORDER BY comments.created_at DESC
			LIMIT $2
			OFFSET $3`

	rows, err := db.Query(qry, videoID, LimitQueryPerRequest, OffSet(page))

	defer rows.Close()

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Comment.GetForVideo() videoID -> %v Query() -> %v Error -> %v", videoID, qry, err)
		return
	}

	return c.parseRows2(db, rows)
}

func (c *Comment) GetVideo(db *system.DB) (video Video, err error) {

	if c.VideoID == 0 {
		return video, c.Errors(ErrorMissingValue, "video_id")
	}

	return video, video.GetVideoByID(db, c.VideoID)
}

func (c *Comment) validateErrors() (err error) {

	if c.UserID == 0 {
		return c.Errors(ErrorMissingValue, "user_id")
	}

	if c.VideoID == 0 {
		return c.Errors(ErrorMissingValue, "video_id")
	}

	return
}

func (c *Comment) parseRows(db *system.DB, rows *sql.Rows) (comments []Comment, err error) {

	for rows.Next() {

		comment := Comment{}

		err = rows.Scan(
			&comment.ID,
			&comment.UserID,
			&comment.VideoID,
			&comment.Title,
			&comment.Content,
			&comment.IsActive,
			&comment.CreatedAt,
			&comment.UpdatedAt)

		if err != nil {
			log.Print("Comment.parseRows() Error -> ", err)
			return
		}

		comment.Publisher.GetUser(db, comment.UserID)

		comments = append(comments, comment)
	}

	return
}

func (c *Comment) parseRows2(db *system.DB, rows *sql.Rows) (comments []Comment, err error) {

	for rows.Next() {

		comment := Comment{}

		err = rows.Scan(
			&comment.ID,
			&comment.UserID,
			&comment.VideoID,
			&comment.Title,
			&comment.Content,
			&comment.IsActive,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.Publisher.ID,
			&comment.Publisher.Avatar,
			&comment.Publisher.Name,
			&comment.Publisher.AccountType,
			&comment.Publisher.CreatedAt,
			&comment.Publisher.UpdatedAt,
		)

		if err != nil {
			log.Print("Comment.parseRows() Error -> ", err)
			return
		}

		comments = append(comments, comment)
	}

	return
}
