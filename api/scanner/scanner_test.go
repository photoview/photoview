package scanner_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/test_utils"
	scanner_utils "github.com/photoview/photoview/api/test_utils/scanner"
)

func TestMain(m *testing.M) {
	os.Exit(test_utils.IntegrationTestRun(m))
}

func TestFullScan(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	pass := "1234"
	user, err := models.RegisterUser(db, "test_user", &pass, true)
	if err != nil {
		t.Fatal("register user error:", err)
	}

	rootAlbum := models.Album{
		Title: "root album",
		Path:  "./test_media/library",
	}

	wantImages := []string{
		"buttercup_close_summer_yellow.jpg",
		"lilac_lilac_bush_lilac.jpg",
		"mount_merapi_volcano_indonesia.jpg",
		"boy1.jpg",
		"boy2.jpg",
		"girl_black_hair2.jpg",
		"girl_blond1.jpg",
		"girl_blond2.jpg",
		"girl_blond3.jpg",
	}
	wantFaceGroups := [][]string{
		{"boy1.jpg", "boy2.jpg"},
		{"girl_black_hair2.jpg"},
		{"girl_blond1.jpg", "girl_blond2.jpg", "girl_blond3.jpg"},
	}

	for i := range wantFaceGroups {
		slices.Sort(wantFaceGroups[i])
	}
	slices.SortFunc(wantFaceGroups, func(a, b []string) int {
		return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
	})

	if err := db.Save(&rootAlbum).Error; err != nil {
		t.Fatal("create root album error:", err)
	}

	if err := db.Model(user).Association("Albums").Append(&rootAlbum); err != nil {
		t.Fatal("bind root album error:", err)
	}

	if err := face_detection.InitializeFaceDetector(db); err != nil {
		t.Fatal("initalize face detector error:", err)
	}

	scanner_utils.RunScannerOnUser(t, db, user)

	t.Run("CheckMedia", func(t *testing.T) {
		var allMedia []*models.Media
		if err := db.Find(&allMedia).Error; err != nil {
			t.Fatal("get all media error:", err)
		}

		want := slices.Clone(wantImages)
		slices.Sort(want)

		got := make([]string, len(allMedia))
		for i, media := range allMedia {
			got[i] = media.Title
		}
		slices.Sort(got)

		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("all media diff:\n%s", diff)
		}
	})

	t.Run("CheckMediaURL", func(t *testing.T) {
		var allMediaURL []*models.MediaURL
		if err := db.Find(&allMediaURL).Error; err != nil {
			t.Fatal("get all media url error:", err)
		}

		if got, want := len(allMediaURL), 18; got != want {
			t.Errorf("got = %d, want: %v", got, want)
		}

		want := slices.Clone(wantImages)
		wantThumbs := slices.Clone(wantImages)
		for _, thumb := range wantThumbs {
			want = append(want, "thumbnail_"+thumb)
		}
		slices.Sort(want)

		got := make([]string, len(allMediaURL))
		for i, media := range allMediaURL {
			got[i] = media.MediaName
		}
		slices.Sort(got)

		if diff := cmp.Diff(got, want, cmp.Comparer(equalNameWithoutSuffix)); diff != "" {
			t.Errorf("all media diff:\n%s", diff)
		}
	})

	t.Run("CheckFaceGroup", func(t *testing.T) {
		ctx, done := context.WithTimeout(t.Context(), time.Second*5)
		defer done()

		waitFor(ctx, t, time.Second/2, func() bool {
			var allFaceGroups []*models.FaceGroup
			if err := db.Find(&allFaceGroups).Error; err != nil {
				t.Fatal("get face groups error:", err)
				return false
			}

			return len(allFaceGroups) == len(wantFaceGroups)
		})
	})

	t.Run("CheckFaces", func(t *testing.T) {
		var allImageFaces []*models.ImageFace
		if err := db.Find(&allImageFaces).Error; err != nil {
			t.Fatal("get face images error:", err)
		}

		for _, face := range allImageFaces {
			if err := face.FillMedia(db); err != nil {
				t.Fatalf("fill media for face %v error: %v", face, err)
			}
		}

		got := groupMediaWithFaces(allImageFaces)

		if diff := cmp.Diff(got, wantFaceGroups); diff != "" {
			t.Errorf("all media diff:\n%s", diff)
		}
	})
}

func equalNameWithoutSuffix(a, b string) bool {
	extA := filepath.Ext(a)
	mainA := strings.TrimRight(a, extA)
	extB := filepath.Ext(b)
	mainB := strings.TrimRight(b, extB)

	// ext names are not same
	if extA != extB {
		return false
	}

	// a is not part of b and b is not part of a
	if strings.Index(mainA, mainB) < 0 && strings.Index(mainB, mainA) < 0 {
		return false
	}

	return true
}

func waitFor(ctx context.Context, t *testing.T, interval time.Duration, checkFn func() bool) {
	t.Helper()

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			t.Fatal("check timeout")
			return
		case <-ticker.C:
		}

		if checkFn() {
			return
		}
	}
}

func groupMediaWithFaces(medias []*models.ImageFace) [][]string {
	grouped := make(map[int][]string)

	for _, media := range medias {
		group := grouped[media.FaceGroupID]
		group = append(group, media.Media.Title)
		grouped[media.FaceGroupID] = group
	}

	ret := make([][]string, 0, len(grouped))
	for _, medias := range grouped {
		slices.Sort(medias)
		ret = append(ret, medias)
	}

	slices.SortFunc(ret, func(a, b []string) int {
		return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
	})

	return ret
}
