package models

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID int `gorm:"primarykey"`
	ModelTimestamps
}

type ModelTimestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
