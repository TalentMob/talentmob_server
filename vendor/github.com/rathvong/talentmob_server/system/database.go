package system

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"sync"
	"time"
	"os"
	"github.com/rathvong/talentmob_server/system"
)

// DB struct will be a global variable in main to handle all db calls
type DB struct {
	*sql.DB
}

var once sync.Once
// Key strings for environment variables
const (
	AWS_ENVIRONMENT_DATABASE_URL    = "DATABASE_AWS"
	HEROKU_ENVIRONMENT_DATABASE_URL = "DATABASE_URL"
)


// Initialized database url set in environment
var (
	//AWS DB URL
	awsDatabaseURL = os.Getenv(AWS_ENVIRONMENT_DATABASE_URL)
	// Heroku DB URL
	herokuDatabaseUrl = os.Getenv(HEROKU_ENVIRONMENT_DATABASE_URL)
)


// find and set database for server
func setDatabaseUrl() (url string) {
	if herokuDatabaseUrl != "" {
		return herokuDatabaseUrl
	} else if awsDatabaseURL != "" {
		return awsDatabaseURL
	}

	return
}

func Database() *DB {
	return system.Connect(awsDatabaseURL + "&sslmode=verify-full&sslrootcert=config/rds-combined-ca-bundle.pem")
}


// connect to database with a url
// url string - location of the database
func Connect(url string) *DB {

	var db *sql.DB

	once.Do(func() {
		db, _ = sql.Open("postgres", url)

		db.SetMaxOpenConns(87) // Sane default
		db.SetMaxIdleConns(0)
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