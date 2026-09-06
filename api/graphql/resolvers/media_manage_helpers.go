package resolvers

// Helper functions for media_manage.go, kept in a separate file (not the
// gqlgen-managed resolver file) since gqlgen's codegen doesn't recognize
// non-resolver top-level declarations and will otherwise comment them out
// on the next `go generate` as "unknown code".

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

// deleteOneMedia moves mediaID's file (and its sidecar file, if any) to a
// hidden trash folder next to its album, and removes it from the library.
// Shared by DeleteMedia and DeleteMediaList so both go through the exact
// same permission check and cleanup steps.
func deleteOneMedia(db *gorm.DB, user *models.User, mediaID int) error {
	var media models.Media
	if err := db.First(&media, mediaID).Error; err != nil {
		return err
	}

	var album models.Album
	if err := db.First(&album, media.AlbumID).Error; err != nil {
		return err
	}

	canDelete, err := user.HasAlbumLevel(db, &album, models.AlbumPermissionLevelUpload)
	if err != nil {
		return err
	}
	if !canDelete {
		return auth.ErrUnauthorized
	}

	trashDir := filepath.Join(album.Path, ".trash")
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return fmt.Errorf("could not create trash folder: %w", err)
	}

	timestamp := time.Now().Unix()
	trashPath := filepath.Join(trashDir, fmt.Sprintf("%d-%s", timestamp, filepath.Base(media.Path)))
	if err := os.Rename(media.Path, trashPath); err != nil {
		return fmt.Errorf("could not move file to trash: %w", err)
	}

	if media.SideCarPath != nil {
		sidecarTrashPath := filepath.Join(trashDir, fmt.Sprintf("%d-%s", timestamp, filepath.Base(*media.SideCarPath)))
		// Best-effort: the primary file is already safely trashed, and the
		// sidecar isn't required for the library entry to be gone.
		_ = os.Rename(*media.SideCarPath, sidecarTrashPath)
	}

	// MediaURL and ImageFace rows point *at* Media (has-many) and are
	// cascade-deleted below, but the cached thumbnail/highres/web-video
	// files MediaURL points at are not - remove them explicitly,
	// mirroring cleanup_tasks' album-level cache cleanup.
	cacheDir := filepath.Join(utils.MediaCachePath(), strconv.Itoa(media.AlbumID), strconv.Itoa(media.ID))
	_ = os.RemoveAll(cacheDir)

	// MediaEXIF/VideoMetadata are the opposite direction (Media points at
	// them via ExifID/VideoMetadataID), so deleting Media does not cascade
	// to them - delete them explicitly to avoid leaving orphan rows.
	deleteErr := db.Transaction(func(tx *gorm.DB) error {
		if media.ExifID != nil {
			if err := tx.Delete(&models.MediaEXIF{}, *media.ExifID).Error; err != nil {
				return err
			}
		}
		if media.VideoMetadataID != nil {
			if err := tx.Delete(&models.VideoMetadata{}, *media.VideoMetadataID).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&media).Error
	})
	if deleteErr != nil {
		// The file is already in the trash but the DB rows remain -
		// surface the error rather than silently leaving a stale entry.
		return fmt.Errorf("moved to trash, but failed to remove library entry: %w", deleteErr)
	}

	return nil
}
