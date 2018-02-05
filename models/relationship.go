package models

import (
	"github.com/rathvong/talentmob_server/system"
	"time"
	"log"
	"database/sql"
)

/**
	Will handle all connections between users representing followers and
	followings relationship. Users will be able to follow and unfollow
	other users and retrieve followers and followings list for each profile.
 */
type Relationship struct {
	BaseModel
	FollowedID uint64 `json:"followed_id"`
	FollowerID uint64 `json:"follower_id"`
	RelationShipType string `json:"relationship_type"`
	IsActive bool `json:"is_active"`
}

/**
	All relationships will have a status according to the
	current relation to each follower or following.
	Users can block, request or accept a relationship
	in the future as a friend's request
 */
type RelationTypes struct {
	Request string
	Block string
	Accepted string
}

/**
	RelationShipType values which can only be supported by the server.

 */
var RelationShipType = RelationTypes{
	Request: "request",
	Block:"block",
	Accepted:"accepted",
}

/**
	Will create a new relationship between users. A user can only follow or be followed once
 */
func (r *Relationship) queryCreate() (qry string){
	return `INSERT INTO relationships 
					(followed_id, follower_id, relationship_type, is_active, created_at, updated_at)
				 VALUES
					($1, $2, $3, $4, $5, $6)
				 RETURNING id`
}

/**
	Update the relationship between users
 */
func (r *Relationship) queryUpdate() (qry string) {
	return `UPDATE relationships SET
					followed_id = $2,
					follower_id = $3,
					relationship_type = $4,
					is_active = $5,
					created_at = $6,
					updated_at = $7
				 WHERE id = $1`
}


/**
	Validate if a user's relationship already exists
 */
func (r *Relationship) queryExists() (qry string) {
	return `SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = $1 AND follower_id = $2)`
}

/**
	Query all followers for the selected User
 */
func (r *Relationship) queryFollowers() (qry string) {
	return `SELECT users.id,
					users.facebook_id,
					users.avatar,
					users.name,
					users.email,
					users.account_type,
					users.minutes_watched,
					users.points,
					users.created_at,
					users.updated_at,
					users.encrypted_password,
					users.favourite_videos_count,
					users.imported_videos_count
				FROM relationships
				INNER JOIN users
				ON users.id = relationships.follower_id
				WHERE relationships.followed_id = $1
				LIMIT $2
				OFFSET $3`
}


/**
	Query all following for selected User
 */
func (r *Relationship) queryFollowing() (qry string) {
	return `SELECT users.id,
					users.facebook_id,
					users.avatar,
					users.name,
					users.email,
					users.account_type,
					users.minutes_watched,
					users.points,
					users.created_at,
					users.updated_at,
					users.encrypted_password,
					users.favourite_videos_count,
					users.imported_videos_count
				FROM relationships
				INNER JOIN users
				ON users.id = relationships.followed_id
				WHERE relationships.follower_id = $1
				LIMIT $2
				OFFSET $3`
}

/**
	Query and retrieve a specific relation by ID
 */
func (r *Relationship) queryGet() (qry string) {
	return `SELECT id,
						followed_id,
						follower_id,
						relationship_type,
						is_active,
						created_at,
						updated_at
				FROM relationships
				WHERE followed_id = $1
				AND follower_id = $2
				LIMIT 1
				`
}

/**
	Query if a user is following another user
 */
func (r *Relationship) queryIsFollowing() (qry string) {
	return `SELECT EXISTS(SELECT 1 FROM relationships WHERE followed_id = $1 AND follower_id = $2 AND is_active = true)`
}



/**
	We must ensure the values needed for a relationship to be created
	is present. If it is not it will cause an error
 */
func (r *Relationship) validateCreate() (err error){
	if r.FollowedID == 0 {
		return r.Errors(ErrorMissingValue, "followed_id")
	}

	if r.FollowerID == 0 {
		return r.Errors(ErrorMissingValue, "follower_id")
	}

	switch r.RelationShipType {
	case RelationShipType.Accepted, RelationShipType.Block, RelationShipType.Request:
	default:
		return r.Errors(ErrorMissingValue, "relationship_type")

	}

	return
}


/**
	Validate if proper values are needed to update relationships
 */
func (r *Relationship) validateUpdate() (err error) {

	if r.ID == 0 {
		return r.Errors(ErrorMissingValue, "id")
	}

	return r.validateCreate()
}

/**
	When creating a new relationship, we have to ensure that the relationship is unique. If a relationship is created
	we can reactivate that relationship without have to notify the target user
 */
func (r *Relationship) New(db *system.DB, followedID uint64, followerID uint64) (err error){

	r.FollowerID = followerID
	r.FollowedID = followedID
	r.RelationShipType = RelationShipType.Accepted


	if exist, err := r.Exists(db, followedID, followerID); exist || err != nil {

		if err != nil {
			return err
		}

		if err = r.Get(db, followedID, followerID); err != nil {
			return err
		}


		r.IsActive = true
		return r.Update(db)
	}

	return r.create(db)
}

/**
	Create a new relationship between users
 */
func (r *Relationship) create(db *system.DB) (err error) {

	if err = r.validateCreate(); err != nil {
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


		Notify(db, r.FollowerID, r.FollowedID, VERB_FOLLOWED, r.FollowerID, OBJECT_USER)
	}()

	if err != nil {
		log.Println("Relationship.Create() Error -> ", err)
		return
	}

	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()

	err = tx.QueryRow(
		r.queryCreate(),
		r.FollowedID,
		r.FollowerID,
		r.RelationShipType,
		r.IsActive,
		r.CreatedAt,
		r.UpdatedAt).Scan(&r.ID)

	if err != nil {
		log.Printf("Relationship.Create() Follower_id -> %v Followed_id -> %v RelationshipType -> %v QueryRow() -> %v Error -> %v", r.FollowerID, r.FollowedID, r.RelationShipType, r.queryCreate(), err)
		return
	}

	return
}


