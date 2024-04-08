package actions_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/graphql/models/actions"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestDeleteUser(t *testing.T) {
	t.Run("Delete regular user", func(t *testing.T) {
		db := test_utils.DatabaseTest(t)
		ids := make([]int, 0)
		db.Model(&models.Role{}).Where("name in (?, ?)", "ADMIN", "USER").Pluck("id", &ids)

		adminUser, err := models.RegisterUser(db, "admin", nil, ids[0])
		assert.NoError(t, err)

		regularUser, err := models.RegisterUser(db, "regular", nil, ids[1])
		assert.NoError(t, err)

		var dbUsers []*models.User
		err = db.Model(models.User{}).Find(&dbUsers).Error
		assert.NoError(t, err)
		assert.Len(t, dbUsers, 2)

		deletedUser, err := actions.DeleteUser(db, regularUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, regularUser.ID, deletedUser.ID)

		err = db.Model(models.User{}).Find(&dbUsers).Error
		assert.NoError(t, err)
		assert.Len(t, dbUsers, 1)
		assert.Equal(t, adminUser.ID, dbUsers[0].ID)
	})

	t.Run("Try to delete sole admin user", func(t *testing.T) {
		db := test_utils.DatabaseTest(t)
		ids := make([]int, 0)
		db.Model(&models.Role{}).Where("name in (?, ?)", "ADMIN", "USER").Pluck("id", &ids)
		adminUser, err := models.RegisterUser(db, "admin", nil, ids[0])
		assert.NoError(t, err)

		_, err = models.RegisterUser(db, "regular", nil, ids[1])
		assert.NoError(t, err)

		var dbUsers []*models.User
		err = db.Model(models.User{}).Find(&dbUsers).Error
		assert.NoError(t, err)
		assert.Len(t, dbUsers, 2)

		_, err = actions.DeleteUser(db, adminUser.ID)
		assert.Error(t, err)

		err = db.Model(models.User{}).Find(&dbUsers).Error
		assert.NoError(t, err)
		assert.Len(t, dbUsers, 2)
	})

	t.Run("Delete admin user when multiple admins exist", func(t *testing.T) {
		db := test_utils.DatabaseTest(t)

		ids := make([]int, 0)
		db.Model(&models.Role{}).Where("name = ?", "ADMIN").Pluck("id", &ids)
		adminUser1, err := models.RegisterUser(db, "admin", nil, ids[0])
		assert.NoError(t, err)

		adminUser2, err := models.RegisterUser(db, "another_admin", nil, ids[0])
		assert.NoError(t, err)

		var dbUsers []*models.User
		err = db.Model(models.User{}).Find(&dbUsers).Error
		assert.NoError(t, err)
		assert.Len(t, dbUsers, 2)

		deletedUser, err := actions.DeleteUser(db, adminUser1.ID)
		assert.NoError(t, err)
		assert.Equal(t, adminUser1.ID, deletedUser.ID)

		err = db.Model(models.User{}).Find(&dbUsers).Error
		assert.NoError(t, err)
		assert.Len(t, dbUsers, 1)
		assert.Equal(t, adminUser2.ID, dbUsers[0].ID)
	})
}
