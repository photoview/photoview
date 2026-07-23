package queue

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_type"
)

// urls is a small helper to build an existingURLs map with placeholder
// (zero-value) *models.MediaURL entries - decide() only ever checks presence
// (== nil) of a purpose, never any field on the URL itself.
func urls(purposes ...models.MediaPurpose) map[models.MediaPurpose]*models.MediaURL {
	m := make(map[models.MediaPurpose]*models.MediaURL, len(purposes))
	for _, p := range purposes {
		m[p] = &models.MediaURL{}
	}
	return m
}

func TestDecide(t *testing.T) {
	hasBlurhash := "blurhash"

	tests := []struct {
		name string
		info gatheredInfo
		want workPlan
	}{
		{
			name: "new web photo, nothing exists yet",
			info: gatheredInfo{
				mediaType:    media_type.TypeJPEG,
				isNewMedia:   true,
				media:        &models.Media{},
				existingURLs: urls(),
			},
			want: workPlan{
				isNewMedia:    true,
				needExif:      true,
				needOriginal:  true,
				needThumbnail: true,
				mediaChanged:  true,
				needFaces:     true,
				needBlurhash:  true,
			},
		},
		{
			name: "new raw photo, nothing exists yet, no sidecar",
			info: gatheredInfo{
				mediaType:    media_type.TypeImage, // non-web-compatible stand-in
				isNewMedia:   true,
				media:        &models.Media{},
				existingURLs: urls(),
			},
			want: workPlan{
				isNewMedia:    true,
				needExif:      true,
				needOriginal:  true,
				needHighres:   true,
				needThumbnail: true,
				mediaChanged:  true,
				needFaces:     true,
				needBlurhash:  true,
			},
		},
		{
			name: "new raw photo, sidecar file present",
			info: gatheredInfo{
				mediaType:    media_type.TypeImage,
				isNewMedia:   true,
				media:        &models.Media{}, // SideCarHash nil -> counts as changed
				existingURLs: urls(),
				sidecarPath:  new("/album/photo.tiff.xmp"),
				sidecarHash:  new("hash1"),
			},
			want: workPlan{
				isNewMedia:     true,
				needExif:       true,
				needOriginal:   true,
				needHighres:    true,
				needThumbnail:  true,
				sidecarChanged: true,
				mediaChanged:   true,
				needFaces:      true,
				needBlurhash:   true,
			},
		},
		{
			name: "existing web photo, nothing missing, has blurhash -> nothing to do",
			info: gatheredInfo{
				mediaType:    media_type.TypeJPEG,
				isNewMedia:   false,
				media:        &models.Media{Blurhash: &hasBlurhash},
				existingURLs: urls(models.MediaOriginal, models.PhotoThumbnail),
			},
			want: workPlan{},
		},
		{
			name: "existing web photo, nothing missing, no blurhash yet",
			info: gatheredInfo{
				mediaType:    media_type.TypeJPEG,
				isNewMedia:   false,
				media:        &models.Media{},
				existingURLs: urls(models.MediaOriginal, models.PhotoThumbnail),
			},
			want: workPlan{
				needBlurhash: true,
			},
		},
		{
			name: "existing web photo, thumbnail cache file missing",
			info: gatheredInfo{
				mediaType:        media_type.TypeJPEG,
				isNewMedia:       false,
				media:            &models.Media{Blurhash: &hasBlurhash},
				existingURLs:     urls(models.MediaOriginal, models.PhotoThumbnail),
				thumbnailMissing: true,
			},
			want: workPlan{
				needThumbnail: true,
				mediaChanged:  true,
				needFaces:     true,
				needBlurhash:  true,
			},
		},
		{
			name: "existing raw photo, sidecar hash changed forces regen even though nothing else missing",
			info: gatheredInfo{
				mediaType:    media_type.TypeImage,
				isNewMedia:   false,
				media:        &models.Media{SideCarHash: new("old"), Blurhash: &hasBlurhash},
				existingURLs: urls(models.MediaOriginal, models.PhotoHighRes, models.PhotoThumbnail),
				sidecarPath:  new("/album/photo.tiff.xmp"),
				sidecarHash:  new("new"),
			},
			want: workPlan{
				needHighres:    true,
				needThumbnail:  true,
				sidecarChanged: true,
				mediaChanged:   true,
				needFaces:      true,
				needBlurhash:   true,
			},
		},
		{
			name: "existing raw photo, sidecar deleted since last scan",
			info: gatheredInfo{
				mediaType:    media_type.TypeImage,
				isNewMedia:   false,
				media:        &models.Media{SideCarPath: new("/album/photo.tiff.xmp"), SideCarHash: new("old"), Blurhash: &hasBlurhash},
				existingURLs: urls(models.MediaOriginal, models.PhotoHighRes, models.PhotoThumbnail),
				sidecarPath:  nil,
			},
			want: workPlan{
				needHighres:    true,
				needThumbnail:  true,
				sidecarChanged: true,
				mediaChanged:   true,
				needFaces:      true,
				needBlurhash:   true,
			},
		},
		{
			name: "existing raw photo, sidecar unchanged, nothing missing -> nothing to do",
			info: gatheredInfo{
				mediaType:    media_type.TypeImage,
				isNewMedia:   false,
				media:        &models.Media{SideCarHash: new("same"), Blurhash: &hasBlurhash},
				existingURLs: urls(models.MediaOriginal, models.PhotoHighRes, models.PhotoThumbnail),
				sidecarPath:  new("/album/photo.tiff.xmp"),
				sidecarHash:  new("same"),
			},
			want: workPlan{},
		},
		{
			name: "new web video, nothing exists yet",
			info: gatheredInfo{
				mediaType:    media_type.TypeMP4,
				isVideo:      true,
				isNewMedia:   true,
				media:        &models.Media{},
				existingURLs: urls(),
			},
			want: workPlan{
				isNewMedia:     true,
				needExif:       true,
				needVideoMeta:  true,
				needOriginal:   true,
				needVideoThumb: true,
				mediaChanged:   true,
				needBlurhash:   true,
			},
		},
		{
			name: "new non-web video, nothing exists yet - no original URL for non-web video",
			info: gatheredInfo{
				mediaType:    media_type.TypeVideo, // non-web-compatible stand-in
				isVideo:      true,
				isNewMedia:   true,
				media:        &models.Media{},
				existingURLs: urls(),
			},
			want: workPlan{
				isNewMedia:     true,
				needExif:       true,
				needVideoMeta:  true,
				needWebVideo:   true,
				needVideoThumb: true,
				mediaChanged:   true,
				needBlurhash:   true,
			},
		},
		{
			name: "existing web video, nothing missing, has blurhash -> nothing to do",
			info: gatheredInfo{
				mediaType:    media_type.TypeMP4,
				isVideo:      true,
				isNewMedia:   false,
				media:        &models.Media{Blurhash: &hasBlurhash},
				existingURLs: urls(models.MediaOriginal, models.VideoThumbnail),
			},
			want: workPlan{},
		},
		{
			name: "existing non-web video, web_video cache file missing despite MediaURL row existing",
			info: gatheredInfo{
				mediaType:       media_type.TypeVideo,
				isVideo:         true,
				isNewMedia:      false,
				media:           &models.Media{Blurhash: &hasBlurhash},
				existingURLs:    urls(models.VideoWeb, models.VideoThumbnail),
				webVideoMissing: true,
			},
			want: workPlan{
				needWebVideo: true,
				mediaChanged: true,
				needBlurhash: true,
			},
		},
		{
			name: "existing video, video thumbnail cache file missing",
			info: gatheredInfo{
				mediaType:         media_type.TypeMP4,
				isVideo:           true,
				isNewMedia:        false,
				media:             &models.Media{Blurhash: &hasBlurhash},
				existingURLs:      urls(models.MediaOriginal, models.VideoThumbnail),
				videoThumbMissing: true,
			},
			want: workPlan{
				needVideoThumb: true,
				mediaChanged:   true,
				needBlurhash:   true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := decide(tc.info)
			if got != tc.want {
				t.Errorf("decide() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestWorkPlanNothingToDo(t *testing.T) {
	tests := []struct {
		name string
		plan workPlan
		want bool
	}{
		{name: "zero value", plan: workPlan{}, want: true},
		{name: "isNewMedia alone does not count as work", plan: workPlan{isNewMedia: true}, want: true},
		{name: "needExif", plan: workPlan{needExif: true}, want: false},
		{name: "needVideoMeta", plan: workPlan{needVideoMeta: true}, want: false},
		{name: "needOriginal", plan: workPlan{needOriginal: true}, want: false},
		{name: "mediaChanged", plan: workPlan{mediaChanged: true}, want: false},
		{name: "needFaces", plan: workPlan{needFaces: true}, want: false},
		{name: "needBlurhash", plan: workPlan{needBlurhash: true}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.plan.nothingToDo(); got != tc.want {
				t.Errorf("nothingToDo() = %v, want %v", got, tc.want)
			}
		})
	}
}
