package resolvers

import (
	"strings"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

	got, err := userOwnedFaceGroup(f.ctxFor(f.alice), f.db, f.alice, own.ID)
	require.NoError(t, err)
	assert.Equal(t, own.ID, got.ID)

	got, err = userOwnedFaceGroup(f.ctxFor(f.alice), f.db, f.alice, foreign.ID)
	assert.Error(t, err)
	assert.Nil(t, got)

	got, err = userOwnedFaceGroup(f.ctxFor(f.alice), f.db, f.alice, foreign.ID+100)
	assert.Error(t, err)
	assert.Nil(t, got)

	got, err = userOwnedFaceGroup(f.ctxFor(admin), f.db, admin, foreign.ID)
	require.NoError(t, err)
	assert.Equal(t, foreign.ID, got.ID)

	got, err = userOwnedFaceGroup(f.ctxFor(admin), f.db, admin, foreign.ID+100)
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

// The lookup is sent in batches, as a single query with every id would exceed
// the number of bound parameters SQLite accepts.
func TestGetUserOwnedImageFacesWithManyIDs(t *testing.T) {
	f := setupTwoTenantFixtures(t)

	group := f.addFaceGroup(t, nil)
	aliceFace := f.addFace(t, f.aliceMedia[0], group)
	bobFace := f.addFace(t, f.bobMedia[0], group)

	ids := make([]int, 0, 40000)
	for id := 100; id < 40100; id++ {
		ids = append(ids, id)
	}
	ids = append(ids, aliceFace.ID, bobFace.ID)

	faces, err := getUserOwnedImageFaces(f.db, f.alice, ids)
	require.NoError(t, err)
	require.Len(t, faces, 1)
	assert.Equal(t, aliceFace.ID, faces[0].ID)

	admin := &models.User{Model: models.Model{ID: 1000}, Admin: true}
	faces, err = getUserOwnedImageFaces(f.db, admin, ids)
	require.NoError(t, err)
	assert.Len(t, faces, 2)
}

// Repeated IDs used to be returned once by a single IN query. Batching must
// keep that, or a later statement binds the repeated list twice.
func TestGetUserOwnedImageFacesDedupesAcrossBatches(t *testing.T) {
	f := setupTwoTenantFixtures(t)

	group := f.addFaceGroup(t, nil)
	aliceFace := f.addFace(t, f.aliceMedia[0], group)
	bobFace := f.addFace(t, f.bobMedia[0], group)

	ids := make([]int, 0, imageFaceLookupBatchSize+2)
	for range imageFaceLookupBatchSize {
		ids = append(ids, aliceFace.ID)
	}
	ids = append(ids, bobFace.ID, aliceFace.ID)

	faces, err := getUserOwnedImageFaces(f.db, f.alice, ids)
	require.NoError(t, err)
	require.Len(t, faces, 1)
	assert.Equal(t, aliceFace.ID, faces[0].ID)
}

// Album membership is a subquery, so owning more albums does not add a bound
// parameter per album on top of the face-ID batch.
func TestGetUserOwnedImageFacesBoundsAlbumParameters(t *testing.T) {
	f := setupTwoTenantFixtures(t)
	for i := range 4 {
		album := models.Album{Title: "extra", Path: "/photos/alice-extra-" + string(rune('a'+i))}
		require.NoError(t, f.db.Save(&album).Error)
		require.NoError(t, f.db.Model(f.alice).Association("Albums").Append(&album))
	}

	group := f.addFaceGroup(t, nil)
	aliceFace := f.addFace(t, f.aliceMedia[0], group)
	bobFace := f.addFace(t, f.bobMedia[0], group)

	ids := make([]int, 0, imageFaceLookupBatchSize+1)
	for id := 100000; id < 100000+imageFaceLookupBatchSize; id++ {
		ids = append(ids, id)
	}
	ids[0] = bobFace.ID
	ids = append(ids, aliceFace.ID)

	var boundCounts []int
	var sawMembershipSubquery bool
	const callbackName = "test:face-lookup-params"
	_ = f.db.Callback().Query().Remove(callbackName)
	require.NoError(t, f.db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		sql := tx.Statement.SQL.String()
		if !strings.Contains(sql, "image_faces") {
			return
		}
		boundCounts = append(boundCounts, len(tx.Statement.Vars))
		if strings.Contains(strings.ToLower(sql), "user_albums") {
			sawMembershipSubquery = true
		}
	}))
	t.Cleanup(func() { _ = f.db.Callback().Query().Remove(callbackName) })

	faces, err := getUserOwnedImageFaces(f.db, f.alice, ids)
	require.NoError(t, err)
	require.Len(t, faces, 1)
	assert.Equal(t, aliceFace.ID, faces[0].ID)
	assert.True(t, sawMembershipSubquery, "album membership must be a user_albums subquery")
	assert.Equal(t, []int{imageFaceLookupBatchSize + 1, 2}, boundCounts)
}

