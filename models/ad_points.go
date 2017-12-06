package models

import (
	"github.com/rathvong/talentmob_server/system"
	"time"
	"log"

	"github.com/jinzhu/now"
)

type AdPoint struct {
	BaseModel
	ID uint64 `json:"id"`
	UserID uint64 `json:"user_id"`
	IsActive bool `json:"is_active"`
}


func (a *AdPoint) queryCreate() (qry string){
	return `INSERT INTO ad_points
						(user_id,
						is_active,
						created_at,
						updated_at)
					VALUES
						($1, $2, $3, $4)
				RETURNING id`
}


func (a *AdPoint) queryUpdate() (qry string){
	return `UPDATE ad_points SET
						user_id = $2,
						is_active = $3,
						created_at = $4,
						updated_at = $5
				WHERE id = $1`
}

func (a *AdPoint) queryCountByDate() (qry string){
	return `SELECT
					count(*)
				FROM ad_points
				WHERE user_id = $1
				AND created_at > $1`
}


func (a *AdPoint) validateCreateErrors() (err error){


	if a.UserID == 0 {
		return a.Errors(ErrorMissingValue, "id")
	}


	return
}

func (a *AdPoint) validateAdCountPerDay(db *system.DB) ( err error){

	loc, _ := time.LoadLocation("EST")

	n := now.BeginningOfDay().In(loc)
	var count int

	if count, err = a.CountByDate(db, a.UserID, n); err != nil {
		return err
	}

	if count > 20 {
		return a.Errors(ErrorExists, "limit is 20 video ads per day")
	}
	return
}


func (a *AdPoint) validateUpdateErrors() (err error){


	if a.ID == 0 {
		return
	}

	return a.validateCreateErrors()
}

func (a *AdPoint) UpdatePoints(db *system.DB) (err error){
	p := Point{}

	if err := p.GetByUserID(db, a.UserID); err != nil {
		panic(err)
	}


	p.AddPoints(POINT_ACTIVITY_AD_WATCHED)


	return p.Update(db)
}

func (a *AdPoint) Create(db *system.DB) (err error){

	if err = a.validateCreateErrors(); err != nil {

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

		a.UpdatePoints(db)
	}()

	if err != nil {
		log.Println("AdPoint.Create() Begin() Error -> ", err)
		return
	}

	a.IsActive = true
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	if err = a.validateAdCountPerDay(db); err != nil {
		return
	}

	err = tx.QueryRow(
		a.queryCreate(),
		a.UserID,
		a.IsActive,
		a.CreatedAt,
		a.UpdatedAt,
		).Scan(&a.ID)


	if err != nil {
		log.Printf("AdPoint.Create() QueryRow() -> %v Error -> %v", a.queryCreate(), err)
		return
	}

	return

}


func (a *AdPoint) Update(db *system.DB) (err error) {

	if err = a.validateUpdateErrors(); err != nil {
		log.Println("AdPoint.Update() Error -> ", err)
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
		log.Println("AdPoint.Update() Begin() Error -> ", err)
		return
	}


	a.UpdatedAt = time.Now()


	_,err = tx.Exec(
		a.queryUpdate(),
		a.ID,
		a.UserID,
		a.IsActive,
		a.CreatedAt,
		a.UpdatedAt,
	)


	if err != nil {
		log.Printf("AdPoint.Update() id -> %v QueryRow() -> %v Error -> %v", a.ID, a.queryUpdate(), err)
		return
	}



	return
}

func (a *AdPoint) CountByDate(db *system.DB, userID uint64, date time.Time)( count int, err error){

	if date.String() == ""{
		err = a.Errors(ErrorMissingValue, "date")
		return
	}

	err = db.QueryRow(a.queryCountByDate(), userID, date).Scan(&count)

	if err != nil {
		log.Printf("AdPoint.CountByDate() userID -> %v QueryRow() -> %v Error -> %v", userID, a.queryCountByDate(), err)
		return
	}

	return
}
