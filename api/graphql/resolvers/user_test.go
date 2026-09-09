package resolvers

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserCanShare(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	r := &mutationResolver{Resolver: &Resolver{database: db}}
	ctx := context.Background()

	t.Run("defaults to false when omitted", func(t *testing.T) {
		user, err := r.CreateUser(ctx, "create_user_default", nil, false, nil, nil)
		assert.NoError(t, err)
		assert.False(t, user.CanShare)
	})

	t.Run("can be set to true at creation", func(t *testing.T) {
		canShare := true
		user, err := r.CreateUser(ctx, "create_user_can_share", nil, false, &canShare, nil)
		assert.NoError(t, err)
		assert.True(t, user.CanShare)
	})
}

func TestUpdateUserCanShare(t *testing.T) {
	db := test_utils.DatabaseTest(t)
	r := &mutationResolver{Resolver: &Resolver{database: db}}
	ctx := context.Background()

	user, err := models.RegisterUser(db, "update_user_can_share", nil, false)
	assert.NoError(t, err)
	assert.False(t, user.CanShare)

	t.Run("can be granted", func(t *testing.T) {
		canShare := true
		updated, err := r.UpdateUser(ctx, user.ID, nil, nil, nil, &canShare)
		assert.NoError(t, err)
		assert.True(t, updated.CanShare)
	})

	t.Run("can be revoked again", func(t *testing.T) {
		canShare := false
		updated, err := r.UpdateUser(ctx, user.ID, nil, nil, nil, &canShare)
		assert.NoError(t, err)
		assert.False(t, updated.CanShare)
	})

	t.Run("omitting it leaves the current value unchanged", func(t *testing.T) {
		canShare := true
		_, err := r.UpdateUser(ctx, user.ID, nil, nil, nil, &canShare)
		assert.NoError(t, err)

		username := "update_user_can_share_renamed"
		updated, err := r.UpdateUser(ctx, user.ID, &username, nil, nil, nil)
		assert.NoError(t, err)
		assert.True(t, updated.CanShare, "canShare must survive an update that doesn't mention it")
	})
}
