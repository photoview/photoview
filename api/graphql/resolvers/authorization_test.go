package resolvers

import (
	"context"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// stubFaceDetector lets the face resolvers run without the dlib models. It only
// records the faces RecognizeUnlabeledFaces is told to report.
type stubFaceDetector struct {
	recognized []*models.ImageFace
}

func (*stubFaceDetector) ReloadFacesFromDatabase(*gorm.DB) error    { return nil }
func (*stubFaceDetector) DetectFaces(*gorm.DB, *models.Media) error { return nil }
func (*stubFaceDetector) MergeCategories(int32, int32)              {}
func (*stubFaceDetector) MergeImageFaces([]int, int32)              {}
func (s *stubFaceDetector) RecognizeUnlabeledFaces(*gorm.DB, *models.User) ([]*models.ImageFace, error) {
	return s.recognized, nil
}

func useStubFaceDetector(t *testing.T) *stubFaceDetector {
	t.Helper()

	previous := face_detection.GlobalFaceDetector
	detector := &stubFaceDetector{}
	face_detection.GlobalFaceDetector = detector
	t.Cleanup(func() { face_detection.GlobalFaceDetector = previous })

	return detector
}

// twoTenantFixtures describes two users that each own one album with two media items.
type twoTenantFixtures struct {
	db         *gorm.DB
	resolver   *Resolver
	alice, bob *models.User
	aliceMedia []*models.Media
	bobMedia   []*models.Media
}

func setupTwoTenantFixtures(t *testing.T) twoTenantFixtures {
	t.Helper()

	db := test_utils.DatabaseTest(t)
	require.NoError(t, db.AutoMigrate(&models.ImageFace{}, &models.FaceGroup{}))

	password := "1234"
	alice, err := models.RegisterUser(db, "alice", &password, false)
	require.NoError(t, err)
	bob, err := models.RegisterUser(db, "bob", &password, false)
	require.NoError(t, err)

	newMedia := func(owner *models.User, name string) []*models.Media {
		album := models.Album{Title: name + " album", Path: "/photos/" + name}
		require.NoError(t, db.Save(&album).Error)
		require.NoError(t, db.Model(owner).Association("Albums").Append(&album))

		items := make([]*models.Media, 0, 2)
		for _, file := range []string{"first.jpg", "second.jpg"} {
			media := models.Media{Title: name + " " + file, Path: album.Path + "/" + file, AlbumID: album.ID}
			require.NoError(t, db.Save(&media).Error)
			url := models.MediaURL{MediaID: media.ID, MediaName: file, Purpose: models.MediaOriginal, ContentType: "image/jpeg"}
			require.NoError(t, db.Save(&url).Error)
			items = append(items, &media)
		}
		return items
	}

	return twoTenantFixtures{
		db:         db,
		resolver:   &Resolver{database: db},
		alice:      alice,
		bob:        bob,
		aliceMedia: newMedia(alice, "alice"),
		bobMedia:   newMedia(bob, "bob"),
	}
}

func (f twoTenantFixtures) ctxFor(user *models.User) context.Context {
	if user == nil {
		return context.Background()
	}
	return auth.AddUserToContext(context.Background(), user)
}

func (f twoTenantFixtures) addFace(t *testing.T, media *models.Media, group *models.FaceGroup) *models.ImageFace {
	t.Helper()

	face := models.ImageFace{FaceGroupID: group.ID, MediaID: media.ID}
	require.NoError(t, f.db.Create(&face).Error)
	return &face
}

func (f twoTenantFixtures) addFaceGroup(t *testing.T, label *string) *models.FaceGroup {
	t.Helper()

	group := models.FaceGroup{Label: label}
	require.NoError(t, f.db.Create(&group).Error)
	return &group
}

// A media share token only grants access to that media item. The `album` field
// on it must not hand the whole album, its sibling media and its sub albums to a
// caller that does not own the album.
func TestMediaAlbumRequiresAlbumOwnership(t *testing.T) {
	f := setupTwoTenantFixtures(t)
	resolver := &mediaResolver{f.resolver}
	shared := f.aliceMedia[0]

	tests := []struct {
		name    string
		user    *models.User
		allowed bool
	}{
		{"anonymous share token holder", nil, false},
		{"user owning another album", f.bob, false},
		{"album owner", f.alice, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			album, err := resolver.Album(f.ctxFor(tt.user), shared)

			if !tt.allowed {
				assert.ErrorIs(t, err, auth.ErrUnauthorized)
				assert.Nil(t, album)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, album)
			assert.Equal(t, shared.AlbumID, album.ID)
		})
	}
}

// Face groups carry the names users gave to the people in their photos, so the
// faces of a media item are only listed to the owner of its album.
func TestMediaFacesOnlyVisibleToAlbumOwner(t *testing.T) {
	useStubFaceDetector(t)
	f := setupTwoTenantFixtures(t)
	resolver := &mediaResolver{f.resolver}

	label := "Carol"
	group := f.addFaceGroup(t, &label)
	f.addFace(t, f.aliceMedia[0], group)

	for _, tt := range []struct {
		name string
		user *models.User
		want int
	}{
		{"anonymous share token holder", nil, 0},
		{"user owning another album", f.bob, 0},
		{"album owner", f.alice, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			faces, err := resolver.Faces(f.ctxFor(tt.user), f.aliceMedia[0])
			require.NoError(t, err)
			assert.Len(t, faces, tt.want)
		})
	}
}
