package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"uniqueIndex;size:255"`
	Name      string    `gorm:"size:255"`
	Role      string    `gorm:"size:50"` // "student", "admin"
	CreatedAt time.Time
	UpdatedAt time.Time
}
