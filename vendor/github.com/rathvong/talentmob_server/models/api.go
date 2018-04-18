package models

import (

	"log"
	"time"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/rathvong/talentmob_server/system"
	"database/sql"
)

const (
	GOOGLE = "google"
	APPLE = "apple"
)

var (
	PushServices = []string {GOOGLE, APPLE}
)

//All tokens to handle connection to server
type Api struct {
	BaseModel
	UserID   uint64 `json:"user_id, omitempty"`
	DeviceID string `json:"device_id, omitempty"`
	PushNotificationToken string `json:"push_notification_token, omitempty"`
	PushNotificationService string `json:"push_notification_service, omitempty"`
	ManufacturerName string `json:"manufacturer_name, omitempty"`
	ManufacturerModel string `json:"manufacturer_model, omitempty"`
	ManufacturerVersion string `json:"manufacturer_version, omitempty"`
	Token    string `json:"token, omitempty"`
	IsActive bool   `json:"is_active, omitempty"`
}

//SQL query to create a row
func (a *Api) queryCreate() (qry string){
	return `INSERT INTO apis
					(user_id,
					token,
					push_notification_token,
					push_notification_service,
					manufacturer_name,
					manufacturer_model,
					manufacturer_version,
					device_id,
					is_active,
					created_at,
					updated_at)
			VALUES
					($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id`
}

func (a *Api) queryUpdate() (qry string){
	return `UPDATE apis SET
						user_id = $2,
						token = $3,
						push_notification_token = $4,
						push_notification_service = $5,
						manufacturer_name = $6,
						manufacturer_model = $7,
						manufacturer_version = $8,
						device_id = $9,
						is_active = $10,
						created_at = $11,
						updated_at = $12
			WHERE	id = $1`
}

//SQL query to retrieve a users api from token
func (a *Api) queryGetByAPIToken() (qry string){
	return `SELECT 	id,
					user_id,
					token,
					push_notification_token,
					push_notification_service,
					manufacturer_name,
					manufacturer_model,
					manufacturer_version,
					device_id,
					is_active,
					created_at,
					updated_at
			FROM apis
			WHERE token = $1`
}

func (a *Api) queryGetPushToken() (qry string){
	return `SELECT 	id,
					user_id,
					token,
					push_notification_token,
					push_notification_service,
					manufacturer_name,
					manufacturer_model,
					manufacturer_version,
					device_id,
					is_active,
					created_at,
					updated_at
			FROM apis
			WHERE push_notification_token = $1`
}


func (a *Api) queryActiveApis() (qry string){
	return `SELECT 	id,
					user_id,
					token,
					push_notification_token,
					push_notification_service,
					manufacturer_name,
					manufacturer_model,
					manufacturer_version,
					device_id,
					is_active,
					created_at,
					updated_at
			FROM apis
			WHERE user_id = $1
			AND is_active = true
			AND push_notification_token != ''`
}

func (a *Api) queryDisableByDeviceID() (qry string){
	return `UPDATE apis SET
				is_active = false
				WHERE device_id = $1`
}



//SQL query to validate if a token exists
func (a *Api) queryAPITokenIsValid() (qry string){
	return `SELECT EXISTS(SELECT 1 FROM apis WHERE token = $1 AND is_active = true)`
}

func (a *Api) queryPushTokenExists() (qry string){
	return `SELECT EXISTS(SELECT 1 FROM apis WHERE push_notification_token = $1 AND is_active = true)`
}

// validate important fields exists
func (a *Api) validateError() (err error){
	if a.UserID == 0 {
		return a.Errors(ErrorMissingValue, "user_id")
	}

	if a.Token == "" {
		return a.Errors(ErrorMissingValue, "token")
	}

	return
}

// Create a new row
func (a *Api) Create(db *system.DB) (err error){

	if err = a.validateError(); err != nil {
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
		log.Println("Api.Create() Error -> ", err)
		return
	}

	a.IsActive = true
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()


	err = tx.QueryRow(a.queryCreate(),
			a.UserID,
			a.Token,
			a.PushNotificationToken,
			a.PushNotificationService,
			a.ManufacturerName,
			a.ManufacturerModel,
			a.ManufacturerVersion,
			a.DeviceID,
			a.IsActive,
			a.CreatedAt,
			a.UpdatedAt).Scan(&a.ID)

	if err != nil {
		log.Printf("Api.Create() QueryRow() -> %v Error -> %v", a.queryCreate(), err)
		return
	}

	log.Println("Api.Create() create successful, id -> ", a.ID)
	return
}

// Update a row
func (a *Api) Update(db *system.DB) (err error){

	if err = a.validateError(); err != nil {
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
		log.Println("Api.Update() Error -> ", err)
		return
	}

	_, err = tx.Exec(a.queryUpdate(),
		a.ID,
		a.UserID,
		a.Token,
		a.PushNotificationToken,
		a.PushNotificationService,
		a.ManufacturerName,
		a.ManufacturerModel,
		a.ManufacturerVersion,
		a.DeviceID,
		a.IsActive,
		a.CreatedAt,
		a.UpdatedAt)

	if err != nil {
		log.Printf("Api.Update() ID -> %v QueryRow() -> %v Error -> %v",a.ID, a.queryUpdate(), err)
		return
	}

	log.Println("Api.Updated()  updated successfully, id -> ", a.ID)
	return
}

