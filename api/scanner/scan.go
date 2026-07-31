package scanner

import (
	"context"
	"fmt"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/queue"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"gorm.io/gorm"
)

// AddUser finds all root albums owned by user and queues them for scanning.
// It does not block until scanning finishes.
func AddUser(db *gorm.DB, user *models.User) error {
	if !queue.Initialized() {
		return fmt.Errorf("scanner queue not initialized")
	}

	cache := scanner_cache.MakeAlbumCache()
	albums, errs := FindAlbumsForUser(db, user, cache)
	for _, err := range errs {
		if err != nil {
			return fmt.Errorf("find albums for user (user_id: %d): %w", user.ID, err)
		}
	}

	for _, album := range albums {
		if err := queue.SubmitAlbum(album, cache); err != nil {
			return err
		}
	}

	return nil
}

// AddAll queues every user's albums for scanning.
func AddAll(db *gorm.DB) error {
	if !queue.Initialized() {
		return fmt.Errorf("scanner queue not initialized")
	}

	var users []*models.User
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("get all users from database: %w", err)
	}

	for _, user := range users {
		if err := AddUser(db, user); err != nil {
			return fmt.Errorf("add user to queue (user_id: %d): %w", user.ID, err)
		}
	}

	return nil
}

// ProcessMedia (re)processes a single already-known media, submitting it to
// the scanner queue and blocking until processing has finished. Used to fix
// up one file (e.g. regenerate a cache file found missing while serving it)
// without starting a dedicated worker for it.
func ProcessMedia(ctx context.Context, db *gorm.DB, media *models.Media) error {
	if !queue.Initialized() {
		return fmt.Errorf("scanner queue not initialized")
	}

	var album models.Album
	if err := db.Model(media).Association("Album").Find(&album); err != nil {
		return err
	}

	cache := scanner_cache.MakeAlbumCache()
	return queue.SubmitMedia(ctx, db, &album, cache, media.Path)
}
