package models

import "time"

type PrintRequest struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"index"`
	Name          string
	InputType     string
	FilePath      string
	Link          string
	Description   string
	Colour        string
	Comments      string
	Status        string    `gorm:"size:50;index"` // "pending", "in_queue", "printing", "finished", "rejected", "cancelled", "on_hold"
	QueuePosition int       `gorm:"index"`
	CreatedAt     time.Time
}
