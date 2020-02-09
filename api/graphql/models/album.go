package models

import (
	"database/sql"
	"strconv"
)

type Album struct {
	AlbumID     int
	Title       string
	ParentAlbum *int
	OwnerID     int
	Path        string
}

func (a *Album) ID() string {
	return strconv.Itoa(a.AlbumID)
}

func NewAlbumFromRow(row *sql.Row) (*Album, error) {
	album := Album{}

	if err := row.Scan(&album.AlbumID, &album.Title, &album.ParentAlbum, &album.OwnerID, &album.Path); err != nil {
		return nil, err
	}

	return &album, nil
}

func NewAlbumsFromRows(rows *sql.Rows) ([]*Album, error) {
	albums := make([]*Album, 0)

	for rows.Next() {
		var album Album
		if err := rows.Scan(&album.AlbumID, &album.Title, &album.ParentAlbum, &album.OwnerID, &album.Path); err != nil {
			return nil, err
		}
		albums = append(albums, &album)
	}

	return albums, nil
}
