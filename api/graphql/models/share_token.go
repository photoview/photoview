package models

import (
	"time"
)

type ShareToken struct {
	Model
	Value    string
	OwnerID  int
	Owner    User
	Expire   *time.Time
	Password *string
	AlbumID  *int
	Album    Album
	MediaID  *int
	Media    *Media
}

func (share *ShareToken) Token() string {
	return share.Value
}
