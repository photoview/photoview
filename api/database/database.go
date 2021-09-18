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

	"github.com/go-sql-driver/mysql"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetMysqlAddress(addressString string) (string, error) {
	if addressString == "" {
		return "", errors.New(fmt.Sprintf("Environment variable %s missing, exiting", utils.EnvMysqlURL.GetName()))
	}

	config, err := mysql.ParseDSN(addressString)
	if err != nil {
		return "", errors.Wrap(err, "Could not parse mysql url")
	}

	config.MultiStatements = true
	config.ParseTime = true

	return config.FormatDSN(), nil
}

func GetPostgresAddress(addressString string) (*url.URL, error) {
	if addressString == "" {
		return nil, errors.New(fmt.Sprintf("Environment variable %s missing, exiting", utils.EnvPostgresURL.GetName()))
	}

	address, err := url.Parse(addressString)
	if err != nil {
		return nil, errors.Wrap(err, "Could not parse postgres url")
	}

	return address, nil
}

func GetSqliteAddress(path string) (*url.URL, error) {
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

func ConfigureDatabase(config *gorm.Config) (*gorm.DB, error) {
	var databaseDialect gorm.Dialector
	switch drivers.DatabaseDriver() {
	case drivers.DatabaseDriverMysql:
		mysqlAddress, err := GetMysqlAddress(utils.EnvMysqlURL.GetValue())
		if err != nil {
			return nil, err
		}
		log.Printf("Connecting to MYSQL database: %s", mysqlAddress)
		databaseDialect = gorm_mysql.Open(mysqlAddress)

	case drivers.DatabaseDriverSqlite:
		sqliteAddress, err := GetSqliteAddress(utils.EnvSqlitePath.GetValue())
		if err != nil {
			return nil, err
		}
		log.Printf("Opening SQLITE database: %s", sqliteAddress)
		databaseDialect = sqlite.Open(sqliteAddress.String())

	case drivers.DatabaseDriverPostgres:
		postgresAddress, err := GetPostgresAddress(utils.EnvPostgresURL.GetValue())
		if err != nil {
			return nil, err
		}
		log.Printf("Connecting to POSTGRES database: %s", postgresAddress.Redacted())
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
		db, err = ConfigureDatabase(&config)
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

var database_models []interface{} = []interface{}{
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
	&models.UserAlbums{},
	&models.UserPreferences{},

	// Face detection
	&models.FaceGroup{},
	&models.ImageFace{},
}

func MigrateDatabase(db *gorm.DB) error {

	if err := db.SetupJoinTable(&models.User{}, "Albums", &models.UserAlbums{}); err != nil {
		log.Printf("Setup UserAlbums join table failed: %v\n", err)
	}

	if err := db.AutoMigrate(database_models...); err != nil {
		log.Printf("Auto migration failed: %v\n", err)
	}

	// v2.1.0 - Replaced by Media.CreatedAt
	if db.Migrator().HasColumn(&models.Media{}, "date_imported") {
		db.Migrator().DropColumn(&models.Media{}, "date_imported")
	}

	// v2.3.0 - Changed type of MediaEXIF.Exposure and MediaEXIF.Flash
	// from string values to decimal and int respectively
	if err := migrate_exif_fields(db); err != nil {
		log.Printf("Failed to run exif fields migration: %v\n", err)
	}

	// // PJ-Watson: Attempt to add new column for FaceGroup.PreviewImageFace
	if !(db.Migrator().HasColumn(&models.FaceGroup{}, "preview_image_face")) {
		db.Migrator().AddColumn(&models.FaceGroup{}, "preview_image_face")
	}
	if err := migrate_face_preview(db); err != nil {
		log.Printf("Failed to run face groups preview image migration: %v\n", err)
	}

	return nil
}

func ClearDatabase(db *gorm.DB) error {
	err := db.Transaction(func(tx *gorm.DB) error {

		db_driver := drivers.DatabaseDriver()

		if db_driver == drivers.DatabaseDriverMysql {
			if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 0;").Error; err != nil {
				return err
			}
		}

		dry_run := tx.Session(&gorm.Session{DryRun: true})
		for _, model := range database_models {
			// get table name of model structure
			table := dry_run.Find(model).Statement.Table

			switch db_driver {
			case drivers.DatabaseDriverPostgres:
				if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
					return err
				}
			case drivers.DatabaseDriverMysql:
				if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table)).Error; err != nil {
					return err
				}
			case drivers.DatabaseDriverSqlite:
				if err := tx.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
					return err
				}
			}

		}

		if db_driver == drivers.DatabaseDriverMysql {
			if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 1;").Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
