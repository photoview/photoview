package models

import "time"

type ShareToken struct {
	TokenID  int
	Value    string
	OwnerID  int
	Expire   *time.Time
	Password *string
	AlbumID  *int
	PhotoID  *int
}

func (share *ShareToken) Token() string {
	return share.Value
}

func (share *ShareToken) ID() int {
	return share.TokenID
}
