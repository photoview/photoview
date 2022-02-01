package models_test

import (
	"fmt"
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeMediaName(t *testing.T) {
	tests := [][2]string{
		{"filename.png", "filename_png"},
		{"../..\\escape", "____escape"},
		{"..", "__"},
		{"..\\/", "__"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("sanitize: %s", test[0]), func(t *testing.T) {
			assert.Equal(t, test[1], models.SanitizeMediaName(test[0]))
		})
	}
}

func TestMediaURLCachePath(t *testing.T) {
	mediaUrl := models.MediaURL{}
	mediaUrl.Media = nil

	_, err := mediaUrl.CachedPath()
	assert.EqualError(t, err, "mediaURL.Media is nil")

	mediaUrl = models.MediaURL{
		Purpose: models.PhotoThumbnail,
		MediaID: 1,
		Media: &models.Media{
			Model: models.Model{
				ID: 1,
			},
			Title:   "media.jpg",
			AlbumID: 2,
		},
		MediaName: "media_thumb.jpg",
	}

	path, err := mediaUrl.CachedPath()

	assert.NoError(t, err)
	assert.Equal(t, "media_cache/2/1/media_thumb.jpg", path)

}

func TestMediaURLGetURL(t *testing.T) {
	photo := models.MediaURL{
		MediaName:   "photo.jpg",
		ContentType: "image/jpeg",
		Purpose:     models.PhotoHighRes,
	}

	assert.Equal(t, "photo/photo.jpg", photo.URL())

	video := models.MediaURL{
		MediaName:   "video.mp4",
		ContentType: "video/mp4",
		Purpose:     models.VideoWeb,
	}

	assert.Equal(t, "video/video.mp4", video.URL())
}

func TestMediaGetThumbnail(t *testing.T) {
	photo := models.Media{
		Title: "test.png",
		Path:  "path/test.png",
		Type:  models.MediaTypePhoto,
		MediaURL: []models.MediaURL{
			{
				MediaName:   "photo.jpg",
				ContentType: "image/jpeg",
				Purpose:     models.PhotoHighRes,
			},
			{
				MediaName:   "thumbnail.jpg",
				ContentType: "image/jpeg",
				Purpose:     models.PhotoThumbnail,
			},
			{
				MediaName:   "photo.png",
				ContentType: "image/png",
				Purpose:     models.MediaOriginal,
			},
		},
	}

	thumb, err := photo.GetThumbnail()
	assert.NoError(t, err)
	assert.Equal(t, thumb.MediaName, "thumbnail.jpg")
	assert.NotNil(t, thumb.Media)

	video := models.Media{
		Title: "video-test.mp4",
		Path:  "path/test.mp4",
		Type:  models.MediaTypePhoto,
		MediaURL: []models.MediaURL{
			{
				MediaName:   "video.mp4",
				ContentType: "video/mp4",
				Purpose:     models.VideoWeb,
			},
			{
				MediaName:   "video-thumbnail.jpg",
				ContentType: "image/jpg",
				Purpose:     models.VideoThumbnail,
			},
		},
	}

	thumb, err = video.GetThumbnail()
	assert.NoError(t, err)
	assert.Equal(t, thumb.MediaName, "video-thumbnail.jpg")
	assert.NotNil(t, thumb.Media)
}
