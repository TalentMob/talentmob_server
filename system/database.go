package system

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// DB struct will be a global variable in main to handle all db calls
type DB struct {
	*sql.DB
}

var once sync.Once

// connect to database with a url
// url string - location of the database
func Connect(url string) *DB {

	var db *sql.DB

	once.Do(func() {
		db, _ = sql.Open("postgres", url)

		db.SetMaxOpenConns(87) // Sane default
		db.SetMaxIdleConns(20)
		db.SetConnMaxLifetime(time.Nanosecond)

		err := db.Ping()
		if err != nil {
			log.Println(err)
			panic(err)
		}

	})

	return &DB{db}
}

// Ping database for connection
func (db *DB) PingConnectionToDatabase() {
	log.Println("PingConnectionToDatabase()")
	err := db.Ping()

	if err != nil {
		panic(err)
	}
}
