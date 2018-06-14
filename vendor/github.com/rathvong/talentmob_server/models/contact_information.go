package models

import (
	"github.com/rathvong/talentmob_server/system"
	"log"
	"time"
)

type ContactInformation struct {
	BaseModel
	UserID      uint64 `json:"user_id"`
	PhoneNumber string `json:"phone_number"`
	InstagramID string `json:"instagram_id"`
}

func (c *ContactInformation) queryCreate() (qry string) {
	return `INSERT INTO contact_information 
						(user_id, phone_number, instagram_id, created_at, updated_at)
						VALUES
						($1, $2, $3, $4, $5)
						returning id`
}

func (c *ContactInformation) queryUpdate() (qry string) {
	return `UPDATE contact_information SET 
						user_id = $2,
						phone_number = $3,
						instagram_id = $4,
						update_at = $5
				WHERE id = $1`
}

func (c *ContactInformation) queryPhoneNumber() (qry string) {
	return `SELECT 
						id,
						user_id,
						phone_number,
						instagram_id,
						created_at,
						updated_at
				FROM
						contact_information
				WHERE phone_number = $1
`
}

func (c *ContactInformation) queryInstagramID() (qry string) {
	return `SELECT 
						id,
						user_id,
						phone_number,
						instagram_id,
						created_at,
						updated_at
				FROM
						contact_information
				WHERE instagram_id = $1`
}

func (c *ContactInformation) queryPhoneExists() (qry string) {
	return `SELECT EXISTS(select 1 from contact_information where phone_number = $1)`
}

func (c *ContactInformation) queryInstagramExists() (qry string) {
	return `SELECT EXISTS(select 1 from contact_information where instagram_id = $1)`
}

func (c *ContactInformation) validateCreate() (err error) {

	if len(c.PhoneNumber) == 0 && len(c.InstagramID) == 0 {
		return c.Errors(ErrorMissingValue, "missing contact")
	}

	if c.UserID == 0 {
		return c.Errors(ErrorMissingValue, "user_id")
	}

	return
}

func (c *ContactInformation) validateUpdate() (err error) {
	if c.ID == 0 {
		return c.Errors(ErrorMissingValue, "id")
	}

	return c.validateCreate()
}

func (c *ContactInformation) Create(db *system.DB) (err error) {

	if err = c.validateCreate(); err != nil {
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
		log.Println("ContactInformation.Create() Error -> ", err)
		return
	}

	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	err = tx.QueryRow(
		c.queryCreate(),
		c.UserID,
		c.PhoneNumber,
		c.InstagramID,
		c.CreatedAt,
		c.UpdatedAt,
	).Scan(&c.ID)

	if err != nil {
		log.Printf("ContactInformation.Create() Query -> %v Error -> %v", c.queryCreate(), err)
		return
	}

	return
}

func (c *ContactInformation) Update(db *system.DB) (err error) {

	if err = c.validateUpdate(); err != nil {
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

	c.UpdatedAt = time.Now()

	_, err = tx.Exec(
		c.queryUpdate(),
		c.UserID,
		c.PhoneNumber,
		c.InstagramID,
		c.UpdatedAt,
	)

	if err != nil {

		log.Printf("ContactInformation.Update() id -> %v query -> %v err -> %v", c.ID, c.queryUpdate(), err)
		return
	}

	return
}

func (c *ContactInformation) GetPhone(db *system.DB, number string) (err error) {

	if len(number) == 0 {
		return c.Errors(ErrorMissingValue, "ContactInformation.GetPhone() missing phone number")
	}

	err = db.QueryRow(c.queryPhoneNumber(), number).Scan(
		&c.ID,
		&c.UserID,
		&c.PhoneNumber,
		&c.InstagramID,
		&c.CreatedAt,
		&c.UpdatedAt)

	if err != nil {
		log.Printf("ContactInformation.GetPhone() id -> %v query -> %v error -> %v", number, c.queryPhoneNumber(), err)
		return
	}

	return
}

func (c *ContactInformation) GetInstagram(db *system.DB, id string) (err error) {
	if len(id) == 0 {
		return c.Errors(ErrorMissingValue, "ContactInformation.GetPhone() missing phone number")
	}

	err = db.QueryRow(c.queryInstagramID(), id).Scan(
		&c.ID,
		&c.UserID,
		&c.PhoneNumber,
		&c.InstagramID,
		&c.CreatedAt,
		&c.UpdatedAt)

	if err != nil {
		log.Printf("ContactInformation.GetInstagramID() id -> %v query -> %v error -> %v", id, c.queryPhoneNumber(), err)
		return
	}

	return
}

func (c *ContactInformation) ExistsPhone(db *system.DB, number string) (exists bool) {

	err := db.QueryRow(c.queryPhoneExists(), number).Scan(&exists)

	if err != nil {
		log.Printf("ContactInformation.ExistsPhone() number -> %v query -> %v error -> %v", number, c.queryPhoneExists(), err)
		return false
	}

	return exists
}

func (c *ContactInformation) ExistsInstagram(db *system.DB, id string) (exists bool) {
	err := db.QueryRow(c.queryPhoneExists(), id).Scan(&exists)

	if err != nil {
		log.Printf("ContactInformation.ExistsInstagram() number -> %v query -> %v error -> %v", id, c.queryPhoneExists(), err)
		return false
	}

	return exists
}
