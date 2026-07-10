package resolvers

import (
	"context"
	"strings"
	"testing"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/test_utils"
)

func setupFaceMutationTest(t *testing.T) (*mutationResolver, context.Context, *models.User) {
	t.Helper()

	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)
	face_detection.InitializeFaceDetector(db)

	pass := "1234"
	user, err := models.RegisterUser(db, "test_user", &pass, true)
	if err != nil {
		t.Fatal("register user error:", err)
	}

	if err := db.AutoMigrate(&models.ImageFace{}, &models.FaceGroup{}, &models.Media{}, &models.Album{}); err != nil {
		t.Fatal("automigrate error:", err)
	}

	r := &mutationResolver{
		Resolver: &Resolver{
			database: db,
		},
	}

	return r, auth.AddUserToContext(context.Background(), user), user
}

func createFaceMutationFixtures(t *testing.T, r *mutationResolver) {
	t.Helper()

	db := r.database

	testAlbum := models.Album{Title: "Test Album", Path: "/test-album"}
	if err := db.Create(&testAlbum).Error; err != nil {
		t.Fatal(err)
	}

	testMedia := []models.Media{
		{Model: models.Model{ID: 1}, Path: "test1", AlbumID: testAlbum.ID},
		{Model: models.Model{ID: 2}, Path: "test2", AlbumID: testAlbum.ID},
		{Model: models.Model{ID: 3}, Path: "test3", AlbumID: testAlbum.ID},
		{Model: models.Model{ID: 4}, Path: "test4", AlbumID: testAlbum.ID},
	}
	if err := db.Create(&testMedia).Error; err != nil {
		t.Fatal(err)
	}

	testFaceGroups := []models.FaceGroup{
		{Model: models.Model{ID: 1}},
		{Model: models.Model{ID: 2}},
		{Model: models.Model{ID: 3}},
		{Model: models.Model{ID: 4}},
	}
	if err := db.Create(&testFaceGroups).Error; err != nil {
		t.Fatal(err)
	}
}

func setupNonAdminUserWithAlbum(t *testing.T, r *mutationResolver) (*models.User, context.Context) {
	t.Helper()
	db := r.database

	pass := "5678"
	user, err := models.RegisterUser(db, "test_nonadmin", &pass, false)
	if err != nil {
		t.Fatal("register non-admin user error:", err)
	}

	var album models.Album
	if err := db.Where("path = ?", "/test-album").First(&album).Error; err != nil {
		t.Fatal("find fixture album error:", err)
	}

	if err := db.Model(user).Association("Albums").Append(&album); err != nil {
		t.Fatal("link non-admin user to album error:", err)
	}

	return user, auth.AddUserToContext(context.Background(), user)
}

