package models

import (
	"database/sql"
	"fmt"
	"strconv"
)

type Photo struct {
	PhotoID int
	Title   string
	Path    string
	AlbumId int
	ExifId  *int
}

type PhotoPurpose string

const (
	PhotoThumbnail PhotoPurpose = "thumbnail"
	PhotoHighRes   PhotoPurpose = "high-res"
	PhotoOriginal  PhotoPurpose = "original"
)

type PhotoURL struct {
	UrlID       int
	PhotoId     int
	PhotoName   string
	Width       int
	Height      int
	Purpose     PhotoPurpose
	ContentType string
}

func (p *Photo) ID() string {
	return strconv.Itoa(p.PhotoID)
}

func NewPhotoFromRow(row *sql.Row) (*Photo, error) {
	photo := Photo{}

	if err := row.Scan(&photo.PhotoID, &photo.Title, &photo.Path, &photo.AlbumId, &photo.ExifId); err != nil {
		return nil, err
	}

	return &photo, nil
}

func NewPhotosFromRows(rows *sql.Rows) ([]*Photo, error) {
	photos := make([]*Photo, 0)

	for rows.Next() {
		var photo Photo
		if err := rows.Scan(&photo.PhotoID, &photo.Title, &photo.Path, &photo.AlbumId, &photo.ExifId); err != nil {
			return nil, err
		}
		photos = append(photos, &photo)
	}

	return photos, nil
}

func (p *PhotoURL) URL() string {
	return fmt.Sprintf("/photo/%s", p.PhotoName)
}

func NewPhotoURLFromRow(row *sql.Row) (*PhotoURL, error) {
	url := PhotoURL{}

	if err := row.Scan(&url.UrlID, &url.PhotoId, &url.PhotoName, &url.Width, &url.Height, &url.Purpose, &url.ContentType); err != nil {
		return nil, err
	}

	return &url, nil
}
