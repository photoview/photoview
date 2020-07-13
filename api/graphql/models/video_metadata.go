package models

import "database/sql"

type VideoMetadata struct {
	MetadataID   int
	Width        int
	Height       int
	Duration     float64
	Codec        *string
	Framerate    *float64
	Bitrate      *int
	ColorProfile *string
	Audio        *string
}

func (metadata *VideoMetadata) ID() int {
	return metadata.MetadataID
}

func (metadata *VideoMetadata) Media() *Media {
	panic("not implemented")
}

func NewVideoMetadataFromRow(row *sql.Row) (*VideoMetadata, error) {
	meta := VideoMetadata{}

	if err := row.Scan(&meta.MetadataID, &meta.Width, &meta.Height, &meta.Duration, &meta.Codec, &meta.Framerate, &meta.Bitrate, &meta.ColorProfile, &meta.Audio); err != nil {
		return nil, err
	}

	return &meta, nil
}
