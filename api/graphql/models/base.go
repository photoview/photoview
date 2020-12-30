package models

import (
	"time"
)

type Model struct {
	ID int `gorm:"primarykey"`
	ModelTimestamps
}

type ModelTimestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}
