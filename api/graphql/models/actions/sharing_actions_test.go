package actions_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestGrantAlbumAccess_OwnerOnly(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "owner", nil, false)
	assert.NoError(t, err)

	recipient, err := models.RegisterUser(db, "recipient", nil, false)
	assert.NoError(t, err)

	thirdParty, err := models.RegisterUser(db, "third_party", nil, false)
	assert.NoError(t, err)

	admin, err := models.RegisterUser(db, "admin", nil, true)
	assert.NoError(t, err)

	album := models.Album{Title: "album", Path: "/photos/album"}
	assert.NoError(t, db.Save(&album).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: owner.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelDelete,
	}).Error)

	t.Run("owner can share up to their own level", func(t *testing.T) {
		permission, err := actions.GrantAlbumAccess(db, owner, album.ID, recipient.ID, models.AlbumPermissionLevelUpload)
		assert.NoError(t, err)
		assert.Equal(t, models.AlbumPermissionLevelUpload, permission.Level)

		grant, err := recipient.EffectiveGrant(db, &album)
		assert.NoError(t, err)
		if assert.NotNil(t, grant) {
			assert.Equal(t, models.AlbumPermissionLevelUpload, grant.Level)
			assert.NotNil(t, grant.GrantedByUserID)
			assert.Equal(t, owner.ID, *grant.GrantedByUserID)
		}
	})

	t.Run("owner cannot grant a higher level than their own", func(t *testing.T) {
		limitedAlbum := models.Album{Title: "limited_album", Path: "/photos/limited_album"}
		assert.NoError(t, db.Save(&limitedAlbum).Error)
		assert.NoError(t, db.Create(&models.UserAlbums{
			UserID: owner.ID, AlbumID: limitedAlbum.ID, Level: models.AlbumPermissionLevelUpload,
		}).Error)

		_, err := actions.GrantAlbumAccess(db, owner, limitedAlbum.ID, recipient.ID, models.AlbumPermissionLevelDelete)
		assert.Error(t, err)

		grant, err := recipient.EffectiveGrant(db, &limitedAlbum)
		assert.NoError(t, err)
		assert.Nil(t, grant, "the rejected grant must not have been created")
	})

	t.Run("a share recipient cannot re-share", func(t *testing.T) {
		_, err := actions.GrantAlbumAccess(db, recipient, album.ID, thirdParty.ID, models.AlbumPermissionLevelRead)
		assert.Error(t, err)

		grant, err := thirdParty.EffectiveGrant(db, &album)
		assert.NoError(t, err)
		assert.Nil(t, grant)
	})

	t.Run("a user with no access at all cannot share", func(t *testing.T) {
		strangerAlbum := models.Album{Title: "other", Path: "/photos/other"}
		assert.NoError(t, db.Save(&strangerAlbum).Error)

		_, err := actions.GrantAlbumAccess(db, thirdParty, strangerAlbum.ID, recipient.ID, models.AlbumPermissionLevelRead)
		assert.Error(t, err)
	})

	t.Run("admin bypasses both the ownership and level checks", func(t *testing.T) {
		permission, err := actions.GrantAlbumAccess(db, admin, album.ID, thirdParty.ID, models.AlbumPermissionLevelDelete)
		assert.NoError(t, err)
		assert.Equal(t, models.AlbumPermissionLevelDelete, permission.Level)
	})

	t.Run("cannot share an album with yourself", func(t *testing.T) {
		_, err := actions.GrantAlbumAccess(db, owner, album.ID, owner.ID, models.AlbumPermissionLevelRead)
		assert.Error(t, err)
	})
}

func TestAlbumPermissions_ExcludesViewer(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "perms_owner", nil, false)
	assert.NoError(t, err)
	recipient, err := models.RegisterUser(db, "perms_recipient", nil, false)
	assert.NoError(t, err)

	album := models.Album{Title: "perms_album", Path: "/photos/perms_album"}
	assert.NoError(t, db.Save(&album).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: owner.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelDelete,
	}).Error)

	_, err = actions.GrantAlbumAccess(db, owner, album.ID, recipient.ID, models.AlbumPermissionLevelRead)
	assert.NoError(t, err)

	permissions, err := actions.AlbumPermissions(db, album.ID, owner.ID)
	assert.NoError(t, err)
	if assert.Len(t, permissions, 1) {
		assert.Equal(t, recipient.ID, permissions[0].User.ID)
	}
}

func TestGrantAlbumAccess_PropagatesToExistingDescendants(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "owner2", nil, false)
	assert.NoError(t, err)
	recipient, err := models.RegisterUser(db, "recipient2", nil, false)
	assert.NoError(t, err)

	root := models.Album{Title: "root", Path: "/photos/root2"}
	assert.NoError(t, db.Save(&root).Error)
	child := models.Album{Title: "child", Path: "/photos/root2/child", ParentAlbumID: &root.ID}
	assert.NoError(t, db.Save(&child).Error)

	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: owner.ID, AlbumID: root.ID, Level: models.AlbumPermissionLevelDelete,
	}).Error)

	// child already existed on disk/in the DB before the share happens, so a
	// plain single-row insert on root wouldn't be enough for the recipient
	// to have access to it.
	_, err = actions.GrantAlbumAccess(db, owner, root.ID, recipient.ID, models.AlbumPermissionLevelRead)
	assert.NoError(t, err)

	grant, err := recipient.EffectiveGrant(db, &child)
	assert.NoError(t, err)
	if assert.NotNil(t, grant) {
		assert.Equal(t, models.AlbumPermissionLevelRead, grant.Level)
	}
}

func TestRevokeAlbumAccess(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "owner3", nil, false)
	assert.NoError(t, err)
	recipient, err := models.RegisterUser(db, "recipient3", nil, false)
	assert.NoError(t, err)
	nonOwner, err := models.RegisterUser(db, "non_owner3", nil, false)
	assert.NoError(t, err)

	album := models.Album{Title: "album3", Path: "/photos/album3"}
	assert.NoError(t, db.Save(&album).Error)
	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: owner.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelDelete,
	}).Error)

	_, err = actions.GrantAlbumAccess(db, owner, album.ID, recipient.ID, models.AlbumPermissionLevelUpload)
	assert.NoError(t, err)

	t.Run("a non-owner cannot revoke", func(t *testing.T) {
		err := actions.RevokeAlbumAccess(db, nonOwner, album.ID, recipient.ID)
		assert.Error(t, err)

		grant, err := recipient.EffectiveGrant(db, &album)
		assert.NoError(t, err)
		assert.NotNil(t, grant)
	})

	t.Run("the owner can revoke", func(t *testing.T) {
		err := actions.RevokeAlbumAccess(db, owner, album.ID, recipient.ID)
		assert.NoError(t, err)

		grant, err := recipient.EffectiveGrant(db, &album)
		assert.NoError(t, err)
		assert.Nil(t, grant)
	})

	t.Run("cannot revoke your own access", func(t *testing.T) {
		err := actions.RevokeAlbumAccess(db, owner, album.ID, owner.ID)
		assert.Error(t, err)

		grant, err := owner.EffectiveGrant(db, &album)
		assert.NoError(t, err)
		assert.NotNil(t, grant, "owner's own access must be untouched")
	})
}
