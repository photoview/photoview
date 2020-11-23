package database

import (
	"log"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/viktorstrate/photoview/api/graphql/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// SetupDatabase connects to the database using environment variables
func SetupDatabase() (*gorm.DB, error) {

	address, err := url.Parse(os.Getenv("MYSQL_URL"))
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse mysql url")
	}

	if address.String() == "" {
		return nil, errors.New("Environment variable MYSQL_URL missing, exiting")
	}

	queryValues := address.Query()
	queryValues.Add("multiStatements", "true")
	queryValues.Add("parseTime", "true")

	address.RawQuery = queryValues.Encode()

	log.Printf("Connecting to database: %s", address)

	db, err := gorm.Open(mysql.Open(address.String()), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to database")
	}

	// var db *sql.DB

	// db, err = sql.Open("mysql", address.String())
	// if err != nil {
	// 	return nil, errors.New("Could not connect to database, exiting")
	// }

	// tryCount := 0

	// for {
	// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()

	// 	if err := db.PingContext(ctx); err != nil {
	// 		if tryCount < 4 {
	// 			tryCount++
	// 			log.Printf("WARN: Could not ping database: %s, Will retry after 1 second", err)
	// 			time.Sleep(time.Second)
	// 			continue
	// 		} else {
	// 			return nil, errors.Wrap(err, "Could not ping database, exiting")
	// 		}
	// 	}

	// 	break
	// }

	// db.SetMaxOpenConns(80)

	return db, nil
}

func MigrateDatabase(db *gorm.DB) error {

	db.AutoMigrate(
		&models.User{},
		&models.AccessToken{},
		&models.SiteInfo{},
		&models.Media{},
		&models.MediaURL{},
		&models.Album{},
		&models.MediaEXIF{},
	)

	return nil
}
