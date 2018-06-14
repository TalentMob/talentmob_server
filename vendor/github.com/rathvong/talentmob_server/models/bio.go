package models

import (
	"github.com/rathvong/talentmob_server/system"
	"log"
)

//Users bio information
type Bio struct {
	BaseModel
	UserID       uint64 `json:"user_id"`
	Bio          string `json:"bio"`
	CatchPhrases string `json:"catch_phrases"`
	Awards       string `json:"awards"`
}

// SQL query to retrieve a users BIO
func (b *Bio) queryGet() (qry string) {
	return `SELECT 	id,
					user_id,
					bio,
					catch_phrases,
					awards,
					created_at,
					updated_at
			FROM bios
			WHERE user_id = $1
			ORDER BY created_at ASC
			LIMIT 1`
}

// SQL query to update a users BIO
func (b *Bio) queryUpdate() (qry string) {
	return `UPDATE bios set
						bio = $2,
						catch_phrases = $3,
						awards = $4,
						updated_at = $5
			WHERE id = $1`
}

// Retrieve a users bio by ID
func (b *Bio) Get(db *system.DB, userID uint64) (err error) {

	if userID == 0 {
		return b.Errors(ErrorMissingValue, "userID")
	}

	err = db.QueryRow(b.queryGet(), userID).Scan(&b.ID,
		&b.UserID,
		&b.Bio,
		&b.CatchPhrases,
		&b.Awards,
		&b.CreatedAt,
		&b.UpdatedAt)

	if err != nil {
		log.Printf("Bio.Get() userID -> %v QueryRow() -> %v Error -> %v", userID, b.queryGet(), err)
		return
	}

	return
}

// Update a users bio by ID
func (b *Bio) Update(db *system.DB) (err error) {
	if b.ID == 0 {
		return b.Errors(ErrorMissingID, "id")
	}

	if b.UserID == 0 {
		return b.Errors(ErrorMissingValue, "UserID")
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
		log.Println("Bio.Update() Begin() ", err)

		return
	}

	_, err = tx.Exec(b.queryUpdate(),
		b.ID,
		b.Bio,
		b.CatchPhrases,
		b.Awards,
		b.UpdatedAt)

	if err != nil {
		log.Printf("Bio.Update() id -> %v Exec() -> %v Error -> %v", b.ID, b.queryUpdate(), err)
		return
	}

	return
}
