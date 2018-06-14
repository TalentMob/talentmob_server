package models

import (
	"database/sql"
	"time"
	"log"

	"github.com/rathvong/talentmob_server/system"
)

type Transcoded struct {
	BaseModel
	VideoID uint64 `json:"video_id"`
	TranscodedWatermarkKey string `json:"transcoded_watermark_key"`
	TranscodedKey string `json:"transcoded_key"`
	TranscodedThumbnailKey string `json:"transcoded_thumbnail_key"`
	WatermarkCompleted bool `json:"watermark_completed"`
	TranscodedCompleted bool `json:"transcoded_completed"`
	IsActive bool `json:"is_active"`
}

func (t *Transcoded) queryExists() string {
	return `SELECT EXISTS(select 1 from transcoded where video_id = $1)`
}

func (t *Transcoded) queryCreate() string {
	return `INSERT INTO transcoded (
							video_id, 
							transcoded_watermark_key,
							transcoded_key,
							transcoded_thumbnail_key,
							watermark_completed,
							transcoded_complated,
							is_active,
							created_at,
							updated_at
								)
						VALUES
						($1, $2, $3, $4, $5, $6, $7, $8, $9)
						RETURNING id
`
}

func (t *Transcoded) queryUpdate() string {
	return `UPDATE transcoded SET 
				transcoded_watermark_key = $2,
				transcoded_key = $2,
				transcoded_thumbnail_key = $3,
				watermark_completed = $4,
				transcoded_complated = $5,
				is_active = $6,
				updated_at = $7
			WHERE id = $1
`
}

func (t *Transcoded) queryNeedTranscodedWatermarkVideo() string {
	return `SELECT 
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
			LEFT JOIN transcoded
			ON transcoded.video_id != videos.id
			AND transcoded.completed_transcode_watermark != true
			AND transcoded.is_active = true
			AND videos.is_active = true
			ORDER BY videos.created_at DESC
`
}

func (t *Transcoded) Update(db *system.DB) error {

	if t.ID == 0 {
		return t.Errors(ErrorMissingID, "transcoded: id")
	}

	if t.VideoID == 0 {
		return t.Errors(ErrorMissingID, "transcoded: video_id")
	}

	if t.TranscodedKey == "" {
		return t.Errors(ErrorMissingValue, "transcoded: transcoded_key")
	}

	if t.TranscodedWatermarkKey == "" {
		return t.Errors(ErrorMissingValue, "transcoded: transcoded_watermark_key")
	}

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
		}
	}()

	if err != nil {
		return err
	}

	t.UpdatedAt = time.Now()

	_, err = tx.Exec(
		t.queryUpdate(),
		t.ID,
		t.TranscodedWatermarkKey,
		t.TranscodedKey,
		t.TranscodedThumbnailKey,
		t.WatermarkCompleted,
		t.TranscodedCompleted,
		t.IsActive,
		t.UpdatedAt,
	)


	if err != nil {
		log.Printf("Transcoded.Update() Query() -> %v Error -> %v", t.queryUpdate(), err)
		return err
	}

	return nil

}

func (t *Transcoded) Exists(db *system.DB, videoID uint64) bool {

	var exists bool
	err := db.QueryRow(t.queryExists(), videoID).Scan(&exists)

	if err != nil {
		log.Println("transcoded.Exists() Error: ", err)
		// true so it doesn't create anything if there is an error
		return true
	}

	return exists
}

func (t *Transcoded) Create(db *system.DB) error {

	if t.VideoID == 0 {
		return t.Errors(ErrorMissingID, "transcoded: video_id")
	}

	if t.TranscodedKey == "" {
		return t.Errors(ErrorMissingValue, "transcoded: transcoded_key")
	}

	if t.TranscodedWatermarkKey == "" {
		return t.Errors(ErrorMissingValue, "transcoded: transcoded_watermark_key")
	}

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
		}
	}()

	if err != nil {
		return err
	}


	t.IsActive = true
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	err = tx.QueryRow(
		t.queryCreate(),
		t.VideoID,
		t.TranscodedWatermarkKey,
		t.TranscodedKey,
		t.TranscodedThumbnailKey,
		t.WatermarkCompleted,
		t.TranscodedCompleted,
		t.IsActive,
		t.CreatedAt,
		t.UpdatedAt,
		).Scan(&t.ID)


	if err != nil {
		log.Printf("Transcoded.Create() Query() -> %v Error -> %v", t.queryCreate(), err)
		return err
	}

	return nil
}

func (t *Transcoded) GetNeedsTranscodedWatermarkVideos(db *system.DB) (videos []Video, err error){
	rows, err := db.Query(t.queryNeedTranscodedWatermarkVideo())

	defer rows.Close()

	if err != nil {
		log.Printf("transcoded.GetNeedsTranscodedWatermarkVideos() Query() -> %v Error: %v", t.queryNeedTranscodedWatermarkVideo(), err)
		return videos, err
	}

	return t.parseVideos(rows)
}

func (t *Transcoded) parseVideos(rows *sql.Rows) ([]Video, error){

	var videos []Video

	videos = make([]Video, 0)

	for rows.Next() {
		video := Video{}
		var trending sql.NullInt64

		err := rows.Scan(
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
			return nil, err
		}

		if trending.Valid {
			video.UpVoteTrendingCount = uint(trending.Int64)
		}

		videos = append(videos, video)
	}

	return videos, nil
}