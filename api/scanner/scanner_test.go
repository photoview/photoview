package scanner_test

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"github.com/photoview/photoview/api/test_utils"
	scanner_utils "github.com/photoview/photoview/api/test_utils/scanner"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
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
		Path:  "./test_media",
	}

	wantWebPhotos := []string{
		"bmp.bmp",
		"gif.gif",
		"jpeg.jpg",
		"png.png",
		"webp.webp",

		"jpg_with_file.jpg",
		"recoverable_bad_rst_marker.jpg",
		"standalone_jpg.jpg",

		"boy1.jpg",
		"boy2.jpg",
		"buttercup_close_summer_yellow.jpg",
		"girl_black_hair2.jpg",
		"girl_blond1.jpg",
		"girl_blond2.jpg",
		"girl_blond3.jpg",
		"lilac_lilac_bush_lilac.jpg",
		"mount_merapi_volcano_indonesia.jpg",

		"left_arrow_normal_web.jpg",
		"up_arrow_90cw_web.jpg",
	}
	wantNonWebPhotos := []string{
		"heif.heif",
		"jpegxl.jxl",
		"jpg2000.jp2",
		"raw_with_file.tiff",
		"raw_with_jpg.tiff",
		"standalone_raw.tiff",
		"tiff.tiff",
		"raw_Canon.CR3",

		"left_arrow_normal_nonweb.tiff",
		"up_arrow_90cw_nonweb.tiff",
	}
	wantWebVideos := []string{
		"mp4.mp4",
		"mpeg.mpg",
		"ogg.ogg",
		"webm.webm",
	}
	wantNonWebVideos := []string{
		"avi.avi",
		"mkv.mkv",
		"quicktime.mov",
		"wmv.wmv",
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

		want := []string{}
		want = append(want, wantWebPhotos...)
		want = append(want, wantNonWebPhotos...)
		want = append(want, wantWebVideos...)
		want = append(want, wantNonWebVideos...)
		slices.Sort(want)

		got := make([]string, len(allMedia))
		for i, media := range allMedia {
			got[i] = media.Title
			if media.Blurhash == nil {
				t.Errorf("media %q(%s) doesn't have Blurhash, while it should have", media.Title, media.Type)
			}
		}
		slices.Sort(got)

		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("all media diff (-got, +want):\n%s", diff)
		}
	})

	t.Run("CheckMediaURL", func(t *testing.T) {
		var allMediaURL []*models.MediaURL
		if err := db.Find(&allMediaURL).Error; err != nil {
			t.Fatal("get all media url error:", err)
		}

		var want []string
		want = append(want, wantWebPhotos...)
		for _, name := range wantWebPhotos {
			want = append(want, "thumbnail_"+strings.ReplaceAll(name, ".", "_")+".jpg")
		}

		want = append(want, wantNonWebPhotos...)
		for _, name := range wantNonWebPhotos {
			want = append(want, "thumbnail_"+strings.ReplaceAll(name, ".", "_")+".jpg")
			want = append(want, "highres_"+strings.ReplaceAll(name, ".", "_")+".jpg")
		}

		want = append(want, wantWebVideos...)
		for _, name := range wantWebVideos {
			want = append(want, "video_thumb_"+strings.ReplaceAll(name, ".", "_")+".jpg")
		}

		for _, name := range wantNonWebVideos {
			want = append(want, "video_thumb_"+strings.ReplaceAll(name, ".", "_")+".jpg")
			want = append(want, "web_video_"+strings.ReplaceAll(name, ".", "_")+".mp4")
		}

		slices.Sort(want)

		if got, want := len(allMediaURL), len(want); got != want {
			t.Errorf("got = %d, want: %v", got, want)
		}

		got := make([]string, len(allMediaURL))
		for i, media := range allMediaURL {
			got[i] = media.MediaName
		}
		slices.Sort(got)

		if diff := cmp.Diff(got, want, cmp.Comparer(equalNameWithoutSuffix)); diff != "" {
			t.Errorf("all media url diff (-got, +want):\n%s", diff)
		}
	})

	t.Run("CheckFaceGroup", func(t *testing.T) {
		var allFaceGroups []*models.FaceGroup
		if err := db.Find(&allFaceGroups).Error; err != nil {
			t.Fatal("get face groups error:", err)
		}

		if got, want := len(allFaceGroups), len(wantFaceGroups); got != want {
			t.Errorf("len(allFaceGroups) = %d, want: %d", got, want)
		}
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
			t.Errorf("all media diff (-got, +want):\n%s", diff)
		}
	})

	t.Run("CheckPhotosOrientation", func(t *testing.T) {
		photoFiles := []string{
			"left_arrow_normal_web.jpg",
			"up_arrow_90cw_web.jpg",
			"left_arrow_normal_nonweb.tiff",
			"up_arrow_90cw_nonweb.tiff",
		}
		for _, filename := range photoFiles {
			var media models.Media
			if err := db.Preload("MediaURL").Where("title = ?", filename).Find(&media).Error; err != nil {
				t.Fatalf("can't find media with name %q: %v", filename, err)
			}

			thumbnail, err := media.GetThumbnail()
			if err != nil {
				t.Fatalf("can't get thumbnail of media %q: %v", filename, err)
			}

			switch {
			case strings.HasPrefix(filename, "up"):
				if thumbnail.Width >= thumbnail.Height {
					t.Errorf("media %q dimension: %dx%d, which should be a vertial photo", filename, thumbnail.Width, thumbnail.Height)
				}
			case strings.HasPrefix(filename, "left"):
				if thumbnail.Width <= thumbnail.Height {
					t.Errorf("media %q dimension: %dx%d, which should be a horizontal photo", filename, thumbnail.Width, thumbnail.Height)
				}
			}
		}
	})
}

func equalNameWithoutSuffix(a, b string) bool {
	extA := filepath.Ext(a)
	mainA := strings.TrimSuffix(a, extA)
	extB := filepath.Ext(b)
	mainB := strings.TrimSuffix(b, extB)

	// ext names are not same
	if extA != extB {
		return false
	}

	// a is not prefix of b and b is not prefix of a
	if strings.HasPrefix(mainA, mainB) && strings.HasPrefix(mainB, mainA) {
		return false
	}

	return true
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

func copyFilelistWithJpgExt(list []string) []string {
	ret := make([]string, 0, len(list))
	for _, f := range list {
		ext := filepath.Ext(f)
		main := strings.TrimSuffix(f, ext)

		ret = append(ret, main+".jpg")
	}

	return ret
}
