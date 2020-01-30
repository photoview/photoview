package database

import (
	"database/sql"
	"log"
	"os"

	// Load mysql driver
	// _ "github.com/go-sql-driver/mysql"

	// Load postgres driver
	_ "github.com/lib/pq"
)

// SetupDatabase connects to the database using environment variables
func SetupDatabase() *sql.DB {

	address := os.Getenv("POSTGRES_URL")
	log.Printf("Connecting to database: %s", address)

	db, err := sql.Open("postgres", address)
	if err != nil {
		log.Fatalf("Could not connect to database: %s\n", err.Error())
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not connect to database: %s\n", err.Error())
	}

	return db
}
