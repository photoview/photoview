package models

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FindOrCreateAlbum looks up an existing Album by album.Path's hash; if none
// exists, it inserts album. If a concurrent caller wins the race to insert
// the same path first, *album is replaced with that caller's row instead of
// erroring, so every concurrent caller converges on the same row. Callers
// are still responsible for ensuring the right Owners are appended
// afterwards - this only makes the row itself idempotent by path.
func FindOrCreateAlbum(db *gorm.DB, album *Album) (created bool, err error) {
	hash := MD5Hash(album.Path)

	var existing Album
	err = db.Where("path_hash = ?", hash).First(&existing).Error
	switch {
	case err == nil:
		*album = existing
		return false, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return false, err
	}

	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path_hash"}},
		DoNothing: true,
	}).Create(album)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	// Lost the race - fetch the row the winner created.
	if err := db.Where("path_hash = ?", hash).First(&existing).Error; err != nil {
		return false, err
	}
	*album = existing
	return false, nil
}

// FindOrCreateMedia looks up an existing Media by media.Path's hash; if none
// exists, it inserts media. If a concurrent caller wins the race to insert
// the same path first, *media is replaced with that caller's row instead of
// erroring, so every concurrent caller converges on the same row.
func FindOrCreateMedia(db *gorm.DB, media *Media) (created bool, err error) {
	hash := MD5Hash(media.Path)

	var existing Media
	err = db.Where("path_hash = ?", hash).First(&existing).Error
	switch {
	case err == nil:
		*media = existing
		return false, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return false, err
	}

	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path_hash"}},
		DoNothing: true,
	}).Create(media)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	// Lost the race - fetch the row the winner created.
	if err := db.Where("path_hash = ?", hash).First(&existing).Error; err != nil {
		return false, err
	}
	*media = existing
	return false, nil
}

// UpsertMedia creates media, or overwrites the existing row for the same
// path with media's current field values. Unlike FindOrCreateMedia (which
// leaves an existing row untouched), this always writes through - used by
// queue's persist() where media has already been fully re-evaluated and a
// second, separate update call would otherwise be needed right after.
func UpsertMedia(db *gorm.DB, media *Media) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "path_hash"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title", "album_id", "date_shot", "type",
			"side_car_path", "side_car_hash", "blurhash",
		}),
	}).Create(media).Error
}

// UpsertMediaURL creates url, or overwrites the existing row for the same
// (media_id, purpose) pair if one already exists - whether from a prior scan
// that's now being refreshed (e.g. after a sidecar change forces the cache
// file to be regenerated) or from a concurrent caller that created it first.
// Callers may pass in a MediaURL that was fetched from an existing row (e.g.
// to update it in place); its ID is cleared here so the insert always
// targets the (media_id, purpose) conflict rather than colliding on the
// primary key too.
func UpsertMediaURL(db *gorm.DB, url *MediaURL) error {
	url.ID = 0
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "media_id"}, {Name: "purpose"}},
		DoUpdates: clause.AssignmentColumns([]string{"media_name", "width", "height", "content_type", "file_size"}),
	}).Create(url).Error
}
