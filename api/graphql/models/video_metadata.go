package models

import (
	"gorm.io/gorm"
)

type VideoMetadata struct {
	gorm.Model
	Width        int
	Height       int
	Duration     float64
	Codec        *string
	Framerate    *float64
	Bitrate      *string
	ColorProfile *string
	Audio        *string
}
