package models

import (
	"github.com/rathvong/talentmob_server/system"

	"log"
	"strings"
)

// This query model will only support video and user
// queries. If the query request does not match
// the data type supported. The server will send out
// an error response
type QueryType int

const (
	VIDEO = "video"
	USER  = "user"
)

const (
	QUERY_VIDEO QueryType = iota
	QUERY_USER
)

var queryTypes = []string{VIDEO, USER}

// Return query type to string
func (q *QueryType) String() (s string) {
	return queryTypes[*q]
}

//validate correct query type
func (q *Query) isValidTableSelected() (valid bool) {
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
	QueryType      QueryType
	Qry            string
	Categories     string //for video queries
	UserID         uint64
	WeeklyInterval int // depracated
}

// JSON REST response
// for a users query request
type QueryResult struct {
	ObjectType string      `json:"object_type"`
	Data       interface{} `json:"data"`
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
// If an empty query is sent, the response would be
// a list of the most recent uploaded and un voted items
func (q *Query) Find(db *system.DB, page int) (result QueryResult, err error) {

	if !q.isValidTableSelected() {
		err = q.Errors(ErrorIncorrectValue, "query_type")
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

		if len(q.Qry) > 0 || len(q.Categories) > 0 {
			result.Data, err = v.Find(db, q.Build(), page, q.UserID, q.WeeklyInterval)
			return
		}

		result.Data, err = v.GetDiscoveryTimeLine(db, q.UserID, page)
	}

	return
}

func (q *Query) Find2(db *system.DB, page int) (result QueryResult, err error) {

	if !q.isValidTableSelected() {
		err = q.Errors(ErrorIncorrectValue, "query_type")
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

		if len(q.Qry) > 0 || len(q.Categories) > 0 {
			result.Data, err = v.Find2(db, q.Build(), page, q.UserID, q.WeeklyInterval)
			return
		}

		result.Data, err = v.GetDiscoveryTimeLine2(db, q.UserID, page)
	}

	return
}

// Separate the string to build a format for the database so it can use to query data
func (q *Query) Build() (qry string) {
	var queryBuilder string

	if len(q.Categories) > 0 && len(q.Qry) > 0 {
		queryBuilder = q.buildQuery() + " | " + q.buildCategories()
	} else if len(q.Categories) > 0 && len(q.Qry) == 0 {
		queryBuilder = q.buildCategories()
	} else if len(q.Qry) > 0 && len(q.Categories) == 0 {
		queryBuilder = q.buildQuery()
	}

	return queryBuilder
}

// Format the categories string to be readable by Database
// the query will return any video that was tagged with the category name.
// It's important to separate each key word with ' | ' to notify the database
// to include these keywords in the ranking.
// The database will rank category keywords higher than the video titles
func (q *Query) buildCategories() (qry string) {
	var queryBuilder string

	queryBuilder = strings.TrimLeft(q.Categories, " ")
	queryBuilder = strings.TrimRight(queryBuilder, " ")

	array := strings.Split(queryBuilder, " ")

	for i, value := range array {
		if i == 0 {
			queryBuilder = value
		} else {
			queryBuilder += " | " + value
		}
	}

	return queryBuilder
}

// This will format the string to look for text in the title
// this query will be ranked higher if the titles include the query words.
// Separating each keyword with ' & ' will rank the videos that include those
// words higher in the rank results.
func (q *Query) buildQuery() (qry string) {
	var queryBuilder string

	queryBuilder = strings.TrimLeft(q.Qry, " ")
	queryBuilder = strings.TrimRight(queryBuilder, " ")

	array := strings.Split(queryBuilder, " ")

	for i, value := range array {
		if i == 0 {
			queryBuilder = value
		} else {
			queryBuilder += " & " + value
		}
	}

	return queryBuilder
}
