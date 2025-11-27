package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Nama     string `json:"nama" gorm:"type:varchar(100)"`
	NomorHP  string `json:"nomor_hp" gorm:"type:varchar(15);unique"`
	Password string `json:"password" gorm:"type:varchar(255)"`
	FCMToken string `json:"fcm_token" gorm:"type:varchar(255)"`
	Token    string `json:"token" gorm:"type:varchar(255)"` // token terbaru
    Pajaks     []PajakKendaraanBermotor `gorm:"foreignKey:UserID"`
    BalikNamas []BalikNamaKendaraan      `gorm:"foreignKey:UserID"`
    Mutasis    []MutasiKendaraan         `gorm:"foreignKey:UserID"`
}
