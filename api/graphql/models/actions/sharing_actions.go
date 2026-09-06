package actions

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// AlbumPermissions lists every user with an explicit access grant directly
// on albumID. Callers must already have verified the viewer is allowed to
// see this (the album's owner, or an admin) - this function itself does no
// authorization check.
func AlbumPermissions(db *gorm.DB, albumID int) ([]*models.AlbumPermission, error) {
	var rows []models.UserAlbums
	if err := db.Where("album_id = ?", albumID).Find(&rows).Error; err != nil {
		return nil, err
	}

	permissions := make([]*models.AlbumPermission, 0, len(rows))
	for _, row := range rows {
		var user models.User
		if err := db.First(&user, row.UserID).Error; err != nil {
			return nil, err
		}
		permissions = append(permissions, &models.AlbumPermission{
			User:  &user,
			Level: row.Level,
		})
	}

	return permissions, nil
}

// ShareAlbum grants targetUserID access to albumID at level, on behalf of
// actor. Only albumID's owner (an explicit grant with no GrantedByUserID,
// i.e. it traces to an admin grant rather than another user's share) may
// do this, and only up to their own level - admins bypass both checks.
// Re-sharing an already-shared album updates the level and re-propagates it
// across the album's current subtree, so already-scanned descendants never
// keep a stale level.
func GrantAlbumAccess(db *gorm.DB, actor *models.User, albumID int, targetUserID int, level models.AlbumPermissionLevel) (*models.AlbumPermission, error) {
	var album models.Album
	if err := db.First(&album, albumID).Error; err != nil {
		return nil, err
	}

	if !actor.Admin {
		grant, err := actor.EffectiveGrant(db, &album)
		if err != nil {
			return nil, err
		}
		if grant == nil || grant.GrantedByUserID != nil {
			return nil, errors.New("only the owner of a folder may share it")
		}
		if level.HigherThan(grant.Level) {
			return nil, errors.New("cannot grant a higher permission level than your own")
		}
	}

	var targetUser models.User
	if err := db.First(&targetUser, targetUserID).Error; err != nil {
		return nil, errors.Wrap(err, "find target user")
	}

	if err := models.PropagateAlbumLevel(db, albumID, targetUserID, level, &actor.ID); err != nil {
		return nil, err
	}

	return &models.AlbumPermission{User: &targetUser, Level: level}, nil
}

// RevokeAlbumAccess removes targetUserID's access to albumID (and every
// current descendant), on behalf of actor. Only the album's owner or an
// admin may do this.
func RevokeAlbumAccess(db *gorm.DB, actor *models.User, albumID int, targetUserID int) error {
	var album models.Album
	if err := db.First(&album, albumID).Error; err != nil {
		return err
	}

	if !actor.Admin {
		grant, err := actor.EffectiveGrant(db, &album)
		if err != nil {
			return err
		}
		if grant == nil || grant.GrantedByUserID != nil {
			return errors.New("only the owner of a folder may revoke access to it")
		}
	}

	return models.RevokeAlbumLevel(db, albumID, targetUserID)
}
