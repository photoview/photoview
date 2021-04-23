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
