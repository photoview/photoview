package database

import (
	"context"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getMysqlAddress() (*url.URL, error) {
	address, err := url.Parse(os.Getenv("PHOTOVIEW_MYSQL_URL"))
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse mysql url")
	}

	if address.String() == "" {
		return nil, errors.New("Environment variable PHOTOVIEW_MYSQL_URL missing, exiting")
	}

	queryValues := address.Query()
	queryValues.Add("multiStatements", "true")
	queryValues.Add("parseTime", "true")

	address.RawQuery = queryValues.Encode()
	return address, nil
}

func getSqliteAddress() (*url.URL, error) {
	path := os.Getenv("PHOTOVIEW_SQLITE_PATH")
	if path == "" {
		path = "photoview.db"
	}

	address, err := url.Parse(path)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not parse sqlite url (%s)", path)
	}

	queryValues := address.Query()
	queryValues.Add("cache", "shared")
	queryValues.Add("mode", "rwc")
	// queryValues.Add("_busy_timeout", "60000") // 1 minute
	address.RawQuery = queryValues.Encode()

	// log.Panicf("%s", address.String())

	return address, nil
}

// SetupDatabase connects to the database using environment variables
func SetupDatabase() (*gorm.DB, error) {

	config := gorm.Config{}

	// Enable database debug logging
	config.Logger = logger.Default.LogMode(logger.Info)

	var databaseDialect gorm.Dialector
	switch drivers.DatabaseDriver() {
	case drivers.DatabaseDriverMysql:
		mysqlAddress, err := getMysqlAddress()
		if err != nil {
			return nil, err
		}
		log.Printf("Connecting to database: %s", mysqlAddress)
		databaseDialect = mysql.Open(mysqlAddress.String())

	case drivers.DatabaseDriverSqlite:
		sqliteAddress, err := getSqliteAddress()
		if err != nil {
			return nil, err
		}
		databaseDialect = sqlite.Open(sqliteAddress.String())
	}

	db, err := gorm.Open(databaseDialect, &config)
	sqlDB, dbErr := db.DB()
	if dbErr != nil {
		log.Println(dbErr)
		return nil, dbErr
	}

	sqlDB.SetMaxOpenConns(80)

	if err != nil {
		for retryCount := 1; retryCount <= 5; retryCount++ {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			if err := sqlDB.PingContext(ctx); err == nil {
				cancel()
				return db, nil
			}

			cancel()

			log.Printf("WARN: Could not ping database: %s. Will retry after 5 seconds\n", err)
			time.Sleep(time.Duration(5) * time.Second)
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
