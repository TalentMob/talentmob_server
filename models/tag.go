package models

import (
	"github.com/rathvong/talentmob_server/system"
	"log"
	"time"
)

// Tags are used to associate a category with a video
// We can use this model to track trending and video count
// associated with a category.
// This model is the binding between Video and Category
// and should be created every time a new video is registered
// to track trends.
type Tag struct {
	BaseModel
	VideoID    uint64 `json:"video_id"`
	CategoryID uint64 `json::"category_id"`
	Title      string `json:"title"`
	IsActive   bool   `json:"is_active"`
}

func (t *Tag) queryCreate() (qry string) {
	return `INSERT INTO tags
						(video_id,
						category_id,
						title,
						is_active,
						created_at,
						updated_at)
				 VALUES
						($1, $2, $3, $4, $5, $6)
				 RETURNING id`
}

func (t *Tag) queryUpdate() (qry string) {
	return `UPDATE tags SET
						video_id = $2,
						category_id = $3,
						title = $4,
						is_active = $5,
						updated_at = $6
				WHERE
						id = $1`
}

func (t *Tag) validateCreateErrors() (err error) {
	if t.Title == "" {
		return t.Errors(ErrorMissingValue, "title")
	}

	if t.VideoID == 0 {
		return t.Errors(ErrorMissingValue, "video_id")
	}

	if t.CategoryID == 0 {
		return t.Errors(ErrorMissingValue, "category_id")
	}

	return
}

func (t *Tag) validateUpdateErrors() (err error) {
	if t.ID == 0 {
		return t.Errors(ErrorMissingValue, "id")
	}

	return t.validateCreateErrors()
}

// Create new tags
func (t *Tag) Create(db *system.DB) (err error) {

	if err = t.validateCreateErrors(); err != nil {

		log.Println("Tag.Create() Error -> ", err)
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
		log.Println("Tag.Create() Begin() Error -> ", err)
		return
	}

	t.IsActive = true
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	err = tx.QueryRow(t.queryCreate(),
		t.VideoID,
		t.CategoryID,
		t.Title,
		t.IsActive,
		t.CreatedAt,
		t.UpdatedAt,
	).Scan(&t.ID)

	if err != nil {
		log.Printf("Tag.Create() title -> %v QueryRow() -> %v Error -> %v", t.Title, t.queryCreate(), err)
		return
	}

	return
}

// Update tags
func (t *Tag) Update(db *system.DB) (err error) {

	if err = t.validateUpdateErrors(); err != nil {
		log.Println("Tag.Update() Error -> ", err)
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
		log.Println("Update.Tag() Begin() Error -> ", err)
		return
	}

	_, err = tx.Exec(t.queryUpdate(),
		t.VideoID,
		t.CategoryID,
		t.Title,
		t.IsActive,
		t.UpdatedAt,
	)

	if err != nil {
		log.Printf("Update.Tag() id -> %v Exec() -> %v Error -> %v", t.ID, t.queryUpdate(), err)
		return
	}

	return
}