func TestCombineFaceGroups(t *testing.T) {
	t.Run("merges face groups when media does not overlap", func(t *testing.T) {
		r, ctx, _ := setupFaceMutationTest(t)
		createFaceMutationFixtures(t, r)

		testDataList := []models.ImageFace{
			{Model: models.Model{ID: 1}, FaceGroupID: 1, MediaID: 1},
			{Model: models.Model{ID: 2}, FaceGroupID: 1, MediaID: 2},
			{Model: models.Model{ID: 3}, FaceGroupID: 2, MediaID: 3},
			{Model: models.Model{ID: 4}, FaceGroupID: 3, MediaID: 4},
		}
		if err := r.database.Create(&testDataList).Error; err != nil {
			t.Fatal(err)
		}

		result, err := r.CombineFaceGroups(ctx, 1, []int{2, 3})
		if err != nil {
			t.Fatal("CombineFaceGroups failed:", err)
		}
		if result == nil || result.ID != 1 {
			t.Fatalf("expected destination face group 1, got %#v", result)
		}

		var imageFaces []*models.ImageFace
		if err := r.database.Where("face_group_id = ?", 1).Find(&imageFaces).Error; err != nil {
			t.Fatal("failed to query destination image faces:", err)
		}

		if len(imageFaces) != 4 {
			t.Fatalf("expected 4 image faces after merge, got %d", len(imageFaces))
		}
	})

	t.Run("rejects merge when destination and source contain the same media", func(t *testing.T) {
		r, ctx, _ := setupFaceMutationTest(t)
		createFaceMutationFixtures(t, r)

		testDataList := []models.ImageFace{
			{Model: models.Model{ID: 1}, FaceGroupID: 1, MediaID: 1},
			{Model: models.Model{ID: 2}, FaceGroupID: 1, MediaID: 2},
			{Model: models.Model{ID: 3}, FaceGroupID: 2, MediaID: 2},
			{Model: models.Model{ID: 4}, FaceGroupID: 2, MediaID: 3},
		}
		if err := r.database.Create(&testDataList).Error; err != nil {
			t.Fatal(err)
		}

		result, err := r.CombineFaceGroups(ctx, 1, []int{2})
		if err == nil {
			t.Fatal("expected CombineFaceGroups to reject duplicate media, got nil error")
		}
		if result != nil {
			t.Fatalf("expected nil result when merge is rejected, got %#v", result)
		}
		if !strings.Contains(err.Error(), "duplicate images") {
			t.Fatalf("expected duplicate image validation error, got: %v", err)
		}

		var sourceCount int64
		if err := r.database.Model(&models.ImageFace{}).Where("face_group_id = ?", 2).Count(&sourceCount).Error; err != nil {
			t.Fatal(err)
		}
		if sourceCount != 2 {
			t.Fatalf("expected source group image faces to remain unchanged, got %d", sourceCount)
		}
	})

	t.Run("rejects merge when multiple sources contain the same media", func(t *testing.T) {
		r, ctx, _ := setupFaceMutationTest(t)
		createFaceMutationFixtures(t, r)

		testDataList := []models.ImageFace{
			{Model: models.Model{ID: 1}, FaceGroupID: 1, MediaID: 1},
			{Model: models.Model{ID: 2}, FaceGroupID: 2, MediaID: 2},
			{Model: models.Model{ID: 3}, FaceGroupID: 3, MediaID: 2},
		}
		if err := r.database.Create(&testDataList).Error; err != nil {
			t.Fatal(err)
		}

		result, err := r.CombineFaceGroups(ctx, 1, []int{2, 3})
		if err == nil {
			t.Fatal("expected CombineFaceGroups to reject duplicate media across sources, got nil error")
		}
		if result != nil {
			t.Fatalf("expected nil result when merge is rejected, got %#v", result)
		}
		if !strings.Contains(err.Error(), "duplicate images") {
			t.Fatalf("expected duplicate image validation error, got: %v", err)
		}
	})

	t.Run("non-admin user merges owned face groups", func(t *testing.T) {
		r, _, _ := setupFaceMutationTest(t)
		createFaceMutationFixtures(t, r)
		_, ctx := setupNonAdminUserWithAlbum(t, r)

		testDataList := []models.ImageFace{
			{Model: models.Model{ID: 1}, FaceGroupID: 1, MediaID: 1},
			{Model: models.Model{ID: 2}, FaceGroupID: 1, MediaID: 2},
			{Model: models.Model{ID: 3}, FaceGroupID: 2, MediaID: 3},
			{Model: models.Model{ID: 4}, FaceGroupID: 3, MediaID: 4},
		}
		if err := r.database.Create(&testDataList).Error; err != nil {
			t.Fatal(err)
		}

		result, err := r.CombineFaceGroups(ctx, 1, []int{2, 3})
		if err != nil {
			t.Fatal("CombineFaceGroups failed for non-admin user:", err)
		}
		if result == nil || result.ID != 1 {
			t.Fatalf("expected destination face group 1, got %#v", result)
		}

		var imageFaces []*models.ImageFace
		if err := r.database.Where("face_group_id = ?", 1).Find(&imageFaces).Error; err != nil {
			t.Fatal("failed to query destination image faces:", err)
		}
		if len(imageFaces) != 4 {
			t.Fatalf("expected 4 image faces after non-admin merge, got %d", len(imageFaces))
		}
	})
}

