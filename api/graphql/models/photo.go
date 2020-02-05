package models

import (
	"database/sql"
	"strconv"
)

type Photo struct {
	PhotoID      int
	Title        string
	Path         string
	OriginalUrl  int
	ThumbnailUrl int
	AlbumId      int
	ExifId       *int
}

type PhotoURL struct {
	UrlId  int
	Token  string
	Width  int
	Height int
}

func (p *Photo) ID() string {
	return strconv.Itoa(p.PhotoID)
}

func NewPhotoFromRow(row *sql.Row) (*Photo, error) {
	photo := Photo{}

	if err := row.Scan(&photo.PhotoID, &photo.Title, &photo.Path, &photo.OriginalUrl, &photo.ThumbnailUrl, &photo.AlbumId, &photo.ExifId); err != nil {
		return nil, err
	}

	return &photo, nil
}

func (p *PhotoURL) URL() string {
	return "URL:" + p.Token
}
