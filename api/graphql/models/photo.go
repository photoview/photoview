package models

import (
	"database/sql"
	"path"

	"github.com/viktorstrate/photoview/api/utils"
)

type Photo struct {
	PhotoID  int
	Title    string
	Path     string
	PathHash string
	AlbumId  int
	ExifId   *int
	Favorite bool
}

func (p *Photo) ID() int {
	return p.PhotoID
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

func NewPhotoFromRow(row *sql.Row) (*Photo, error) {
	photo := Photo{}

	if err := row.Scan(&photo.PhotoID, &photo.Title, &photo.Path, &photo.PathHash, &photo.AlbumId, &photo.ExifId, &photo.Favorite); err != nil {
		return nil, err
	}

	return &photo, nil
}

func NewPhotosFromRows(rows *sql.Rows) ([]*Photo, error) {
	photos := make([]*Photo, 0)

	for rows.Next() {
		var photo Photo
		if err := rows.Scan(&photo.PhotoID, &photo.Title, &photo.Path, &photo.PathHash, &photo.AlbumId, &photo.ExifId, &photo.Favorite); err != nil {
			return nil, err
		}
		photos = append(photos, &photo)
	}

	rows.Close()

	return photos, nil
}

func (p *PhotoURL) URL() string {

	imageUrl := utils.ApiEndpointUrl()
	imageUrl.Path = path.Join(imageUrl.Path, "photo", p.PhotoName)

	return imageUrl.String()
}

func NewPhotoURLFromRow(row *sql.Row) (*PhotoURL, error) {
	url := PhotoURL{}

	if err := row.Scan(&url.UrlID, &url.PhotoId, &url.PhotoName, &url.Width, &url.Height, &url.Purpose, &url.ContentType); err != nil {
		return nil, err
	}

	return &url, nil
}

func NewPhotoURLFromRows(rows *sql.Rows) ([]*PhotoURL, error) {
	urls := make([]*PhotoURL, 0)

	for rows.Next() {
		var url PhotoURL
		if err := rows.Scan(&url.UrlID, &url.PhotoId, &url.PhotoName, &url.Width, &url.Height, &url.Purpose, &url.ContentType); err != nil {
			return nil, err
		}
		urls = append(urls, &url)
	}

	rows.Close()

	return urls, nil
}
