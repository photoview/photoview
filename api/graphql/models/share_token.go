package models

import (
	"time"
)

type ShareToken struct {
	Model
	Value    string     `gorm:"not null"`
	OwnerID  int        `gorm:"not null;index"`
	Owner    User       `gorm:"constraint:OnDelete:CASCADE;"`
	Expire   *time.Time `gorm:"index"`
	Password *string
	AlbumID  *int   `gorm:"index"`
	Album    *Album `gorm:"constraint:OnDelete:CASCADE;"`
	MediaID  *int   `gorm:"index"`
	Media    *Media `gorm:"constraint:OnDelete:CASCADE;"`
}

func (share *ShareToken) Token() string {
	return share.Value
}
