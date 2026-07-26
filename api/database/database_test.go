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
