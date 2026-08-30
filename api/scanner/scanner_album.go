package scanner

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/log"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func NewRootAlbum(db *gorm.DB, rootPath string, owner *models.User) (*models.Album, error) {
	rootPath = filepath.Clean(rootPath)

	if !ValidRootPath(rootPath) {
		return nil, ErrorInvalidRootPath
	}

	if !path.IsAbs(rootPath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		rootPath = path.Join(wd, rootPath)
	}

	album := models.Album{
		Title:  path.Base(rootPath),
		Path:   rootPath,
		Owners: []models.User{*owner},
	}

	created, err := models.FindOrCreateAlbum(db, &album)
	if err != nil {
		return nil, err
	}
	if created {
		return &album, nil
	}

	var matchedUserAlbumCount int64
	if err := db.Table("user_albums").Where("user_id = ?", owner.ID).Where("album_id = ?", album.ID).Count(&matchedUserAlbumCount).Error; err != nil {
		return nil, err
	}

	if matchedUserAlbumCount > 0 {
		return nil, errors.New(fmt.Sprintf("user already owns a path containing this path: %s", rootPath))
	}

	if err := db.Model(owner).Association("Albums").Append(&album); err != nil {
		return nil, errors.Wrap(err, "add owner to already existing album")
	}

	return &album, nil
}

var ErrorInvalidRootPath = errors.New("invalid root path")

func ValidRootPath(rootPath string) bool {
	resolvedPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		log.Warn(nil, "invalid root path", "root_path", rootPath, "error", err)
		return false
	}

	// Confirm the resolved path is actually a directory, not a file.
	info, err := os.Stat(resolvedPath)
	if err != nil {
		log.Warn(nil, "invalid root path after symlink resolution",
			"root_path", rootPath, "resolved_path", resolvedPath, "error", err)
		return false
	}
	if !info.IsDir() {
		log.Warn(nil, "root path is not a directory", "root_path", rootPath)
		return false
	}

	return true
}
