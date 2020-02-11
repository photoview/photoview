package database

import (
	"database/sql"
	"log"
	"net/url"
	"os"

	// Load mysql driver
	_ "github.com/go-sql-driver/mysql"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	// Migrate from file
	_ "github.com/golang-migrate/migrate/source/file"
)

// SetupDatabase connects to the database using environment variables
func SetupDatabase() *sql.DB {

	address, err := url.Parse(os.Getenv("MYSQL_URL"))
	if err != nil {
		log.Fatalf("Could not parse mysql url: %s\n", err)
	}

	queryValues := address.Query()
	queryValues.Add("multiStatements", "true")
	queryValues.Add("parseTime", "true")

	address.RawQuery = queryValues.Encode()

	log.Printf("Connecting to database: %s", address)

	db, err := sql.Open("mysql", address.String())
	if err != nil {
		log.Fatalf("Could not connect to database: %s\n", err.Error())
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not connect to database: %s\n", err.Error())
	}

	return db
}

func MigrateDatabase(db *sql.DB) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"mysql",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			log.Println("Database is up to date")
		} else {
			return err
		}
	} else {
		log.Println("Database migrated")
	}

	return nil
}
