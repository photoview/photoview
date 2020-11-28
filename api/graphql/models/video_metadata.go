package models

type VideoMetadata struct {
	Model
	Width        int     `gorm:"not null"`
	Height       int     `gorm:"not null"`
	Duration     float64 `gorm:"not null"`
	Codec        *string
	Framerate    *float64
	Bitrate      *string
	ColorProfile *string
	Audio        *string
}

func (metadata *VideoMetadata) Media() *Media {
	panic("not implemented")
}
