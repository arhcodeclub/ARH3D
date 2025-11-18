package models

import "time"

type Filament struct {
	ID        uint      `gorm:"primaryKey"`
	Type      string    `gorm:"size:100"`
	Colour    string    `gorm:"size:100"`
	Hex       string    `gorm:"size:7"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
