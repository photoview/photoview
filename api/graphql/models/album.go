package models

import (
	"crypto/md5"
	"encoding/hex"

	"gorm.io/gorm"
)

type Album struct {
	Model
	Title         string `gorm:"not null"`
	ParentAlbumID *int   `gorm:"index"`
	ParentAlbum   *Album `gorm:"constraint:OnDelete:SET NULL;"`
	// OwnerID       int `gorm:"not null"`
	// Owner         User
	Owners   []User `gorm:"many2many:user_albums"`
	Path     string `gorm:"not null"`
	PathHash string `gorm:"unique"`
	IgnoreFiles string
}

func (a *Album) FilePath() string {
	return a.Path
}

func (a *Album) BeforeSave(tx *gorm.DB) (err error) {
	hash := md5.Sum([]byte(a.Path))
	a.PathHash = hex.EncodeToString(hash[:])
	return nil
}

func (a *Album) GetChildren(db *gorm.DB) (children []*Album, err error) {
	err = db.Raw(`
	WITH recursive sub_albums AS (
		SELECT * FROM albums AS root WHERE id = ?
		UNION ALL
		SELECT child.* FROM albums AS child JOIN sub_albums ON child.parent_album_id = sub_albums.id
	)

	SELECT * FROM sub_albums
	`, a.ID).Find(&children).Error

	return children, err
}
