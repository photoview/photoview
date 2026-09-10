package database_test

import (
	"testing"

	"github.com/photoview/photoview/api/database"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	test_utils.UnitTestRun(m)
}

func TestMigrateDatabaseAddsShareTokenLabel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`
		CREATE TABLE share_tokens (
			id integer PRIMARY KEY,
			value text NOT NULL,
			owner_id integer NOT NULL
		)
	`).Error)
	require.False(t, db.Migrator().HasColumn(&models.ShareToken{}, "label"))

	require.NoError(t, database.MigrateDatabase(db))
	require.True(t, db.Migrator().HasColumn(&models.ShareToken{}, "label"))
	require.False(t, db.Migrator().HasIndex(&models.ShareToken{}, "Label"))
}

func TestMigrateDatabaseBackfillsUserAlbumGrants(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	user, err := models.RegisterUser(db, "backfill_user", nil, false)
	require.NoError(t, err)
	album := models.Album{Title: "backfill_album", Path: "/photos/backfill_album"}
	require.NoError(t, db.Save(&album).Error)

	// Simulate a UserAlbums row that predates grant provenance tracking: a
	// direct insert with no matching UserAlbumGrant row, bypassing
	// PropagateAlbumLevel entirely.
	require.NoError(t, db.Create(&models.UserAlbums{
		UserID: user.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelUpload,
	}).Error)

	var countBefore int64
	require.NoError(t, db.Model(&models.UserAlbumGrant{}).
		Where("user_id = ? AND album_id = ?", user.ID, album.ID).Count(&countBefore).Error)
	require.Zero(t, countBefore)

	require.NoError(t, database.MigrateDatabase(db))

	var grant models.UserAlbumGrant
	require.NoError(t, db.Where("user_id = ? AND album_id = ?", user.ID, album.ID).First(&grant).Error)
	require.Equal(t, album.ID, grant.SourceAlbumID, "a backfilled grant is self-sourced on the album it already sits on")
	require.Equal(t, models.AlbumPermissionLevelUpload, grant.Level)

	// Running the migration again must not duplicate the backfilled row.
	require.NoError(t, database.MigrateDatabase(db))
	var countAfter int64
	require.NoError(t, db.Model(&models.UserAlbumGrant{}).
		Where("user_id = ? AND album_id = ?", user.ID, album.ID).Count(&countAfter).Error)
	require.EqualValues(t, 1, countAfter)
}
