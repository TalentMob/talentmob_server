package system

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

// DB struct will be a global variable in main to handle all db calls
type DB struct {
	*sql.DB
}


// connect to database with a url
// url string - location of the database
func Connect(url string) *DB {

	db, _ := sql.Open("postgres", url)
	err := db.Ping()
	if err != nil {
		log.Println(err)
		panic(err)
	}



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