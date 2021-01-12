package database

import (
	"context"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	config := gorm.Config{}

	// Enable database debug logging
	config.Logger = logger.Default.LogMode(logger.Info)

	db, err := gorm.Open(mysql.Open(address.String()), &config)
	sqlDB, dbErr := db.DB()
	if dbErr != nil {
		log.Println(dbErr)
		return nil, dbErr
	}

	sqlDB.SetMaxOpenConns(80)

	if err != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for retryCount := 1; retryCount <= 5; retryCount++ {
			select {
			case <-ctx.Done():
				log.Println(ctx.Err())
				return nil, err
			default:
				if err := sqlDB.PingContext(ctx); err != nil {
					log.Printf("WARN: Could not ping database: %s, Will retry after 1 second", err)
					time.Sleep(time.Second)
				} else {
					return db, nil
				}
			}
		}
		return nil, err
	}

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
		&models.VideoMetadata{},
		&models.ShareToken{},
		&models.UserMediaData{},
	)

	return nil
}
