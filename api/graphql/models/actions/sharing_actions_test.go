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
	owner.CanShare = true
	assert.NoError(t, db.Save(owner).Error)

	recipient, err := models.RegisterUser(db, "recipient", nil, false)
	assert.NoError(t, err)
	recipient.CanShare = true
	assert.NoError(t, db.Save(recipient).Error)

	thirdParty, err := models.RegisterUser(db, "third_party", nil, false)
	assert.NoError(t, err)
	thirdParty.CanShare = true
	assert.NoError(t, db.Save(thirdParty).Error)

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

	t.Run("an owner without CanShare is refused even with full access", func(t *testing.T) {
		noShare, err := models.RegisterUser(db, "no_share_owner", nil, false)
		assert.NoError(t, err)
		assert.NoError(t, db.Create(&models.UserAlbums{
			UserID: noShare.ID, AlbumID: album.ID, Level: models.AlbumPermissionLevelDelete, GrantedByUserID: nil,
		}).Error)

		_, err = actions.GrantAlbumAccess(db, noShare, album.ID, thirdParty.ID, models.AlbumPermissionLevelRead)
		assert.Error(t, err)

		noShare.CanShare = true
		assert.NoError(t, db.Save(noShare).Error)

		_, err = actions.GrantAlbumAccess(db, noShare, album.ID, thirdParty.ID, models.AlbumPermissionLevelRead)
		assert.NoError(t, err, "granting CanShare should unlock sharing without any other change")
	})
}

func TestAlbumPermissions_ExcludesViewer(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "perms_owner", nil, false)
	assert.NoError(t, err)
	owner.CanShare = true
	assert.NoError(t, db.Save(owner).Error)
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
	owner.CanShare = true
	assert.NoError(t, db.Save(owner).Error)
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

func TestGrantAlbumAccess_CoexistsWithForeignGrantInSubtree(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "owner4", nil, false)
	assert.NoError(t, err)
	owner.CanShare = true
	assert.NoError(t, db.Save(owner).Error)
	recipient, err := models.RegisterUser(db, "recipient4", nil, false)
	assert.NoError(t, err)

	root := models.Album{Title: "root4", Path: "/photos/root4"}
	assert.NoError(t, db.Save(&root).Error)
	child := models.Album{Title: "child4", Path: "/photos/root4/child", ParentAlbumID: &root.ID}
	assert.NoError(t, db.Save(&child).Error)

	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: owner.ID, AlbumID: root.ID, Level: models.AlbumPermissionLevelDelete,
	}).Error)

	t.Run("re-sharing to the same recipient updates the level", func(t *testing.T) {
		_, err := actions.GrantAlbumAccess(db, owner, root.ID, recipient.ID, models.AlbumPermissionLevelRead)
		assert.NoError(t, err)

		_, err = actions.GrantAlbumAccess(db, owner, root.ID, recipient.ID, models.AlbumPermissionLevelUpload)
		assert.NoError(t, err)

		grant, err := recipient.EffectiveGrant(db, &root)
		assert.NoError(t, err)
		if assert.NotNil(t, grant) {
			assert.Equal(t, models.AlbumPermissionLevelUpload, grant.Level)
		}

		assert.NoError(t, actions.RevokeAlbumAccess(db, owner, root.ID, recipient.ID))
	})

	t.Run("owner sharing a subtree coexists with an independent grant on a descendant instead of overwriting it", func(t *testing.T) {
		// An admin already granted recipient Upload access directly on
		// child, independently of anything owner does at root.
		assert.NoError(t, models.PropagateAlbumLevel(db, child.ID, recipient.ID, models.AlbumPermissionLevelUpload, nil))
		t.Cleanup(func() {
			assert.NoError(t, models.RevokeAlbumLevel(db, child.ID, recipient.ID))
		})

		_, err := actions.GrantAlbumAccess(db, owner, root.ID, recipient.ID, models.AlbumPermissionLevelRead)
		assert.NoError(t, err)

		grant, err := recipient.EffectiveGrant(db, &child)
		assert.NoError(t, err)
		if assert.NotNil(t, grant) {
			assert.Equal(t, models.AlbumPermissionLevelUpload, grant.Level, "the higher of the two independent sources should win")
			assert.Nil(t, grant.GrantedByUserID, "an owner-rooted source keeps the effective grant owner-rooted")
		}

		// Revoking owner's share must leave the admin's independent grant
		// on child intact, since it traces to a different source.
		assert.NoError(t, actions.RevokeAlbumAccess(db, owner, root.ID, recipient.ID))

		grant, err = recipient.EffectiveGrant(db, &child)
		assert.NoError(t, err)
		if assert.NotNil(t, grant, "the admin's grant on the descendant must survive the owner's revoke") {
			assert.Equal(t, models.AlbumPermissionLevelUpload, grant.Level)
			assert.Nil(t, grant.GrantedByUserID)
		}
	})
}

func TestRevokeAlbumAccess_LeavesForeignGrantInSubtreeIntact(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "owner5", nil, false)
	assert.NoError(t, err)
	owner.CanShare = true
	assert.NoError(t, db.Save(owner).Error)
	recipient, err := models.RegisterUser(db, "recipient5", nil, false)
	assert.NoError(t, err)

	root := models.Album{Title: "root5", Path: "/photos/root5"}
	assert.NoError(t, db.Save(&root).Error)
	child := models.Album{Title: "child5", Path: "/photos/root5/child", ParentAlbumID: &root.ID}
	assert.NoError(t, db.Save(&child).Error)

	assert.NoError(t, db.Create(&models.UserAlbums{
		UserID: owner.ID, AlbumID: root.ID, Level: models.AlbumPermissionLevelDelete,
	}).Error)

	_, err = actions.GrantAlbumAccess(db, owner, root.ID, recipient.ID, models.AlbumPermissionLevelRead)
	assert.NoError(t, err)

	// A different, independent source (e.g. an admin) also grants recipient
	// access directly on the descendant - unrelated to owner's share.
	assert.NoError(t, models.PropagateAlbumLevel(db, child.ID, recipient.ID, models.AlbumPermissionLevelUpload, nil))

	assert.NoError(t, actions.RevokeAlbumAccess(db, owner, root.ID, recipient.ID))

	grant, err := recipient.EffectiveGrant(db, &child)
	assert.NoError(t, err)
	if assert.NotNil(t, grant, "the other source's grant on the descendant must be untouched") {
		assert.Equal(t, models.AlbumPermissionLevelUpload, grant.Level)
	}

	rootGrant, err := recipient.EffectiveGrant(db, &root)
	assert.NoError(t, err)
	assert.Nil(t, rootGrant, "the revoked source's own access to root should be gone")
}

func TestRevokeAlbumAccess(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	owner, err := models.RegisterUser(db, "owner3", nil, false)
	assert.NoError(t, err)
	owner.CanShare = true
	assert.NoError(t, db.Save(owner).Error)
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