// Retrieve an api by token
func (a *Api) GetByAPIToken(db *system.DB, token string) (err error) {
	if token == "" {
		return a.Errors(ErrorMissingValue, "token - Api.GetByAPIToken")
	}

	err = db.QueryRow(a.queryGetByAPIToken(), token).Scan(
		&a.ID,
		&a.UserID,
		&a.Token,
		&a.PushNotificationToken,
		&a.PushNotificationService,
		&a.ManufacturerName,
		&a.ManufacturerModel,
		&a.ManufacturerVersion,
		&a.DeviceID,
		&a.IsActive,
		&a.CreatedAt,
		&a.UpdatedAt)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("Api.GetByAPIToken() Token -> %v QueryRow() -> %v Error -> %v", token, a.queryGetByAPIToken(), err)
		return
	}

	return
}

func (a *Api) IsPushServiceValid( ) (valid bool){

	for _, value := range PushServices {
		if value == a.PushNotificationService {
			return true
		}
	}

	return false
}


func (a *Api) GetPushNotificationToken(db *system.DB, token string) (err error) {
	if token == "" {
		return a.Errors(ErrorMissingValue, "token - Api.GetPushNotificationToken")
	}

	err = db.QueryRow(a.queryGetPushToken(), token).Scan(
		&a.ID,
		&a.UserID,
		&a.Token,
		&a.PushNotificationToken,
		&a.PushNotificationService,
		&a.ManufacturerName,
		&a.ManufacturerModel,
		&a.ManufacturerVersion,
		&a.DeviceID,
		&a.IsActive,
		&a.CreatedAt,
		&a.UpdatedAt)

	if err != nil && err != sql.ErrNoRows{
		log.Printf("Api.GetPushNotificationToken() Token -> %v QueryRow() -> %v Error -> %v", token, a.queryGetPushToken(), err)
		return
	}

	return
}

// Validate if token exists
func (a *Api) APITokenExists(db *system.DB, token string) (exists bool, err error){
	if token == "" {
		return false, a.Errors(ErrorMissingValue, "token - Api.APITokenExists")
	}

	err = db.QueryRow(a.queryAPITokenIsValid(), token).Scan(&exists)

	if err != nil {
		log.Printf("Api.APITokenExists() Token -> %v QueryRow() -> %v Error -> %v", token, a.queryAPITokenIsValid(), err)
		return
	}


	log.Println("APITokenExists() exists -> ", exists)
	return
}


// Check is push notification token exists
func (a *Api) PushTokenExists(db *system.DB, token string) (exists bool, err error){
	if token == "" {
		return false, a.Errors(ErrorMissingValue, "token - Api.PushTokenExists")
	}

	err = db.QueryRow(a.queryPushTokenExists(), token).Scan(&exists)

	if err != nil {
		log.Printf("Api.PushTokenExists() Token -> %v QueryRow() -> %v Error -> %v", token, a.queryPushTokenExists(), err)
		return
	}


	log.Println("PushTokenExists() exists -> ", exists)
	return
}

// Soft delete an api token
func (a *Api) Delete(db *system.DB) (err error){
	if a.ID == 0 {
		a.Errors(ErrorMissingID, "id")
		return
	}

	a.IsActive = false

	return a.Update(db)
}

func (a *Api) GetAllActiveAPIs(db *system.DB, userID uint64) (apis []Api, err error){

	if userID == 0 {
		return apis, a.Errors(ErrorMissingValue, "user_id")
	}

	rows, err := db.Query(a.queryActiveApis(), userID)

	defer rows.Close()

	if err != nil {
		log.Printf("Api.GetAllActiveAPIs() Query() -> %v Error -> %v", a.queryActiveApis(), err)
		return
	}

	return a.parseRows(rows)
}

func (a *Api) RemoveOLDAPIs(db *system.DB, deviceID string) (err error){

	if deviceID == "" {

		return a.Errors(ErrorMissingValue, "a.RemoveOLDAPIs() deviceID")
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
		log.Println("Api.RemoveOLDAPIS() Error -> ", err)
		return
	}

	_, err = tx.Exec(a.queryDisableByDeviceID(), deviceID)

	if err != nil {
		log.Printf("Api.RemoveOLDAPIs() deviceID -> %v Query -> %v Error -> %v", deviceID, a.queryDisableByDeviceID(), err)
		return
	}

	return
}

func (a *Api) parseRows(rows *sql.Rows) (apis []Api, err error){

	var count int
	for rows.Next() {
		api := Api{}

		err = rows.Scan(
			&api.ID,
			&api.UserID,
			&api.Token,
			&api.PushNotificationToken,
			&api.PushNotificationService,
			&api.ManufacturerName,
			&api.ManufacturerModel,
			&api.ManufacturerVersion,
			&api.DeviceID,
			&api.IsActive,
			&api.CreatedAt,
			&api.UpdatedAt,
		)

		if err != nil {
			log.Println("Api.parseRows() Scan() Error ->", err)
			return
		}

		apis = append(apis, api)
		count++
	}

	log.Println("Api.QueryActiveApis.ParseRows() apis -> ", count)
	return
}


// Generate a new SHA token
func (a *Api) GenerateAccessToken() {
	key := []byte("!TALENTMOB_2017_")
	h := hmac.New(sha256.New, key)
	t := time.Now()
	h.Write([]byte(t.String()))
	a.Token = base64.URLEncoding.EncodeToString(h.Sum(nil))
}