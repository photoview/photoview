package models

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Media struct {
	Model
	Title           string         `gorm:"not null"`
	Path            string         `gorm:"not null"`
	PathHash        string         `gorm:"not null;unique"`
	AlbumID         int            `gorm:"not null;index"`
	Album           Album          `gorm:"constraint:OnDelete:CASCADE;"`
	ExifID          *int           `gorm:"index"`
	Exif            *MediaEXIF     `gorm:"constraint:OnDelete:CASCADE;"`
	MediaURL        []MediaURL     `gorm:"constraint:OnDelete:CASCADE;"`
	DateShot        time.Time      `gorm:"not null"`
	Type            MediaType      `gorm:"not null;index"`
	VideoMetadataID *int           `gorm:"index"`
	VideoMetadata   *VideoMetadata `gorm:"constraint:OnDelete:CASCADE;"`
	SideCarPath     *string
	SideCarHash     *string      `gorm:"unique"`
	Faces           []*ImageFace `gorm:"constraint:OnDelete:CASCADE;"`

	// Only used internally
	CounterpartPath *string `gorm:"-"`
}

func (Media) TableName() string {
	return "media"
}

func (m *Media) BeforeSave(tx *gorm.DB) error {
	// Update path hash
	m.PathHash = MD5Hash(m.Path)

	return nil
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
	MediaID     int          `gorm:"not null;index"`
	Media       *Media       `gorm:"constraint:OnDelete:CASCADE;"`
	MediaName   string       `gorm:"not null"`
	Width       int          `gorm:"not null"`
	Height      int          `gorm:"not null"`
	Purpose     MediaPurpose `gorm:"not null;index"`
	ContentType string       `gorm:"not null"`
	FileSize    int64        `gorm:"not null"`
}

func (p *MediaURL) URL() string {

	imageURL := utils.ApiEndpointUrl()
	if p.Purpose != VideoWeb {
		imageURL.Path = path.Join(imageURL.Path, "photo", p.MediaName)
	} else {
		imageURL.Path = path.Join(imageURL.Path, "video", p.MediaName)
	}

	return imageURL.String()
}

func (p *MediaURL) CachedPath() (string, error) {
	var cachedPath string

	if p.Media == nil {
		return "", errors.New("mediaURL.Media is nil")
	}

	if p.Purpose == PhotoThumbnail || p.Purpose == PhotoHighRes || p.Purpose == VideoThumbnail {
		cachedPath = path.Join(utils.MediaCachePath(), strconv.Itoa(int(p.Media.AlbumID)), strconv.Itoa(int(p.MediaID)), p.MediaName)
	} else if p.Purpose == MediaOriginal {
		cachedPath = p.Media.Path
	} else {
		return "", errors.New(fmt.Sprintf("cannot determine cache path for purpose (%s)", p.Purpose))
	}

	return cachedPath, nil
}

func SanitizeMediaName(mediaName string) string {
	result := mediaName
	result = strings.ReplaceAll(result, "/", "")
	result = strings.ReplaceAll(result, "\\", "")
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, ".", "_")
	return result
}
