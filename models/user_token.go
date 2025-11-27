package models

import (
	"time"

	"gorm.io/gorm"
)

type UserToken struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	NomorHP   string         `gorm:"type:varchar(20);unique" json:"nomor_hp"`
	FCMToken  string         `gorm:"type:varchar(255)" json:"fcm_token"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserToken) TableName() string {
	return "user_tokens"
}
