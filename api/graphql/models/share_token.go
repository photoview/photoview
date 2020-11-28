package models

import (
	"time"
)

type ShareToken struct {
	Model
	Value    string `gorm:"not null"`
	OwnerID  int    `gorm:"not null"`
	Owner    User
	Expire   *time.Time
	Password *string
	AlbumID  *int
	Album    *Album
	MediaID  *int
	Media    *Media
}

func (share *ShareToken) Token() string {
	return share.Value
}
