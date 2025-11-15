package models

import "time"

type LoginToken struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"index;size:255"`
	TokenHash string    `gorm:"uniqueIndex;size:255"`
	ExpiresAt time.Time `gorm:"index"`
	UsedAt    *time.Time
	CreatedAt time.Time
}
