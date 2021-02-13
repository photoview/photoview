package models

import (
	"path"
	"strings"
	"time"

	"github.com/photoview/photoview/api/utils"
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
	SideCarHash     *string `gorm:"unique"`

	// Only used internally
	CounterpartPath *string `gorm:"-"`
}

func (Media) TableName() string {
	return "media"
}

func (m *Media) BeforeSave(tx *gorm.DB) error {
	// Update hashes
	m.PathHash = MD5Hash(m.Path)

	if m.SideCarPath != nil {
		encodedHash := MD5Hash(*m.SideCarPath)
		m.SideCarHash = &encodedHash
	}

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
	Media       Media        `gorm:"constraint:OnDelete:CASCADE;"`
	MediaName   string       `gorm:"not null"`
	Width       int          `gorm:"not null"`
	Height      int          `gorm:"not null"`
	Purpose     MediaPurpose `gorm:"not null;index"`
	ContentType string       `gorm:"not null"`
	FileSize    int64        `gorm:"not null"`
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
