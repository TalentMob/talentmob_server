package models

import (

	"log"
	"time"
	"golang.org/x/crypto/bcrypt"

	"github.com/rathvong/util"
	"github.com/rathvong/talentmob_server/system"

	"database/sql"

)

// The main struct for users account
// Users will be able to accumulate points, minutes watched
// and update there avatars. A new api token will be given to the
// user every time they sign up or re login
type User struct {
	BaseModel
	Api                  Api    `json:"api, omitempty"`
	Bio 				 Bio 	`json:"bio"`
	FacebookID           string `json:"facebook_id"`
	Avatar               string `json:"avatar"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	AccountType          int    `json:"account_type"`
	MinutesWatched       uint64 `json:"minutes_watched"`
	Points               uint64 `json:"points"`
	Password             string `json:"password, omitempty"`
	ImportedVideosCount  int    `json:"imported_videos_count"`
	FavouriteVideosCount int    `json:"favourite_videos_count"`
	EncryptedPassword    string `json:"-"`
}

type ProfileUser struct {
	ID                   uint64 `json:"id"`
	Bio                  Bio    `json:"bio"`
	Name                 string `json:"name"`
	Avatar               string `json:"avatar"`
	ImportedVideosCount  int    `json:"imported_videos_count"`
	FavouriteVideosCount int    `json:"favourite_videos_count"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}


func (p *ProfileUser) GetUser(db *system.DB, userID uint64) (err error){
	user := User{}

	if err = user.Get(db, userID); err != nil {
		return
	}

	if err = user.Bio.Get(db, userID); err != nil {
		return
	}

	p.ID = user.ID
	p.Name = user.Name
	p.Avatar = user.Avatar
	p.FavouriteVideosCount = user.FavouriteVideosCount
	p.ImportedVideosCount = user.ImportedVideosCount
	p.Bio = user.Bio
	p.CreatedAt = user.CreatedAt
	p.UpdatedAt = user.UpdatedAt

	return
}



// SQL query to create a row in users table
func (u *User) queryCreate() (qry string){
	return `INSERT INTO users
						(facebook_id,
			 			avatar,
						name,
						email,
						account_type,
						minutes_watched,
						points,
						created_at,
						updated_at,
						encrypted_password,
						favourite_videos_count,
						imported_videos_count)
			VALUES
						($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id`
}

// SQL query to update a row in users table
func (u *User) queryUpdate() (qry string){
	return `UPDATE 	users
			SET
					facebook_id = $2,
					avatar = $3,
					name = $4,
					email = $5,
					account_type = $6,
					minutes_watched = $7,
					points = $8,
					updated_at = $9,
					encrypted_password = $10,
					favourite_videos_count = $11,
					imported_videos_count = $12
			WHERE	id = $1`
}

// SQL query to retrieve a user by email
func (u *User) queryGetByEmail() (qry string){
	return `SELECT  id,
				    facebook_id,
				    avatar,
				    name,
				    email,
					account_type,
					minutes_watched,
					points,
					created_at,
					updated_at,
					encrypted_password,
					favourite_videos_count,
					imported_videos_count
			FROM
					users
			WHERE	email = $1`
}

// SQL query to retrieve a user by email
func (u *User) queryGetByID() (qry string){
	return `SELECT  id,
				    facebook_id,
				    avatar,
				    name,
				    email,
					account_type,
					minutes_watched,
					points,
					created_at,
					updated_at,
					encrypted_password,
					favourite_videos_count,
					imported_videos_count
			FROM
					users
			WHERE	id = $1`
}


// SQL query to retrieve a user by facebook_id
func (u *User) queryGetByFacebookID() (qry string){
	return `SELECT  id,
				    facebook_id,
				    avatar,
				    name,
				    email,
					account_type,
					minutes_watched,
					points,
					created_at,
					updated_at,
					encrypted_password,
					favourite_videos_count,
					imported_videos_count
			FROM
					users
			WHERE	facebook_id = $1`
}

func (u *User) queryGetByName() (qry string){
	return `SELECT
					id,
        			facebook_id,
					avatar,
					name,
					email,
					account_type,
					minutes_watched,
					points,
					created_at,
					updated_at,
					encrypted_password,
					favourite_videos_count,
					imported_videos_count
				FROM users
				WHERE
					name ILIKE $1
				LIMIT $2
				OFFSET $3`


}

