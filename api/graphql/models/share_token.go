package models

import (
	"time"

	"gorm.io/gorm"
)

type ShareToken struct {
	gorm.Model
	Value    string
	OwnerID  uint
	Owner    User
	Expire   *time.Time
	Password *string
	AlbumID  *uint
	Album    Album
	MediaID  *uint
	Media    *Media
}

func (share *ShareToken) Token() string {
	return share.Value
}
