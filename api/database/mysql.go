package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	// Load mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// SetupDatabase connects to the database using environment variables
func SetupDatabase() *sql.DB {

	host := os.Getenv("MYSQL_HOST")
	database := os.Getenv("MYSQL_DATABASE")
	username := os.Getenv("MYSQL_USERNAME")
	password := os.Getenv("MYSQL_PASSWORD")

	if host == "" || database == "" || username == "" {
		log.Fatalln("Database host, name and username are required")
	}

	address := fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, host, database)
	log.Printf("Connecting to database: %s", address)

	db, err := sql.Open("mysql", address)
	if err != nil {
		log.Fatalf("Could not connect to database: %s\n", err.Error())
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not connect to database: %s\n", err.Error())
	}

	return db
}
