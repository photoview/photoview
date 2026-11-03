package resolvers

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A face group the caller has no image in must be treated as missing, rather
// than yielding an empty group without an error that callers could carry on with.
func TestUserOwnedFaceGroupRequiresOwnedImage(t *testing.T) {
	f := setupTwoTenantFixtures(t)
	admin := &models.User{Model: models.Model{ID: 1000}, Admin: true}

	own := f.addFaceGroup(t, nil)
	foreign := f.addFaceGroup(t, nil)
	f.addFace(t, f.aliceMedia[0], own)
	f.addFace(t, f.bobMedia[0], foreign)

	got, err := userOwnedFaceGroup(f.db, f.alice, own.ID)
	require.NoError(t, err)
	assert.Equal(t, own.ID, got.ID)

	got, err = userOwnedFaceGroup(f.db, f.alice, foreign.ID)
	assert.Error(t, err)
	assert.Nil(t, got)

	got, err = userOwnedFaceGroup(f.db, f.alice, foreign.ID+100)
	assert.Error(t, err)
	assert.Nil(t, got)

	got, err = userOwnedFaceGroup(f.db, admin, foreign.ID)
	require.NoError(t, err)
	assert.Equal(t, foreign.ID, got.ID)

	got, err = userOwnedFaceGroup(f.db, admin, foreign.ID+100)
	assert.Error(t, err)
	assert.Nil(t, got)
}

// The face group mutations act on the IDs of the request, so they must stop at
// face groups of other users.
func TestFaceGroupMutationsRejectForeignFaceGroups(t *testing.T) {
	useStubFaceDetector(t)
	mutation := func(f twoTenantFixtures) *mutationResolver { return &mutationResolver{f.resolver} }

	t.Run("combineFaceGroups source", func(t *testing.T) {
		f := setupTwoTenantFixtures(t)
		own := f.addFaceGroup(t, nil)
		foreign := f.addFaceGroup(t, nil)
		f.addFace(t, f.aliceMedia[0], own)
		bobFace := f.addFace(t, f.bobMedia[0], foreign)

		_, err := mutation(f).CombineFaceGroups(f.ctxFor(f.alice), own.ID, []int{foreign.ID})
		assert.Error(t, err)

		var reloaded models.ImageFace
		require.NoError(t, f.db.First(&reloaded, bobFace.ID).Error)
		assert.Equal(t, foreign.ID, reloaded.FaceGroupID, "the other user's face must stay in its group")
		var groups int64
		require.NoError(t, f.db.Model(&models.FaceGroup{}).Where("id = ?", foreign.ID).Count(&groups).Error)
		assert.EqualValues(t, 1, groups)
	})

	t.Run("combineFaceGroups destination", func(t *testing.T) {
		f := setupTwoTenantFixtures(t)
		own := f.addFaceGroup(t, nil)
		foreign := f.addFaceGroup(t, nil)
		aliceFace := f.addFace(t, f.aliceMedia[0], own)
		f.addFace(t, f.bobMedia[0], foreign)

		_, err := mutation(f).CombineFaceGroups(f.ctxFor(f.alice), foreign.ID, []int{own.ID})
		assert.Error(t, err)

		var reloaded models.ImageFace
		require.NoError(t, f.db.First(&reloaded, aliceFace.ID).Error)
		assert.Equal(t, own.ID, reloaded.FaceGroupID)
	})

	t.Run("moveImageFaces destination", func(t *testing.T) {
		f := setupTwoTenantFixtures(t)
		own := f.addFaceGroup(t, nil)
		foreign := f.addFaceGroup(t, nil)
		aliceFace := f.addFace(t, f.aliceMedia[0], own)
		f.addFace(t, f.bobMedia[0], foreign)

		_, err := mutation(f).MoveImageFaces(f.ctxFor(f.alice), []int{aliceFace.ID}, foreign.ID)
		assert.Error(t, err)

		var reloaded models.ImageFace
		require.NoError(t, f.db.First(&reloaded, aliceFace.ID).Error)
		assert.Equal(t, own.ID, reloaded.FaceGroupID)
	})

	t.Run("setFaceGroupLabel", func(t *testing.T) {
		f := setupTwoTenantFixtures(t)
		foreign := f.addFaceGroup(t, nil)
		f.addFace(t, f.bobMedia[0], foreign)
		label := "renamed"

		_, err := mutation(f).SetFaceGroupLabel(f.ctxFor(f.alice), foreign.ID, &label)
		assert.Error(t, err)

		var reloaded models.FaceGroup
		require.NoError(t, f.db.First(&reloaded, foreign.ID).Error)
		assert.Nil(t, reloaded.Label)
	})
}

// recognizeUnlabeledFaces reclassifies every face of the unlabeled groups the
// caller has a picture in. Faces of other users' pictures can be in those
// groups, and must not be reported back, since each carries its media.
func TestRecognizeUnlabeledFacesOnlyReturnsOwnFaces(t *testing.T) {
	detector := useStubFaceDetector(t)
	f := setupTwoTenantFixtures(t)

	group := f.addFaceGroup(t, nil)
	aliceFace := f.addFace(t, f.aliceMedia[0], group)
	bobFace := f.addFace(t, f.bobMedia[0], group)
	detector.recognized = []*models.ImageFace{aliceFace, bobFace}

	faces, err := (&mutationResolver{f.resolver}).RecognizeUnlabeledFaces(f.ctxFor(f.alice))
	require.NoError(t, err)

	ids := make([]int, len(faces))
	for i, face := range faces {
		ids[i] = face.ID
	}
	assert.Equal(t, []int{aliceFace.ID}, ids)
}
