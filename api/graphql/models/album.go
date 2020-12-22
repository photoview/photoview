package models

import (
	"crypto/md5"
	"encoding/hex"

	"gorm.io/gorm"
)

type Album struct {
	Model
	Title         string `gorm:"not null"`
	ParentAlbumID *int
	ParentAlbum   *Album
	// OwnerID       int `gorm:"not null"`
	// Owner         User
	Owners   []User `gorm:"many2many:user_albums"`
	Path     string `gorm:"not null"`
	PathHash string `gorm:"unique"`
}

func (a *Album) FilePath() string {
	return a.Path
}

func (a *Album) BeforeSave(tx *gorm.DB) (err error) {
	hash := md5.Sum([]byte(a.Path))
	a.PathHash = hex.EncodeToString(hash[:])
	return nil
}