func TestMoveImageFaces(t *testing.T) {
	t.Run("moves image faces when destination does not contain selected media", func(t *testing.T) {
		r, ctx, _ := setupFaceMutationTest(t)
		createFaceMutationFixtures(t, r)

		testDataList := []models.ImageFace{
			{Model: models.Model{ID: 1}, FaceGroupID: 1, MediaID: 1},
			{Model: models.Model{ID: 2}, FaceGroupID: 2, MediaID: 2},
		}
		if err := r.database.Create(&testDataList).Error; err != nil {
			t.Fatal(err)
		}

		result, err := r.MoveImageFaces(ctx, []int{1}, 2)
		if err != nil {
			t.Fatal("MoveImageFaces failed:", err)
		}
		if result == nil || result.ID != 2 {
			t.Fatalf("expected destination face group 2, got %#v", result)
		}

		var movedFace models.ImageFace
		if err := r.database.First(&movedFace, 1).Error; err != nil {
			t.Fatal(err)
		}
		if movedFace.FaceGroupID != 2 {
			t.Fatalf("expected image face to move to group 2, got group %d", movedFace.FaceGroupID)
		}
	})

	t.Run("rejects move when destination already contains selected media", func(t *testing.T) {
		r, ctx, _ := setupFaceMutationTest(t)
		createFaceMutationFixtures(t, r)

		testDataList := []models.ImageFace{
			{Model: models.Model{ID: 1}, FaceGroupID: 1, MediaID: 1},
			{Model: models.Model{ID: 2}, FaceGroupID: 2, MediaID: 1},
		}
		if err := r.database.Create(&testDataList).Error; err != nil {
			t.Fatal(err)
		}

		result, err := r.MoveImageFaces(ctx, []int{1}, 2)
		if err == nil {
			t.Fatal("expected MoveImageFaces to reject duplicate media, got nil error")
		}
		if result != nil {
			t.Fatalf("expected nil result when move is rejected, got %#v", result)
		}
		if !strings.Contains(err.Error(), "duplicate images") {
			t.Fatalf("expected duplicate image validation error, got: %v", err)
		}

		var originalFace models.ImageFace
		if err := r.database.First(&originalFace, 1).Error; err != nil {
			t.Fatal(err)
		}
		if originalFace.FaceGroupID != 1 {
			t.Fatalf("expected rejected move to leave image face in group 1, got group %d", originalFace.FaceGroupID)
		}
	})

	t.Run("rejects move when selected faces contain duplicate media", func(t *testing.T) {
		r, ctx, _ := setupFaceMutationTest(t)
		createFaceMutationFixtures(t, r)

		testDataList := []models.ImageFace{
			{Model: models.Model{ID: 1}, FaceGroupID: 1, MediaID: 1},
			{Model: models.Model{ID: 2}, FaceGroupID: 3, MediaID: 1},
			{Model: models.Model{ID: 3}, FaceGroupID: 2, MediaID: 2},
		}
		if err := r.database.Create(&testDataList).Error; err != nil {
			t.Fatal(err)
		}

		result, err := r.MoveImageFaces(ctx, []int{1, 2}, 2)
		if err == nil {
			t.Fatal("expected MoveImageFaces to reject duplicate selected media, got nil error")
		}
		if result != nil {
			t.Fatalf("expected nil result when move is rejected, got %#v", result)
		}
		if !strings.Contains(err.Error(), "duplicate images") {
			t.Fatalf("expected duplicate image validation error, got: %v", err)
		}
	})

	t.Run("non-admin user moves owned image faces", func(t *testing.T) {
		r, _, _ := setupFaceMutationTest(t)
		createFaceMutationFixtures(t, r)
		_, ctx := setupNonAdminUserWithAlbum(t, r)

		testDataList := []models.ImageFace{
			{Model: models.Model{ID: 1}, FaceGroupID: 1, MediaID: 1},
			{Model: models.Model{ID: 2}, FaceGroupID: 2, MediaID: 2},
		}
		if err := r.database.Create(&testDataList).Error; err != nil {
			t.Fatal(err)
		}

		result, err := r.MoveImageFaces(ctx, []int{1}, 2)
		if err != nil {
			t.Fatal("MoveImageFaces failed for non-admin user:", err)
		}
		if result == nil || result.ID != 2 {
			t.Fatalf("expected destination face group 2, got %#v", result)
		}

		var movedFace models.ImageFace
		if err := r.database.First(&movedFace, 1).Error; err != nil {
			t.Fatal(err)
		}
		if movedFace.FaceGroupID != 2 {
			t.Fatalf("expected image face to move to group 2, got group %d", movedFace.FaceGroupID)
		}
	})
}

func TestGetUserOwnedImageFaces(t *testing.T) {
	t.Run("returns empty when image face IDs list is empty", func(t *testing.T) {
		r, _, user := setupFaceMutationTest(t)

		faces, err := getUserOwnedImageFaces(r.database, user, []int{})
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
		if len(faces) != 0 {
			t.Fatalf("expected 0 faces for empty input, got %d", len(faces))
		}
	})

	t.Run("non-admin user with no album associations returns empty", func(t *testing.T) {
		r, _, _ := setupFaceMutationTest(t)
		createFaceMutationFixtures(t, r)

		testDataList := []models.ImageFace{
			{Model: models.Model{ID: 1}, FaceGroupID: 1, MediaID: 1},
		}
		if err := r.database.Create(&testDataList).Error; err != nil {
			t.Fatal(err)
		}

		pass := "5678"
		userNoAlbums, err := models.RegisterUser(r.database, "test_noalbums", &pass, false)
		if err != nil {
			t.Fatal("register user error:", err)
		}

		faces, err := getUserOwnedImageFaces(r.database, userNoAlbums, []int{1})
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
		if len(faces) != 0 {
			t.Fatalf("expected 0 faces for user with no albums, got %d", len(faces))
		}
	})
}