// SQL query to validate if a row exists with email
func (u *User) queryEmailExists() (qry string){
	return `SELECT EXISTS(SELECT 1 FROM USERS WHERE email = $1)`
}

// SQL query to validate if facebook_id exists
func (u *User) queryFacebookIDExists() (qry string){
	return `SELECT EXISTS(SELECT 1 FROM USERS WHERE facebook_id = $1)`
}

// SQL query to validate if id exists
func (u *User) queryIDExists() (qry string){
	return `SELECT EXISTS(SELECT 1 FROM USERS WHERE id = $1)`
}


// SQL query to retrieve followers - incomplete
func (u *User) queryGetFollowers() (qry string){
	return ``
}

// SQL query to retrieve who the user is following - incomplete
func (u *User) queryGetFollowing() (qry string){
	return ``
}

// Validate and ensure important columns have value
func (u *User) validateError() (err error){
	if u.Name  == ""{
		return u.Errors(ErrorMissingValue, "name")
	}

	if u.Email == "" {
		return u.Errors(ErrorMissingValue, "email")
	}

	if u.EncryptedPassword == "" {
		return u.Errors(ErrorMissingValue, "encrypted_password")
	}

	return nil
}

func (u *User) GeneratePassword(){
	u.Password = util.RandStringBytesMaskImprSrc(10)
	u.EncryptPassword()
}


// Create a users
func (u *User) Create(db *system.DB)(err error) {

	if err = u.validateError(); err != nil {
		return err
	}

	if exists, err := u.EmailExists(db, u.Email); exists || err != nil {
		if err == nil {
			return u.Errors(ErrorExists, "email")
		}
		return err
	}

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

	}()

	if err != nil {
		log.Println("User.Create() Begin -> ", err)
		return
	}

	//initialize date values
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	err = tx.QueryRow(u.queryCreate(),
		u.FacebookID,
		u.Avatar,
		u.Name,
		u.Email,
		u.AccountType,
		u.MinutesWatched,
		u.Points,
		u.CreatedAt,
		u.UpdatedAt,
		u.EncryptedPassword,
		u.ImportedVideosCount,
		u.FavouriteVideosCount).Scan(&u.ID)

	if err != nil {
		log.Printf("User.Create() QueryRow() -> %v Error -> %v", u.queryCreate(), err)
		return
	}


	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return
	}

	u.Api.UserID = u.ID
	err = u.Api.Create(db)

	if err != nil {
		return
	}


	log.Println("User.Create() user created -> ", u.ID)
	return
}

// Update a user
func (u *User) Update(db *system.DB) (err error) {

	if err = u.validateError(); err != nil {
		return err
	}

	if u.ID == 0 {
		return u.Errors(ErrorMissingValue, "id")
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
		log.Println("User.Update() Begin -> ", err)
		return
	}

	_, err = tx.Exec(u.queryUpdate(),
		u.ID,
		u.FacebookID,
		u.Avatar,
		u.Name,
		u.Email,
		u.AccountType,
		u.MinutesWatched,
		u.Points,
		u.UpdatedAt,
		u.EncryptedPassword,
		u.ImportedVideosCount,
		u.FavouriteVideosCount)

	if err != nil {
		log.Printf("User.Update() Exec() -> %v Error -> %v", u.queryUpdate(), err)
		return
	}


	log.Println("User.Update() Update complete.")
	return
}

// Check if a user exists
func (u *User) EmailExists(db *system.DB, email string) (exists bool, err error){

	if email == "" {
		return false, u.Errors(ErrorMissingValue, "email")
	}

	err = db.QueryRow(u.queryEmailExists(), email).Scan(&exists)

	if err != nil {
		log.Printf("User.EmailExists() Email -> %v QueryRow() -> %v error -> %v", email, u.queryEmailExists(), err)
		return
	}

	log.Println("User.EmailExists() email exists -> ", exists)

	return
}

// Check if a user exists
func (u *User) IDExists(db *system.DB, id uint64) (exists bool, err error){

	if id == 0{
		return false, u.Errors(ErrorMissingValue, "id")
	}

	err = db.QueryRow(u.queryIDExists(), id).Scan(&exists)

	if err != nil {
		log.Printf("User.IDExists() id -> %v QueryRow() -> %v error -> %v", id, u.queryIDExists(), err)
		return
	}

	log.Println("User.IDExists() email exists -> ", exists)

	return
}


