package models

import (
	"time"
)

type ShareToken struct {
	Model
	Value    string `gorm:"not null"`
	OwnerID  int    `gorm:"not null"`
	Owner    User   `gorm:"constraint:OnDelete:CASCADE;"`
	Expire   *time.Time
	Password *string
	AlbumID  *int
	Album    *Album `gorm:"constraint:OnDelete:CASCADE;"`
	MediaID  *int
	Media    *Media `gorm:"constraint:OnDelete:CASCADE;"`
}

func (share *ShareToken) Token() string {
	return share.Value
}
