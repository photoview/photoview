package database

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getMysqlAddress() (*url.URL, error) {
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
	return address, nil
}

// SetupDatabase connects to the database using environment variables
func SetupDatabase() (*gorm.DB, error) {

	config := gorm.Config{}

	// Enable database debug logging
	config.Logger = logger.Default.LogMode(logger.Info)

	var databaseDialect gorm.Dialector
	switch strings.ToLower(os.Getenv("PHOTOVIEW_DATABASE_DRIVER")) {
	case "mysql":
		mysqlAddress, err := getMysqlAddress()
		if err != nil {
			return nil, err
		}
		log.Printf("Connecting to database: %s", mysqlAddress)
		databaseDialect = mysql.Open(mysqlAddress.String())

	case "sqlite":
		path := os.Getenv("PHOTOVIEW_SQLITE_PATH")
		if path == "" {
			path = "photoview.db"
		}
		databaseDialect = sqlite.Open(path)
	}

	db, err := gorm.Open(databaseDialect, &config)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to database")
	}

	// TODO: Add connection retries

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
