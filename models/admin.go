package models

import (
    "time"
    "gorm.io/gorm"
)

type Admin struct {
    ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
    Username  string         `json:"username" gorm:"unique;not null"`
    Password  string         `json:"password" gorm:"not null"`
    Nama      string         `json:"nama" gorm:"not null"`
    Role      string         `json:"role" gorm:"default:'admin'"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
