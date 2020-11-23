package models

import (
	"gorm.io/gorm"
)

type Album struct {
	gorm.Model
	Title       string
	ParentAlbum *int
	OwnerID     int
	Owner       User
	Path        string
	PathHash    string
}

func (a *Album) FilePath() string {
	return a.Path
}
