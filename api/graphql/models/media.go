package models

import (
	"path"
	"strings"
	"time"

	"github.com/viktorstrate/photoview/api/utils"
)

type Media struct {
	Model
	Title           string
	Path            string
	PathHash        string
	AlbumID         int
	Album           Album
	ExifID          *int
	Exif            *MediaEXIF
	MediaURL        []MediaURL
	DateShot        time.Time
	DateImported    time.Time
	Favorite        bool
	Type            MediaType
	VideoMetadataID *int
	VideoMetadata   *VideoMetadata
	SideCarPath     *string
	SideCarHash     *string
}

func (Media) TableName() string {
	return "media"
}

type MediaPurpose string

const (
	PhotoThumbnail MediaPurpose = "thumbnail"
	PhotoHighRes   MediaPurpose = "high-res"
	MediaOriginal  MediaPurpose = "original"
	VideoWeb       MediaPurpose = "video-web"
	VideoThumbnail MediaPurpose = "video-thumbnail"
)

type MediaURL struct {
	Model
	MediaID     int
	Media       Media
	MediaName   string
	Width       int
	Height      int
	Purpose     MediaPurpose
	ContentType string
	FileSize    int64
}

func (p *MediaURL) URL() string {

	imageUrl := utils.ApiEndpointUrl()
	if p.Purpose != VideoWeb {
		imageUrl.Path = path.Join(imageUrl.Path, "photo", p.MediaName)
	} else {
		imageUrl.Path = path.Join(imageUrl.Path, "video", p.MediaName)
	}

	return imageUrl.String()
}

func SanitizeMediaName(mediaName string) string {
	result := mediaName
	result = strings.ReplaceAll(result, "/", "")
	result = strings.ReplaceAll(result, "\\", "")
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, ".", "_")
	return result
}
