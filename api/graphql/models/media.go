package models

import (
	"path"
	"strings"
	"time"

	"github.com/viktorstrate/photoview/api/utils"
	"gorm.io/gorm"
)

type Media struct {
	gorm.Model
	Title           string
	Path            string
	PathHash        string
	AlbumId         uint
	Album           Album
	ExifId          *uint
	Exif            MediaEXIF
	MediaURL        []MediaURL
	DateShot        time.Time
	DateImported    time.Time
	Favorite        bool
	Type            MediaType
	VideoMetadataId *int
	SideCarPath     *string
	SideCarHash     *string
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
	gorm.Model
	MediaID     uint
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
