package models

type VideoMetadata struct {
	Model
	Width        int
	Height       int
	Duration     float64
	Codec        *string
	Framerate    *float64
	Bitrate      *string
	ColorProfile *string
	Audio        *string
}

func (metadata *VideoMetadata) Media() *Media {
	panic("not implemented")
}
