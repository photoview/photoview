package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getMysqlAddress() (*url.URL, error) {
	addressString := utils.EnvMysqlURL.GetValue()
	if addressString == "" {
		return nil, errors.New(fmt.Sprintf("Environment variable %s missing, exiting", utils.EnvMysqlURL.GetName()))
	}

	address, err := url.Parse(addressString)
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse mysql url")
	}

	queryValues := address.Query()
	queryValues.Add("multiStatements", "true")
	queryValues.Add("parseTime", "true")

	address.RawQuery = queryValues.Encode()
	return address, nil
}

func getPostgresAddress() (*url.URL, error) {
	addressString := utils.EnvPostgresURL.GetValue()
	if addressString == "" {
		return nil, errors.New(fmt.Sprintf("Environment variable %s missing, exiting", utils.EnvPostgresURL.GetName()))
	}

	address, err := url.Parse(addressString)
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse postgres url")
	}

	return address, nil
}

func getSqliteAddress() (*url.URL, error) {
	path := utils.EnvSqlitePath.GetValue()
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

func configureDatabase(config *gorm.Config) (*gorm.DB, error) {
	var databaseDialect gorm.Dialector
	switch drivers.DatabaseDriver() {
	case drivers.DatabaseDriverMysql:
		mysqlAddress, err := getMysqlAddress()
		if err != nil {
			return nil, err
		}
		log.Printf("Connecting to MYSQL database: %s", mysqlAddress)
		databaseDialect = mysql.Open(mysqlAddress.String())

	case drivers.DatabaseDriverSqlite:
		sqliteAddress, err := getSqliteAddress()
		if err != nil {
			return nil, err
		}
		log.Printf("Opening SQLITE database: %s", sqliteAddress)
		databaseDialect = sqlite.Open(sqliteAddress.String())

	case drivers.DatabaseDriverPostgres:
		postgresAddress, err := getPostgresAddress()
		if err != nil {
			return nil, err
		}
		log.Printf("Connecting to POSTGRES database: %s", postgresAddress)
		databaseDialect = postgres.Open(postgresAddress.String())
	}

	db, err := gorm.Open(databaseDialect, config)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// SetupDatabase connects to the database using environment variables
func SetupDatabase() (*gorm.DB, error) {

	config := gorm.Config{}

	// Configure database logging
	if utils.DevelopmentMode() {
		config.Logger = logger.Default.LogMode(logger.Info)
	} else {
		config.Logger = logger.Default.LogMode(logger.Warn)
	}

	var db *gorm.DB

	for retryCount := 1; retryCount <= 5; retryCount++ {

		var err error
		db, err = configureDatabase(&config)
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr != nil {
				return nil, dbErr
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err = sqlDB.PingContext(ctx)
			cancel()

			sqlDB.SetMaxOpenConns(80)

			if err == nil {
				return db, nil
			}
		}

		log.Printf("WARN: Could not ping database: %s. Will retry after 5 seconds\n", err)
		time.Sleep(time.Duration(5) * time.Second)
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
