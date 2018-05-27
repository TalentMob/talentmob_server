package models

import (
	"errors"
	"fmt"
	"time"
)

type BaseErrors int

// The max number of queries returned
// Change the limit to retrieve more from each query
const (
	LimitQueryPerRequest = 10
)

// Error code list for models
const (
	ErrorMissingID BaseErrors = iota
	ErrorMissingValue
	ErrorIncorrectValue
	ErrorExists
	ErrorUserDoesNotExist
)

// Base model for each structure
type BaseModel struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// create a new error message for models
// code BaseErrors
// column string
func (b *BaseModel) Errors(code BaseErrors, key string) (err error) {

	switch code {
	case ErrorMissingID:
		return errors.New("missing id")
	case ErrorMissingValue:
		return fmt.Errorf("missing value for %v", key)
	case ErrorIncorrectValue:
		return fmt.Errorf("incorrect value for %v", key)
	case ErrorExists:
		return fmt.Errorf("value already exists for %v", key)
	case ErrorUserDoesNotExist:
		return fmt.Errorf("user does not exist for %v", key)
	default:
		return errors.New("unknown error")
	}

}

// calculate offset for each page from queries
func OffSet(page int) (offset int) {
	page--

	if page < 0 {
		page = 0
	}

	return page * LimitQueryPerRequest
}

func OffSetWithLimit(page int, limit int) (offset int) {
	page--

	if page < 0 {
		page = 0
	}

	return page * limit
}
