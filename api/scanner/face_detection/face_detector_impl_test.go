//go:build !no_face_detection

package face_detection

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

// descriptorAt returns a descriptor whose components all equal value. Two of
// them lie a squared distance of 128*(a-b)^2 apart, which is what go-face
// compares with the tolerance of 0.2.
func descriptorAt(value float32) models.FaceDescriptor {
	var d models.FaceDescriptor
	for i := range d {
		d[i] = value
	}
	return d
}

func TestMergeCategoriesUpdatesTheRecognizer(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	previous := GlobalFaceDetector
	t.Cleanup(func() { GlobalFaceDetector = previous })

	album := models.Album{Title: "album", Path: "/merge-categories"}
	if err := db.Create(&album).Error; err != nil {
		t.Fatal(err)
	}
	media := []models.Media{
		{Title: "one", Path: "/merge-categories/one.jpg", AlbumID: album.ID},
		{Title: "two", Path: "/merge-categories/two.jpg", AlbumID: album.ID},
	}
	if err := db.Create(&media).Error; err != nil {
		t.Fatal(err)
	}

	source, dest, unrelated := &models.FaceGroup{}, &models.FaceGroup{}, &models.FaceGroup{}
	for _, group := range []*models.FaceGroup{source, dest, unrelated} {
		if err := db.Create(group).Error; err != nil {
			t.Fatal(err)
		}
	}

	// 128 * 0.1^2 = 1.28 apart: far beyond the tolerance of 0.2.
	sourceFace, unrelatedFace := descriptorAt(0.05), descriptorAt(-0.05)
	faces := []models.ImageFace{
		{FaceGroupID: source.ID, MediaID: media[0].ID, Descriptor: sourceFace},
		{FaceGroupID: unrelated.ID, MediaID: media[1].ID, Descriptor: unrelatedFace},
	}
	if err := db.Create(&faces).Error; err != nil {
		t.Fatal(err)
	}

	if err := InitializeFaceDetector(db); err != nil {
		t.Fatal("initialize face detector error:", err)
	}
	detector, ok := GlobalFaceDetector.(*faceDetector)
	if !ok {
		t.Fatalf("GlobalFaceDetector is %T, want *faceDetector", GlobalFaceDetector)
	}

	classify := func(descriptor models.FaceDescriptor) int32 {
		detector.mutex.Lock()
		defer detector.mutex.Unlock()
		return detector.classifyDescriptor([128]float32(descriptor))
	}

	// Hand the samples over the way the scan of a new photo does, so the
	// recognizer starts out knowing the source group.
	detector.mutex.Lock()
	detector.rec.SetSamples(detector.faceDescriptors, detector.faceGroupIDs)
	detector.mutex.Unlock()
	if got := classify(sourceFace); got != int32(source.ID) {
		t.Fatalf("classify before merge = %d, want group %d", got, source.ID)
	}

	detector.MergeCategories(int32(source.ID), int32(dest.ID))

	if got := classify(sourceFace); got != int32(dest.ID) {
		t.Errorf("classify after merge = %d, want the destination group %d", got, dest.ID)
	}
	if got := classify(unrelatedFace); got != int32(unrelated.ID) {
		t.Errorf("classify of an unrelated face after merge = %d, want its group %d", got, unrelated.ID)
	}
}
