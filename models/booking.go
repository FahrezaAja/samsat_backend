package models

import (
	"time"

	"gorm.io/gorm"
)

// PajakKendaraanBermotor model
type PajakKendaraanBermotor struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Nama           string    `gorm:"type:varchar(255)" json:"nama"`
	NomorKendaraan string    `gorm:"type:varchar(50)" json:"nomor_kendaraan"`
	NomorHP        string    `gorm:"type:varchar(20)" json:"nomor_hp"`
	Tanggal        time.Time `gorm:"type:date" json:"tanggal"`
	Kampung        string    `gorm:"type:varchar(255)" json:"kampung"`
	Catatan        string    `gorm:"type:text" json:"catatan"`
	Status         string    `gorm:"type:varchar(50)" json:"status"`
	Tim            string    `gorm:"type:varchar(100)" json:"tim"`
	CatatanAdmin   string    `gorm:"type:text" json:"catatan_admin"`
	Harga          string    `gorm:"type:varchar(50)" json:"harga"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	UserID         uint      `gorm:"column:user_id" json:"user_id"`
}

// BalikNamaKendaraan model
type BalikNamaKendaraan struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Nama           string    `gorm:"type:varchar(255)" json:"nama"`
	NomorKendaraan string    `gorm:"type:varchar(50)" json:"nomor_kendaraan"`
	NomorHP        string    `gorm:"type:varchar(20)" json:"nomor_hp"`
	Tanggal        time.Time `gorm:"type:date" json:"tanggal"`
	Kampung        string    `gorm:"type:varchar(255)" json:"kampung"`
	Catatan        string    `gorm:"type:text" json:"catatan"`
	Status         string    `gorm:"type:varchar(50)" json:"status"`
	Tim            string    `gorm:"type:varchar(100)" json:"tim"`
	CatatanAdmin   string    `gorm:"type:text" json:"catatan_admin"`
	Harga          string    `gorm:"type:varchar(50)" json:"harga"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	UserID         uint      `gorm:"column:user_id" json:"user_id"`
}

// MutasiKendaraan model
type MutasiKendaraan struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Nama           string    `gorm:"type:varchar(255)" json:"nama"`
	NomorKendaraan string    `gorm:"type:varchar(50)" json:"nomor_kendaraan"`
	NomorHP        string    `gorm:"type:varchar(20)" json:"nomor_hp"`
	Tanggal        time.Time `gorm:"type:date" json:"tanggal"`
	Kampung        string    `gorm:"type:varchar(255)" json:"kampung"`
	Catatan        string    `gorm:"type:text" json:"catatan"`
	Status         string    `gorm:"type:varchar(50)" json:"status"`
	Tim            string    `gorm:"type:varchar(100)" json:"tim"`
	CatatanAdmin   string    `gorm:"type:text" json:"catatan_admin"`
	Harga          string    `gorm:"type:varchar(50)" json:"harga"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	UserID         uint      `gorm:"column:user_id" json:"user_id"`
}

type Booking struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Nama           string         `json:"nama"`
	NomorKendaraan string         `json:"nomor_kendaraan"`
	NomorHp        string         `json:"nomor_hp"`
	JenisLayanan   string         `json:"jenis_layanan"`
	Tanggal        string         `json:"tanggal"`
	Status         string         `json:"status"`
	CatatanAdmin   string         `json:"catatan_admin"`
	Tim            string         `json:"tim"`
	Harga          string            `json:"harga"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type DashboardStats struct {
    TotalKendaraan   int `json:"total_kendaraan"`
    TotalPembayaran  int `json:"total_pembayaran"`
    TotalPengguna    int `json:"total_pengguna"`
}

type BalikNama struct {
	ID             uint   `json:"id"`
	NomorKendaraan string `json:"nomor_kendaraan"`
	PemilikBaru    string `json:"pemilik_baru"`
	NomorHP        string `json:"nomor_hp"`
	Status         string `json:"status"`
}

type Mutasi struct {
	ID             uint   `json:"id"`
	NomorKendaraan string `json:"nomor_kendaraan"`
	Nama           string `json:"nama"`
	NomorHP        string `json:"nomor_hp"`
	Status         string `json:"status"`
}

type BookingDashboard struct {
	ID             uint      `json:"id"`
	NomorKendaraan string    `json:"nomor_kendaraan"`
	Nama           string    `json:"nama"`
	NomorHP        string    `json:"nomor_hp"`
	Status         string    `json:"status"`
	Tanggal        time.Time `json:"tanggal"`
}

func (PajakKendaraanBermotor) TableName() string {
	return "pajak_kendaraan_bermotor"
}

func (BalikNamaKendaraan) TableName() string {
	return "balik_nama_kendaraan"
}

func (MutasiKendaraan) TableName() string {
	return "mutasi_kendaraan"
}

func (Booking) TableName() string {
	return "bookings"
}
