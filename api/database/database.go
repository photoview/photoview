package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/database/migrations"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/utils"

	"github.com/go-sql-driver/mysql"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GetMysqlAddress(addressString string) (string, error) {
	if addressString == "" {
		return "", fmt.Errorf("Environment variable %s missing, exiting", utils.EnvMysqlURL.GetName())
	}

	config, err := mysql.ParseDSN(addressString)
	if err != nil {
		return "", fmt.Errorf("could not parse mysql url: %w", err)
	}

	config.MultiStatements = true
	config.ParseTime = true

	return config.FormatDSN(), nil
}

func GetPostgresAddress(addressString string) (*url.URL, error) {
	if addressString == "" {
		return nil, fmt.Errorf("Environment variable %s missing, exiting", utils.EnvPostgresURL.GetName())
	}

	address, err := url.Parse(addressString)
	if err != nil {
		return nil, fmt.Errorf("could not parse postgres url: %w", err)
	}

	return address, nil
}

func GetSqliteAddress(path string) (*url.URL, error) {
	if path == "" {
		path = "photoview.db"
	}

	address, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("could not parse sqlite url (%s): %w", path, err)
	}

	queryValues := address.Query()
	queryValues.Add("cache", "shared")
	queryValues.Add("mode", "rwc")
	// queryValues.Add("_busy_timeout", "60000") // 1 minute
	queryValues.Add("_journal_mode", "WAL")    // Write-Ahead Logging (WAL) mode
	queryValues.Add("_locking_mode", "NORMAL") // allows concurrent reads and writes
	queryValues.Add("_foreign_keys", "ON")     // Enforc foreign key constraints.
	address.RawQuery = queryValues.Encode()

	// log.Panicf("%s", address.String())

	return address, nil
}

func ConfigureDatabase(config *gorm.Config) (*gorm.DB, error) {
	var databaseDialect gorm.Dialector
	driver := drivers.DatabaseDriverFromEnv()
	log.Printf("Utilizing %s database driver based on environment variables", driver)

	switch driver {
	case drivers.MYSQL:
		mysqlAddress, err := GetMysqlAddress(utils.EnvMysqlURL.GetValue())
		if err != nil {
			return nil, err
		}
		databaseDialect = gorm_mysql.Open(mysqlAddress)
	case drivers.SQLITE:
		sqliteAddress, err := GetSqliteAddress(utils.EnvSqlitePath.GetValue())
		if err != nil {
			return nil, err
		}
		databaseDialect = sqlite.Open(sqliteAddress.String())

	case drivers.POSTGRES:
		postgresAddress, err := GetPostgresAddress(utils.EnvPostgresURL.GetValue())
		if err != nil {
			return nil, err
		}
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
	&models.UserAlbumGrant{},
	&models.UserAlbumData{},
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
		return fmt.Errorf("auto migration failed: %w", err)
	}

	// Once SetupJoinTable registers UserAlbums as the join table for
	// User.Albums, GORM's schema cache for that struct type is pinned to the
	// relationship's own minimal schema (just the user_id/album_id key
	// columns) for the remainder of the process - a plain AutoMigrate call
	// (in either order relative to SetupJoinTable, and on every subsequent
	// call once this has happened once in this process) silently ignores
	// the Level/GrantedByUserID fields actually declared on the struct. This
	// is process-wide (GORM's schema cache isn't scoped to *gorm.DB), so it
	// also affects every later MigrateDatabase call in the same process
	// (e.g. repeated test runs), not just this one. Ensuring these two
	// columns exist is therefore done with raw SQL, bypassing GORM's
	// struct-based schema resolution for this table entirely.
	if err := ensureUserAlbumsColumns(db); err != nil {
		return fmt.Errorf("ensure user_albums columns failed: %w", err)
	}

	// v2.1.0 - Replaced by Media.CreatedAt
	if db.Migrator().HasColumn(&models.Media{}, "date_imported") {
		db.Migrator().DropColumn(&models.Media{}, "date_imported")
	}

	// v2.3.0 - Changed type of MediaEXIF.Exposure and MediaEXIF.Flash
	// from string values to decimal and int respectively
	if err := migrateExifFields(db); err != nil {
		log.Printf("Failed to run exif fields migration: %v\n", err)
	}

	// Remove invalid GPS data from DB
	if err := migrations.MigrateForExifGPSCorrection(db); err != nil {
		log.Printf("Failed to run exif GPS correction migration: %v\n", err)
	}

	// Replaced by per-album UserAlbums.Level. Unlike the cosmetic/data-quality
	// migrations above, a failure here must stop startup: the new level
	// column defaults to READ, so continuing would silently downgrade any
	// user who previously had can_upload=true until this migration
	// eventually succeeds.
	if err := migrations.MigrateCanUploadToAlbumLevel(db); err != nil {
		return fmt.Errorf("run can_upload to album level migration: %w", err)
	}

	// Backfill self-sourced UserAlbumGrant rows for any UserAlbums row that
	// predates grant provenance tracking, so PropagateAlbumLevel/
	// RevokeAlbumLevel's source-scoped writes behave correctly the first
	// time they touch a pre-existing grant. Must run after the can_upload
	// migration above, since that migration corrects user_albums.level
	// in place - backfilling before it would freeze the pre-correction level
	// into the grant row, which a later recompute could then revert back to.
	// This can't recover provenance for grants that were already silently
	// clobbered before this migration existed - it only prevents new
	// clobbers going forward.
	if err := backfillUserAlbumGrants(db); err != nil {
		return fmt.Errorf("backfill user_album_grants failed: %w", err)
	}

	// v2.5.0 - Remove Thumbnail Method for Downsampliing filters
	if db.Migrator().HasColumn(&models.SiteInfo{}, "thumbnail_method") {
		db.Migrator().DropColumn(&models.SiteInfo{}, "thumbnail_method")
	}

	return nil
}