// recognizeUnlabeledFaces exposes each face's media. Administrators see the
// full result without a user_albums row. A regular user still only sees faces
// in albums they own. Move and detach keep the same administrator exception.
func TestRecognizeUnlabeledFacesAllowsAdministratorWithoutAlbumMembership(t *testing.T) {
	detector := useStubFaceDetector(t)
	f := setupTwoTenantFixtures(t)

	password := "1234"
	admin, err := models.RegisterUser(f.db, "admin", &password, true)
	require.NoError(t, err)

	group := f.addFaceGroup(t, nil)
	aliceFace := f.addFace(t, f.aliceMedia[0], group)
	bobFace := f.addFace(t, f.bobMedia[0], group)
	detector.recognized = []*models.ImageFace{aliceFace, bobFace}

	faces, err := (&mutationResolver{f.resolver}).RecognizeUnlabeledFaces(f.ctxFor(admin))
	require.NoError(t, err)
	ids := make([]int, len(faces))
	for i, face := range faces {
		ids[i] = face.ID
	}
	assert.Equal(t, []int{aliceFace.ID, bobFace.ID}, ids)

	faces, err = (&mutationResolver{f.resolver}).RecognizeUnlabeledFaces(f.ctxFor(f.bob))
	require.NoError(t, err)
	ids = make([]int, len(faces))
	for i, face := range faces {
		ids[i] = face.ID
	}
	assert.Equal(t, []int{bobFace.ID}, ids)

	owned, err := getUserOwnedImageFaces(f.db, admin, []int{aliceFace.ID, bobFace.ID})
	require.NoError(t, err)
	assert.Len(t, owned, 2)
}

// A websocket connection keeps one user value. FillAlbums leaves User.Albums
// populated, so a later operation on that connection used to authorize face
// groups after the album row was removed. Each operation loads membership again.
func TestFaceGroupPathsRefreshMembershipAfterRevocation(t *testing.T) {
	useStubFaceDetector(t)
	f := setupTwoTenantFixtures(t)

	group := f.addFaceGroup(t, nil)
	face := f.addFace(t, f.aliceMedia[0], group)

	connection := f.ctxFor(f.alice)
	require.NoError(t, f.alice.FillAlbums(f.db))
	require.NotEmpty(t, f.alice.Albums)

	var membershipQueries int
	const callbackName = "test:face-group-membership"
	_ = f.db.Callback().Query().Remove(callbackName)
	require.NoError(t, f.db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if strings.Contains(strings.ToLower(tx.Statement.SQL.String()), "user_albums") {
			membershipQueries++
		}
	}))
	t.Cleanup(func() { _ = f.db.Callback().Query().Remove(callbackName) })

	first := auth.WithAlbumAccessCache(connection)
	owned, err := userOwnedFaceGroup(first, f.db, f.alice, group.ID)
	require.NoError(t, err)
	assert.Equal(t, group.ID, owned.ID)

	imageFaces, err := (&faceGroupResolver{f.resolver}).ImageFaces(first, group, nil)
	require.NoError(t, err)
	require.Len(t, imageFaces, 1)
	assert.Equal(t, face.ID, imageFaces[0].ID)

	count, err := (&faceGroupResolver{f.resolver}).ImageFaceCount(first, group)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	groups, err := (&queryResolver{f.resolver}).MyFaceGroups(first, nil)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, group.ID, groups[0].ID)

	gotGroup, err := (&queryResolver{f.resolver}).FaceGroup(first, group.ID)
	require.NoError(t, err)
	require.NotNil(t, gotGroup)
	assert.Equal(t, group.ID, gotGroup.ID)
	assert.Equal(t, 1, membershipQueries)

	// Delete the membership row directly. Association().Delete would also drop
	// the album from the in-memory slice, which hides the stale-list bug.
	require.NoError(t, f.db.Exec("DELETE FROM user_albums WHERE user_id = ?", f.alice.ID).Error)
	require.NotEmpty(t, f.alice.Albums, "the connection still holds the album list from before revocation")
	afterRemoval := membershipQueries

	next := auth.WithAlbumAccessCache(connection)
	owned, err = userOwnedFaceGroup(next, f.db, f.alice, group.ID)
	assert.Error(t, err)
	assert.Nil(t, owned)

	imageFaces, err = (&faceGroupResolver{f.resolver}).ImageFaces(next, group, nil)
	require.NoError(t, err)
	assert.Empty(t, imageFaces)

	count, err = (&faceGroupResolver{f.resolver}).ImageFaceCount(next, group)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	groups, err = (&queryResolver{f.resolver}).MyFaceGroups(next, nil)
	require.NoError(t, err)
	assert.Empty(t, groups)

	gotGroup, err = (&queryResolver{f.resolver}).FaceGroup(next, group.ID)
	require.NoError(t, err)
	require.NotNil(t, gotGroup)
	assert.NotEqual(t, group.ID, gotGroup.ID)

	label := "revoked"
	_, err = (&mutationResolver{f.resolver}).SetFaceGroupLabel(next, group.ID, &label)
	assert.Error(t, err)

	var reloaded models.FaceGroup
	require.NoError(t, f.db.First(&reloaded, group.ID).Error)
	assert.Nil(t, reloaded.Label)
	assert.Equal(t, afterRemoval+1, membershipQueries)

	// A regular second user still cannot label the group.
	_, err = (&mutationResolver{f.resolver}).SetFaceGroupLabel(f.ctxFor(f.bob), group.ID, &label)
	assert.Error(t, err)
}
