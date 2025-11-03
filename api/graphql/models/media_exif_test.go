package models_test

import (
	"testing"
	"time"

	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
)

// Different database has different behavior when storing date with timezone.
// - SQLite: keep the original timezone
// - MySQL/MariaDB: store in UTC
// - PostgreSQL: store in the timezone of the db client
// We cannot maintain consistent behavior across different databases without the extra offset field.

const (
	layout           = "2006:01:02 15:04:05"
	layoutWithOffset = "2006:01:02 15:04:05-07:00"
)

func TestDatabaseReproduceDateWithOffset(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	tests := []struct {
		name      string
		pattern   string
		date      string
		offsetSec int
		want      string
	}{
		{"NoTimezone", layout, "2025:11:01 14:02:03", 60 * 60, "2025-11-01T15:02:03+01:00"},
		{"WithTimezone", layoutWithOffset, "2025:11:01 14:02:03+01:00", 2 * 60 * 60, "2025-11-01T15:02:03+02:00"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			date, err := time.Parse(tc.pattern, tc.date)
			if err != nil {
				t.Fatalf("parse time %q in pattern %q error: %v", tc.date, tc.pattern, err)
			}
			exif := models.MediaEXIF{
				DateShot:      &date,
				OffsetSecShot: &tc.offsetSec,
			}

			if err := db.Save(&exif).Error; err != nil {
				t.Fatalf("store exif error: %v", err)
			}

			var got models.MediaEXIF
			if err := db.Where("id = ?", exif.ID).First(&got).Error; err != nil {
				t.Fatalf("get exif error: %v", err)
			}

			if got, want := got.DateShotWithOffset().Format(time.RFC3339), tc.want; got != want {
				t.Errorf("got = %q, want: %q", got, want)
			}
		})
	}
}

func TestDatabaseReproduceDateWithoutOffsetSQLite(t *testing.T) {
	if driver := drivers.DatabaseDriverFromEnv(); driver != drivers.SQLITE {
		t.Logf("skip testing with database %q", driver)
		t.Skip()
	}

	db := test_utils.DatabaseTest(t)

	tests := []struct {
		name    string
		pattern string
		date    string
		want    string
	}{
		{"NoTimezone", layout, "2025:11:01 14:02:03", "2025-11-01T14:02:03Z"},
		{"WithTimezone", layoutWithOffset, "2025:11:01 14:02:03+01:00", "2025-11-01T14:02:03+01:00"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			date, err := time.Parse(tc.pattern, tc.date)
			if err != nil {
				t.Fatalf("parse time %q in pattern %q error: %v", tc.date, tc.pattern, err)
			}
			exif := models.MediaEXIF{
				DateShot:      &date,
				OffsetSecShot: nil,
			}

			if err := db.Save(&exif).Error; err != nil {
				t.Fatalf("store exif error: %v", err)
			}

			var got models.MediaEXIF
			if err := db.Where("id = ?", exif.ID).First(&got).Error; err != nil {
				t.Fatalf("get exif error: %v", err)
			}

			if got, want := got.DateShotWithOffset().Format(time.RFC3339), tc.want; got != want {
				t.Errorf("got = %q, want: %q", got, want)
			}
		})
	}
}

func TestDatabaseReproduceDateWithoutOffsetMySQL(t *testing.T) {
	if driver := drivers.DatabaseDriverFromEnv(); driver != drivers.MYSQL {
		t.Logf("skip testing with database %q", driver)
		t.Skip()
	}

	db := test_utils.DatabaseTest(t)

	tests := []struct {
		name    string
		pattern string
		date    string
		want    string
	}{
		{"NoTimezone", layout, "2025:11:01 14:02:03", "2025-11-01T14:02:03Z"},
		{"WithTimezone", layoutWithOffset, "2025:11:01 14:02:03+01:00", "2025-11-01T13:02:03Z"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			date, err := time.Parse(tc.pattern, tc.date)
			if err != nil {
				t.Fatalf("parse time %q in pattern %q error: %v", tc.date, tc.pattern, err)
			}
			exif := models.MediaEXIF{
				DateShot:      &date,
				OffsetSecShot: nil,
			}

			if err := db.Save(&exif).Error; err != nil {
				t.Fatalf("store exif error: %v", err)
			}

			var got models.MediaEXIF
			if err := db.Where("id = ?", exif.ID).First(&got).Error; err != nil {
				t.Fatalf("get exif error: %v", err)
			}

			if got, want := got.DateShotWithOffset().Format(time.RFC3339), tc.want; got != want {
				t.Errorf("got = %q, want: %q", got, want)
			}
		})
	}
}

func timeToLocal(t *testing.T, value string) string {
	t.Helper()

	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q in rfc3339 error: %v", value, err)
	}

	return date.In(time.Local).Format(time.RFC3339)
}

func TestDatabaseReproduceDateWithoutOffsetPostgreSQL(t *testing.T) {
	if driver := drivers.DatabaseDriverFromEnv(); driver != drivers.POSTGRES {
		t.Logf("skip testing with database %q", driver)
		t.Skip()
	}

	db := test_utils.DatabaseTest(t)

	tests := []struct {
		name    string
		pattern string
		date    string
		want    string
	}{
		{"NoTimezone", layout, "2025:11:01 14:02:03", timeToLocal(t, "2025-11-01T14:02:03Z")},
		{"WithTimezone", layoutWithOffset, "2025:11:01 14:02:03+01:00", timeToLocal(t, "2025-11-01T13:02:03Z")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			date, err := time.Parse(tc.pattern, tc.date)
			if err != nil {
				t.Fatalf("parse time %q in pattern %q error: %v", tc.date, tc.pattern, err)
			}
			exif := models.MediaEXIF{
				DateShot:      &date,
				OffsetSecShot: nil,
			}

			if err := db.Save(&exif).Error; err != nil {
				t.Fatalf("store exif error: %v", err)
			}

			var got models.MediaEXIF
			if err := db.Where("id = ?", exif.ID).First(&got).Error; err != nil {
				t.Fatalf("get exif error: %v", err)
			}

			if got, want := got.DateShotWithOffset().Format(time.RFC3339), tc.want; got != want {
				t.Errorf("got = %q, want: %q", got, want)
			}
		})
	}
}