// Check if a user exists by facebook_id
func (u *User) FacebookIDExists(db *system.DB, facebookID string) (exists bool, err error){

	if facebookID == "" {
		return false, u.Errors(ErrorMissingValue, "facebookID")
	}

	err = db.QueryRow(u.queryFacebookIDExists(), facebookID).Scan(&exists)

	if err != nil {
		log.Printf("User.FacebookIDExists() facebookID -> %v QueryRow() -> %v error -> %v", facebookID, u.queryFacebookIDExists(), err)
		return
	}

	log.Println("User.FacebookIDExists() facebookID exists -> ", exists)

	return
}


// Get user from database by email
func (u *User) GetByEmail(db *system.DB, email string) (err error) {

	if email == "" {
		return u.Errors(ErrorMissingValue, "email")
	}

	err = db.QueryRow(u.queryGetByEmail(), email).Scan(&u.ID,
		&u.FacebookID,
		&u.Avatar,
		&u.Name,
		&u.Email,
		&u.AccountType,
		&u.MinutesWatched,
		&u.Points,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.EncryptedPassword,
		&u.FavouriteVideosCount,
		&u.ImportedVideosCount)

	if err != nil {
		log.Printf("User.Get() Email -> %v QueryRow() -> %v Error -> %v", email, u.queryGetByEmail(), err)
		return
	}

	return
}

// Get user from database by id
func (u *User) Get(db *system.DB, id uint64) (err error) {

	if id == 0 {
		return u.Errors(ErrorMissingValue, "email")
	}

	err = db.QueryRow(u.queryGetByID(), id).Scan(&u.ID,
		&u.FacebookID,
		&u.Avatar,
		&u.Name,
		&u.Email,
		&u.AccountType,
		&u.MinutesWatched,
		&u.Points,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.EncryptedPassword,
		&u.FavouriteVideosCount,
		&u.ImportedVideosCount)

	if err != nil {
		log.Printf("User.Get() id -> %v QueryRow() -> %v Error -> %v", id, u.queryGetByID(), err)
		return
	}

	return
}


// Get user from by facebook id
func (u *User) GetByFacebookID(db *system.DB, id string) (err error){

	if id == "" {
		return u.Errors(ErrorMissingValue, "id")
	}

	err = db.QueryRow(u.queryGetByFacebookID(), id).Scan(&u.ID,
		&u.FacebookID,
		&u.Avatar,
		&u.Name,
		&u.Email,
		&u.AccountType,
		&u.MinutesWatched,
		&u.Points,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.EncryptedPassword,
		&u.FavouriteVideosCount,
		&u.ImportedVideosCount)

	if err != nil {
		log.Printf("User.GetByFacebookID() id -> %v QueryRow() -> %v Error -> %v", id, u.queryGetByFacebookID(), err)
		return
	}


	return
}

// Encrypt user password for the database
func (u *User) EncryptPassword() error {

	if u.Password != "" {
		if pw, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10); err == nil {
			u.Password = ""

			if err != nil {
				log.Println(err)
				return err
			}
			u.EncryptedPassword = string(pw)
			return nil
		}
	}


	return u.Errors(ErrorMissingValue, "password")
}

// Decrypt password
func (u *User) DecryptHashPassword() (validated bool) {

	if bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(u.Password)) != nil {
		return false
	}

	return true

}


func (u *User) Find(db *system.DB,qry string,  page int) (users []User, err error){

	name := "%" + qry + "%"

	rows, err := db.Query(u.queryGetByName(), name, LimitQueryPerRequest, offSet(page))

	defer rows.Close()

	if err != nil {
		log.Printf("User.Find() name -> %v  \nQuery() -> %v \nError -> %v", name, u.queryGetByName(), err)
		return
	}

	return u.parseRows(rows)
}



func (u *User) parseRows(rows *sql.Rows) (users []User, err error){

	for rows.Next() {
		user := User{}

		err = rows.Scan(&user.ID,
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
			&user.ImportedVideosCount)


		if err != nil {
			log.Println("User.parseRows() Error -> ", err)
			return
		}

		users = append(users, user)
	}

	return
}
