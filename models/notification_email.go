package models

import (
	"github.com/rathvong/talentmob_server/system"
	"errors"
	"time"
	"log"
)

type NotificationEmail struct {
	BaseModel
	Address string `json:"address"`
	IsActive bool `json:"is_active"`

}


func (n *NotificationEmail) queryCreate() (qry string){
	return `INSERT INTO notification_emails
						(address, is_active, created_at, updated_at)
				VALUES	
						($1, $2, $3, $4)
				RETURNING id`
}



func (n *NotificationEmail) Create(db *system.DB) (err error){


	if n.Address == "" {
		err = errors.New("NotificationEmail.Create() Error -> Missing address")
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
		log.Println("", err)
		return
	}

	n.IsActive = true
	n.UpdatedAt = time.Now()
	n.CreatedAt = time.Now()


	err = tx.QueryRow(
		n.queryCreate(),
		n.Address,
		n.IsActive,
		n.CreatedAt,
		n.UpdatedAt,
		).Scan(&n.ID)


	if err != nil {
		log.Printf("NotificationEmail.Create() QueryRow() -> %v Error -> %v", n.queryCreate(), err)
		return
	}


	log.Println("NotificationEmail.Created() ID -> ", n.ID)

	return
}
