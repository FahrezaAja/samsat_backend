package models

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    string         `json:"user_id"`
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	Type      string         `json:"type"` // info, success, failed, etc.
	IsRead    bool           `json:"is_read"`
	Tanggal   *time.Time     `gorm:"type:date" json:"tanggal"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
}
