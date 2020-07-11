package models

import (
	"database/sql"
	"path"

	"github.com/viktorstrate/photoview/api/utils"
)

type MediaType string

const (
	MediaTypePhoto MediaType = "photo"
	MediaTypeVide  MediaType = "video"
)

type Media struct {
	MediaID  int
	Title    string
	Path     string
	PathHash string
	AlbumId  int
	ExifId   *int
	Favorite bool
	Type     MediaType
}

func (p *Media) ID() int {
	return p.MediaID
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
	UrlID       int
	MediaId     int
	MediaName   string
	Width       int
	Height      int
	Purpose     MediaPurpose
	ContentType string
}

func NewMediaFromRow(row *sql.Row) (*Media, error) {
	media := Media{}

	if err := row.Scan(&media.MediaID, &media.Title, &media.Path, &media.PathHash, &media.AlbumId, &media.ExifId, &media.Favorite, &media.Type); err != nil {
		return nil, err
	}

	return &media, nil
}

func NewMediaFromRows(rows *sql.Rows) ([]*Media, error) {
	medias := make([]*Media, 0)

	for rows.Next() {
		var media Media
		if err := rows.Scan(&media.MediaID, &media.Title, &media.Path, &media.PathHash, &media.AlbumId, &media.ExifId, &media.Favorite); err != nil {
			return nil, err
		}
		medias = append(medias, &media)
	}

	rows.Close()

	return medias, nil
}

func (p *MediaURL) URL() string {

	imageUrl := utils.ApiEndpointUrl()
	imageUrl.Path = path.Join(imageUrl.Path, "photo", p.MediaName)

	return imageUrl.String()
}

func NewMediaURLFromRow(row *sql.Row) (*MediaURL, error) {
	url := MediaURL{}

	if err := row.Scan(&url.UrlID, &url.MediaId, &url.MediaName, &url.Width, &url.Height, &url.Purpose, &url.ContentType); err != nil {
		return nil, err
	}

	return &url, nil
}

func NewMediaURLFromRows(rows *sql.Rows) ([]*MediaURL, error) {
	urls := make([]*MediaURL, 0)

	for rows.Next() {
		var url MediaURL
		if err := rows.Scan(&url.UrlID, &url.MediaId, &url.MediaName, &url.Width, &url.Height, &url.Purpose, &url.ContentType); err != nil {
			return nil, err
		}
		urls = append(urls, &url)
	}

	rows.Close()

	return urls, nil
}
