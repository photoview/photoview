package models

import (
	"crypto/md5"
	"encoding/hex"

	"gorm.io/gorm"
)

type Album struct {
	Model
	Title         string
	ParentAlbumID *int
	ParentAlbum   *Album
	OwnerID       int
	Owner         User
	Path          string
	PathHash      string `gorm:"unique"`
}

func (a *Album) FilePath() string {
	return a.Path
}

func (a *Album) BeforeSave(tx *gorm.DB) (err error) {
	hash := md5.Sum([]byte(a.Path))
	a.PathHash = hex.EncodeToString(hash[:])
	return nil
}
