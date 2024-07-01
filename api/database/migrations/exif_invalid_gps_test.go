package migrations_test

import (
	"bufio"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoview/photoview/api/database/migrations"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
)

func TestExifMigration(t *testing.T) {
	envFile, err := os.Open("/home/runner/work/photoview/photoview/api/testing.env")
	if err != nil {
		t.Fatalf("Failed to open environment file: %v", err)
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			t.Fatalf("Invalid line in environment file: %s", line)
		}
		key, value := parts[0], strings.Trim(parts[1], "'")
		os.Setenv(key, value)
	}

	db := test_utils.DatabaseTest(t)
	defer db.Exec("DELETE FROM media_exif") // Clean up after test

	// Create test data
	exifEntries := []models.MediaEXIF{
		{GPSLatitude: floatPtr(90.1), GPSLongitude: floatPtr(90.0)},   // Invalid GPSLatitude
		{GPSLatitude: floatPtr(-90.1), GPSLongitude: floatPtr(-90.0)}, // Invalid GPSLatitude
		{GPSLatitude: floatPtr(90.0), GPSLongitude: floatPtr(90.1)},   // Invalid GPSLongitude
		{GPSLatitude: floatPtr(-90.0), GPSLongitude: floatPtr(-90.1)}, // Invalid GPSLongitude
		{GPSLatitude: floatPtr(90.0), GPSLongitude: floatPtr(90.0)},   // Valid GPS data
		{GPSLatitude: floatPtr(-90.0), GPSLongitude: floatPtr(-90.0)}, // Valid GPS data
		{GPSLatitude: floatPtr(90.1), GPSLongitude: floatPtr(90.1)},   // Invalid GPSLatitude and GPSLongitude
		{GPSLatitude: floatPtr(-90.1), GPSLongitude: floatPtr(-90.1)}, // Invalid GPSLatitude and GPSLongitude
	}

	// Insert test data
	for _, entry := range exifEntries {
		assert.NoError(t, db.Create(&entry).Error)
	}

	// Run migration
	assert.NoError(t, migrations.MigrateForExifGPSCorrection(db))

	// Validate the results
	var results []models.MediaEXIF
	assert.NoError(t, db.Find(&results).Error)

	for _, entry := range results {
		if entry.GPSLatitude != nil {
			assert.LessOrEqual(t, math.Abs(*entry.GPSLatitude), 90.0, "GPSLatitude should be within [-90, 90]: %+v", entry)
		}
		if entry.GPSLongitude != nil {
			assert.LessOrEqual(t, math.Abs(*entry.GPSLongitude), 90.0, "GPSLongitude should be within [-90, 90]: %+v", entry)
		}
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
