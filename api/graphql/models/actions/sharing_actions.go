package actions

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// AlbumPermissions lists every user *other than viewerID* with an explicit
// access grant directly on albumID - i.e. who this album has been shared
// with, not including the viewer's own (usually admin-granted) access to
// it. Callers must already have verified the viewer is allowed to see this
// (the album's owner, or an admin) - this function itself does no
// authorization check.
func AlbumPermissions(db *gorm.DB, albumID int, viewerID int) ([]*models.AlbumPermission, error) {
	var rows []models.UserAlbums
	if err := db.Where("album_id = ? AND user_id != ?", albumID, viewerID).Find(&rows).Error; err != nil {
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

// targetHasForeignGrantInSubtree reports whether targetUserID already holds
// a grant, somewhere in albumID's current subtree (albumID included), that
// doesn't trace back to actorID - e.g. an admin's direct root grant on a
// nested share, or a different owner's independent share of a descendant
// folder. UserAlbums has exactly one row per (user, album), so
// PropagateAlbumLevel/RevokeAlbumLevel would silently overwrite or delete
// such a row; callers use this to refuse the operation instead.
func targetHasForeignGrantInSubtree(db *gorm.DB, albumID int, targetUserID int, actorID int) (bool, error) {
	var album models.Album
	if err := db.First(&album, albumID).Error; err != nil {
		return false, err
	}

	subtree, err := album.GetChildren(db, nil)
	if err != nil {
		return false, err
	}

	ids := make([]int, len(subtree))
	for i, a := range subtree {
		ids[i] = a.ID
	}

	var count int64
	err = db.Model(&models.UserAlbums{}).
		Where("user_id = ? AND album_id IN (?)", targetUserID, ids).
		Where("granted_by_user_id IS NULL OR granted_by_user_id != ?", actorID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ShareAlbum grants targetUserID access to albumID at level, on behalf of
// actor. Only albumID's owner (an explicit grant with no GrantedByUserID,
// i.e. it traces to an admin grant rather than another user's share) may
// do this, and only up to their own level - admins bypass both checks.
// Re-sharing an already-shared album updates the level and re-propagates it
// across the album's current subtree, so already-scanned descendants never
// keep a stale level.
func GrantAlbumAccess(db *gorm.DB, actor *models.User, albumID int, targetUserID int, level models.AlbumPermissionLevel) (*models.AlbumPermission, error) {
	if actor.ID == targetUserID {
		return nil, errors.New("cannot share an album with yourself")
	}

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

		hasForeign, err := targetHasForeignGrantInSubtree(db, albumID, targetUserID, actor.ID)
		if err != nil {
			return nil, err
		}
		if hasForeign {
			return nil, errors.New("target user already has access to part of this folder from another source")
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
	if actor.ID == targetUserID {
		return errors.New("cannot revoke your own access")
	}

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

		hasForeign, err := targetHasForeignGrantInSubtree(db, albumID, targetUserID, actor.ID)
		if err != nil {
			return err
		}
		if hasForeign {
			return errors.New("cannot revoke access that was granted by someone else")
		}
	}

	return models.RevokeAlbumLevel(db, albumID, targetUserID)
}
