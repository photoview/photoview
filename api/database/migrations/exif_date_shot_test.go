package migrations_test

import (
	"testing"
	"time"

	"github.com/photoview/photoview/api/database/migrations"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
)

func parseTime(t *testing.T, timeStr string, loc *time.Location) *time.Time {
	t.Helper()

	ret, err := time.ParseInLocation("2006-01-02T15:04:05", timeStr, loc)
	if err != nil {
		t.Fatalf("parse time %q error: %v", timeStr, err)
	}

	return &ret
}

func TestExifDateShotMigration(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	defer db.Exec("DELETE FROM media_exif") // Clean up after test

	cases := []struct {
		from *time.Time
		want string
	}{
		{parseTime(t, "2025-10-10T12:00:00", time.Local), "2025-10-10T12:00:00.000"},
		// It can't test with timezone other than local timezone, because different database treats timezone in different ways.
	}

	// Insert test data
	for _, c := range cases {
		entry := models.MediaEXIF{
			DateShot: c.from,
		}

		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("can't create entry %v: %v", entry, err)
		}
	}

	// Run migration
	if err := migrations.MigrateDateShot(db); err != nil {
		t.Fatalf("can't migrate dateShot to dateShotStr: %v", err)
	}

	// Validate the results
	var results []models.MediaEXIF
	if err := db.Order("id").Find(&results).Error; err != nil {
		t.Fatalf("can't query media exif after migration: %v", err)
	}

	for i, entry := range results {
		if got, want := *entry.DateShotStr, cases[i].want; got != want {
			t.Errorf("migrate dateShot %v error: got: %q, want: %q", cases[i].from, got, want)
		}
	}
}
