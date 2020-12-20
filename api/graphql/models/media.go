package models

import (
	"crypto/md5"
	"encoding/hex"
	"path"
	"strings"
	"time"

	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

type Media struct {
	Model
	Title        string `gorm:"not null"`
	Path         string `gorm:"not null"`
	PathHash     string `gorm:"not null"`
	AlbumID      int    `gorm:"not null"`
	Album        Album
	ExifID       *int
	Exif         *MediaEXIF
	MediaURL     []MediaURL
	DateShot     time.Time `gorm:"not null"`
	DateImported time.Time `gorm:"not null"`
	// Favorite        bool      `gorm:"not null, default:false"`
	Type            MediaType `gorm:"not null"`
	VideoMetadataID *int
	VideoMetadata   *VideoMetadata
	SideCarPath     *string
	SideCarHash     *string

	// Only used internally
	CounterpartPath *string `gorm:-`
}

func (Media) TableName() string {
	return "media"
}

func (m *Media) BeforeSave(tx *gorm.DB) error {
	// Update hashes
	hash := md5.Sum([]byte(m.Path))
	m.PathHash = hex.EncodeToString(hash[:])

	if m.SideCarPath != nil {
		hash = md5.Sum([]byte(*m.SideCarPath))
		encodedHash := hex.EncodeToString(hash[:])
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
	MediaID     int `gorm:"not null"`
	Media       Media
	MediaName   string       `gorm:"not null"`
	Width       int          `gorm:"not null"`
	Height      int          `gorm:"not null"`
	Purpose     MediaPurpose `gorm:"not null"`
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
