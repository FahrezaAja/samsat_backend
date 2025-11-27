package config

import (
	"log"
	"samsat_backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "samsat_user@tcp(127.0.0.1:3306)/samsat_db?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi database: ", err)
	}

	DB = database
	log.Println("Database connected")

	err = DB.AutoMigrate(
		&models.PajakKendaraanBermotor{},
		&models.BalikNamaKendaraan{},
		&models.MutasiKendaraan{},
		&models.User{},
		&models.Admin{},
		&models.Status{},
		&models.Booking{},
		&models.SimpleNotification{},
		&models.UserToken{},
		&models.BookingDashboard{},
		&models.Notification{},

	)
	if err != nil {
		log.Fatal("Gagal migrasi database: ", err)
	}
}
