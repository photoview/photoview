package models_test

import (
	"testing"
	"time"

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
		date      string
		offsetSec *int
		want      string
	}{
		{"NoSubSecNoOffset", "2025:11:01 14:02:03", nil, "2025-11-01T14:02:03"},
		{"SubSecNoOffset", "2025:11:01 14:02:03.123", nil, "2025-11-01T14:02:03.123"},
		{"NoSubSecWithOffset", "2025:11:01 14:02:03", new(60 * 60), "2025-11-01T14:02:03+01:00"},
		{"SubSecWithOffset", "2025:11:01 14:02:03.123", new(60 * 60), "2025-11-01T14:02:03.123+01:00"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			date, err := time.ParseInLocation(layout, tc.date, time.UTC)
			if err != nil {
				t.Fatalf("parse time %q error: %v", tc.date, err)
			}

			exif := models.MediaEXIF{
				DateShot:      &date,
				OffsetSecShot: tc.offsetSec,
			}

			if err := db.Save(&exif).Error; err != nil {
				t.Fatalf("store exif error: %v", err)
			}

			var got models.MediaEXIF
			if err := db.Where("id = ?", exif.ID).First(&got).Error; err != nil {
				t.Fatalf("get exif error: %v", err)
			}

			if got, want := *got.DateShotWithOffset(), tc.want; got != want {
				t.Errorf("got = %q, want: %q", got, want)
			}
		})
	}
}
