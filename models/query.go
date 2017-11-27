package models

import (
	"github.com/rathvong/talentmob_server/system"

	"log"
)

// Allow only certain query types for discovery screen
type QueryType int

const (
	VIDEO = "video"
	USER = "user"
)

const (
	QUERY_VIDEO QueryType =  iota
	QUERY_USER

)

var queryTypes  = []string {VIDEO, USER}

// Return query type to string
func (q *QueryType) String() (s string) {
	return queryTypes[*q]
}


//validate correct query type
func (q *Query) isValidTableSelected() (valid bool){
	for _, value := range queryTypes {

		if value == q.QueryType.String() {
			return true
		}
	}

	return false
}

//Handles all queries calls for specific objects
type Query struct {
	BaseModel
	QueryType QueryType
	Qry string
	Category string //for video queries
	UserID uint64
	WeeklyInterval int // depracated
}


type QueryResult struct {
	ObjectType string `json:"object_type"`
	Data interface{} `json:"data"`
}

// set query type
func (q *Query) SetQueryType(qt string) (err error) {
	switch qt {
	case USER:
		q.QueryType = QUERY_USER

	case VIDEO:
		q.QueryType = QUERY_VIDEO
	default:
		return q.Errors(ErrorIncorrectValue, "query_type")
	}

	return
}


// Perform query
func (q *Query) Find(db *system.DB, page int) (result QueryResult, err error){

	if !q.isValidTableSelected(){
		err = q.Errors(ErrorMissingValue, "query_type")
		log.Println("Query.Find() Error -> ", err)
		return
	}



	switch q.QueryType {

	case QUERY_USER:
		u := User{}
		result.ObjectType = USER
		result.Data, err = u.Find(db, q.Qry, page)

	case QUERY_VIDEO:
		v := Video{}
		result.ObjectType = VIDEO

		if len(q.Qry) > 0 {
			result.Data, err = v.Find(db, q.Qry, page, q.UserID, q.WeeklyInterval)
			return
		}

		result.Data, err = v.Recent(db, q.UserID, page, q.WeeklyInterval)
	}

	return
}



