package migrations_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoview/photoview/api/database/migrations"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
)

func TestDedupMediaURLsMigration(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	defer db.Exec("DELETE FROM media_urls")
	defer db.Exec("DELETE FROM media")
	defer db.Exec("DELETE FROM albums")

	album := &models.Album{Title: "dedup test", Path: "/dedup-test"}
	assert.NoError(t, db.Create(album).Error)

	media := &models.Media{
		Title:    "photo.jpg",
		Path:     "/dedup-test/photo.jpg",
		AlbumID:  album.ID,
		Type:     models.MediaTypePhoto,
		DateShot: time.Now(),
	}
	assert.NoError(t, db.Create(media).Error)

	// AutoMigrate has already created the unique (media_id, purpose) index
	// by the time DatabaseTest returns. Drop it so duplicate rows - the
	// exact pre-migration state this migration exists to clean up - can be
	// inserted below.
	assert.NoError(t, db.Migrator().DropIndex(&models.MediaURL{}, "idx_media_url_media_purpose"))

	base := time.Now().Truncate(time.Second)

	type row struct {
		purpose     models.MediaPurpose
		mediaName   string
		contentType string
		updatedAt   time.Time
	}
	rows := []row{
		// Duplicate group: three thumbnail rows, the highest updated_at
		// should be the one that survives.
		{models.PhotoThumbnail, "thumb-old.jpg", "image/jpeg", base},
		{models.PhotoThumbnail, "thumb-newest.jpg", "image/jpeg", base.Add(2 * time.Hour)},
		{models.PhotoThumbnail, "thumb-mid.jpg", "image/jpeg", base.Add(1 * time.Hour)},

		// Duplicate group with a tied updated_at: the higher id (the
		// later insert) should be the tie-break winner.
		{models.PhotoHighRes, "highres-a.jpg", "image/jpeg", base},
		{models.PhotoHighRes, "highres-b.jpg", "image/jpeg", base},

		// Not a duplicate: only one row for this purpose, must survive untouched.
		{models.VideoWeb, "video.mp4", "video/mp4", base},
	}
	for _, r := range rows {
		url := models.MediaURL{
			MediaID:     media.ID,
			Purpose:     r.purpose,
			MediaName:   r.mediaName,
			ContentType: r.contentType,
		}
		url.UpdatedAt = r.updatedAt
		assert.NoError(t, db.Create(&url).Error)
	}

	assert.NoError(t, migrations.MigrateForDuplicateMediaURLs(db))

	var remaining []models.MediaURL
	assert.NoError(t, db.Where("media_id = ?", media.ID).Find(&remaining).Error)

	byPurpose := make(map[models.MediaPurpose][]models.MediaURL, 3)
	for _, u := range remaining {
		byPurpose[u.Purpose] = append(byPurpose[u.Purpose], u)
	}

	if assert.Len(t, byPurpose[models.PhotoThumbnail], 1) {
		assert.Equal(t, "thumb-newest.jpg", byPurpose[models.PhotoThumbnail][0].MediaName)
	}
	if assert.Len(t, byPurpose[models.PhotoHighRes], 1) {
		assert.Equal(t, "highres-b.jpg", byPurpose[models.PhotoHighRes][0].MediaName)
	}
	if assert.Len(t, byPurpose[models.VideoWeb], 1) {
		assert.Equal(t, "video.mp4", byPurpose[models.VideoWeb][0].MediaName)
	}
}
