package models

import "time"

type PrintRequest struct {
    ID          uint      `gorm:"primaryKey"`
    Name        string
    InputType   string    // "file", "link", "description"
    FilePath    string
    Link        string
    Description string
    Colour      string
    Comments    string
    CreatedAt   time.Time
}