// ensureUserAlbumsColumns adds the level/granted_by_user_id columns to
// user_albums if they don't already exist, using raw SQL and each driver's
// own column-introspection mechanism rather than GORM's struct-based
// schema resolution (see the comment at its call site for why).
func ensureUserAlbumsColumns(db *gorm.DB) error {
	existing, err := existingColumns(db, "user_albums")
	if err != nil {
		return fmt.Errorf("list user_albums columns: %w", err)
	}

	if !existing["level"] {
		if err := db.Exec("ALTER TABLE user_albums ADD COLUMN level VARCHAR(16) NOT NULL DEFAULT 'READ'").Error; err != nil {
			return fmt.Errorf("add level column: %w", err)
		}
	}

	if !existing["granted_by_user_id"] {
		if err := db.Exec("ALTER TABLE user_albums ADD COLUMN granted_by_user_id BIGINT").Error; err != nil {
			return fmt.Errorf("add granted_by_user_id column: %w", err)
		}
	}

	return nil
}

// backfillUserAlbumGrants inserts a self-sourced UserAlbumGrant row (source
// = the album the grant already sits on) for every UserAlbums row that has
// no UserAlbumGrant row yet. Idempotent: only touches (user, album) pairs
// with zero existing grant rows, so it's safe to run on every startup.
func backfillUserAlbumGrants(db *gorm.DB) error {
	return db.Exec(`
		INSERT INTO user_album_grants (user_id, album_id, source_album_id, level, granted_by_user_id)
		SELECT ua.user_id, ua.album_id, ua.album_id, ua.level, ua.granted_by_user_id
		FROM user_albums ua
		WHERE NOT EXISTS (
			SELECT 1 FROM user_album_grants uag
			WHERE uag.user_id = ua.user_id AND uag.album_id = ua.album_id
		)
	`).Error
}

// existingColumns returns the set of column names that currently exist on
// tableName, using whichever introspection mechanism the active driver
// supports.
func existingColumns(db *gorm.DB, tableName string) (map[string]bool, error) {
	var names []string

	if drivers.SQLITE.MatchDatabase(db) {
		var rows []struct{ Name string }
		if err := db.Raw(fmt.Sprintf("PRAGMA table_info(%s)", tableName)).Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			names = append(names, r.Name)
		}
	} else if drivers.POSTGRES.MatchDatabase(db) {
		if err := db.Raw(
			"SELECT column_name FROM information_schema.columns WHERE table_name = ? AND table_schema = current_schema()",
			tableName,
		).Scan(&names).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.Raw(
			"SELECT column_name FROM information_schema.columns WHERE table_name = ? AND table_schema = DATABASE()",
			tableName,
		).Scan(&names).Error; err != nil {
			return nil, err
		}
	}

	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[strings.ToLower(n)] = true
	}
	return set, nil
}

func ClearDatabase(db *gorm.DB) error {
	var errs []error
	for _, model := range database_models {
		if err := db.Migrator().DropTable(model); err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("drop tables error: %w", err)
	}

	return nil
}
