package models

import (
	"time"
)

// Model SimpleNotification untuk notifikasi sederhana
type SimpleNotification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NomorHP   string    `gorm:"size:255;not null;index" json:"nomor_hp"`
	Status    string    `gorm:"size:255;not null" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