/**
	update a relationship between users
 */
func (r *Relationship) Update(db *system.DB) (err error) {

	if err = r.validateUpdate(); err != nil {
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
		log.Println("Relationship.Update() Error -> ", err)
		return
	}

	_, err = tx.Exec(
		r.queryUpdate(),
		r.ID,
		r.FollowedID,
		r.FollowerID,
		r.RelationShipType,
		r.IsActive,
		r.CreatedAt,
		r.UpdatedAt,)

	if err != nil {
		log.Printf("Relationship.Update() Follower_id -> %v Followed_id -> %v RelationshipType -> %v QueryRow() -> %v Error -> %v", r.FollowerID, r.FollowedID, r.RelationShipType, r.queryUpdate(), err)
		return
	}

	return
}

/**
	We want to validate if a user relation already exists before creating a new relationship incase a user has unfollowed
	their target user. We can retrieve that relationship and declare it active.
 */
func (r *Relationship) Exists(db *system.DB, followedID uint64, followerID uint64) (exists bool, err error){
	if followedID == 0 || followerID == 0 {
		return false, r.Errors(ErrorMissingValue, "followed_id -> " + string(followedID) + "follower_id -> " + string(followerID))
	}

	err = db.QueryRow(r.queryExists(), followedID, followerID).Scan(&exists)

	if err != nil {
		log.Printf("Relationship.Exists() followed_id -> %v follower_id -> %v QueryRow() -> %v Err -> %v", r.FollowedID, r.FollowerID, r.queryExists(), err)
		return
	}


	return
}

/**
	Returns true if a user is following another user
 */
func (r *Relationship) IsFollowing(db *system.DB, followedID uint64, followerID uint64) (exists bool, err error){
	if followedID == 0 || followerID == 0 {
		return false, r.Errors(ErrorMissingValue, "followed_id -> " + string(followedID) + "follower_id -> " + string(followerID))
	}

	err = db.QueryRow(r.queryIsFollowing(), followedID, followerID).Scan(&exists)

	if err != nil {
		log.Printf("Relationship.IsFollowing() followed_id -> %v follower_id -> %v QueryRow() -> %v Err -> %v", r.FollowedID, r.FollowerID, r.queryIsFollowing(), err)
		return
	}


	return
}

/**
	Retrieve the relationship between a follower and followed
 */
func (r *Relationship) Get(db *system.DB, followedID uint64, followerID uint64 ) (err error) {
	if followedID == 0 || followerID == 0 {
		return  r.Errors(ErrorMissingValue, "followed_id -> " + string(followedID) + "follower_id -> " + string(followerID))
	}

	err = db.QueryRow(r.queryGet(), followedID, followerID).Scan(
		&r.ID,
		&r.FollowedID,
		&r.FollowerID,
		&r.RelationShipType,
		&r.IsActive,
		&r.CreatedAt,
		&r.UpdatedAt,)

	if err != nil {
		log.Printf("Relationship.Get() followed_id -> %v follower_id -> %v QueryRow() -> %v Err -> %v", r.FollowedID, r.FollowerID, r.queryGet(), err)
		return
	}

	return
}

/**
	retrieve all the users followed by the selected user
 */
func (r *Relationship) GetFollowing(db *system.DB, userID uint64, page int) (users []User, err error) {
	if userID == 0 {
		return users, r.Errors(ErrorMissingValue, "user_id")
	}

	rows, err := db.Query(r.queryFollowing(), userID, LimitQueryPerRequest, offSet(page))

	defer rows.Close()


	if err != nil {
		log.Printf("Relationship.GetFollowing() UserID -> %v Query() -> %v Error -> %v", userID, r.queryFollowing(), err)
		return
	}

	return r.ParseRows(db, rows)
}

/**
	retrieve all the users following the selected user
 */
func (r *Relationship) GetFollowers(db *system.DB, userID uint64, page int) (users []User, err error) {

	if userID == 0 {
		return users, r.Errors(ErrorMissingValue, "user_id")
	}

	rows, err := db.Query(r.queryFollowers(), userID, LimitQueryPerRequest, offSet(page))

	defer rows.Close()


	if err != nil {
		log.Printf("Relationship.GetFollowers() UserID -> %v Query() -> %v Error -> %v", userID, r.queryFollowers(), err)
		return
	}


	return r.ParseRows(db, rows)
}

/**
	Parse data rows retrieve by followers and following query
 */
func (r *Relationship) ParseRows(db *system.DB, rows *sql.Rows) (users []User, err error) {

	for rows.Next() {
		user := User{}

		err = rows.Scan(
			&user.ID,
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
			&user.ImportedVideosCount,
			&user.TotalVotesReceived,)


		if err != nil {
			log.Println("Relationship.ParseRows()", err)
			return
		}

		users = append(users, user)
	}

	return
}


/**
	Will populate a user list with following data
 */
func (r *Relationship) PopulateFollowingData(db *system.DB, userID uint64, users []User) (result []User, err error) {
	if userID == 0 {

		return result, r.Errors(ErrorMissingValue, "userID")
	}

	if len(users) == 0 {
		return result, r.Errors(ErrorMissingValue, "users")
	}

	for _, user := range users {


		user.IsFollowing, err = r.IsFollowing(db, user.ID, userID)

		if err != nil {
			return result, err
		}


		result = append(result, user)

	}

	return
}