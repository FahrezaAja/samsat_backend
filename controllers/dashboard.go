package controllers

import (
	"net/http"
	"samsat_backend/config"
	"samsat_backend/models"

	"github.com/gin-gonic/gin"
)

func DashboardStats(c *gin.Context) {
	var bookings []models.BookingDashboard
	if err := config.DB.Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data booking"})
		return
	}

	totalKendaraan := make(map[string]bool)
	totalPembayaran := 0
	totalPengguna := make(map[string]bool)

	for _, b := range bookings {
		if b.NomorKendaraan != "" {
			totalKendaraan[b.NomorKendaraan] = true
		}
		if b.Status == "Selesai" {
			totalPembayaran++
		}
		if b.NomorHP != "" {
			totalPengguna[b.NomorHP] = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"total_kendaraan": len(totalKendaraan),
			"total_pembayaran": totalPembayaran,
			"total_pengguna": len(totalPengguna),
		},
	})
}

func GetDashboardStats(c *gin.Context) {
    // ===== Ambil semua data =====
    var bookings []models.Booking
    var baliknama []models.BalikNama
    var mutasi []models.Mutasi

    config.DB.Find(&bookings)
    config.DB.Find(&baliknama)
    config.DB.Find(&mutasi)

    // ===== Hitung Total Kendaraan Unik =====
    kendaraanMap := make(map[string]bool)

    for _, b := range bookings {
        kendaraanMap[b.NomorKendaraan] = true
    }
    for _, bn := range baliknama {
        kendaraanMap[bn.NomorKendaraan] = true
    }
    for _, m := range mutasi {
        kendaraanMap[m.NomorKendaraan] = true
    }

    totalKendaraan := len(kendaraanMap)

    // ===== Hitung Total Pembayaran (status = "Selesai") =====
    totalPembayaran := 0
    for _, b := range bookings {
        if b.Status == "Selesai" {
            totalPembayaran++
        }
    }

    // ===== Hitung Pengguna Terdaftar (nomor_hp UNIQUE) =====
    userMap := make(map[string]bool)

    for _, b := range bookings {
        userMap[b.NomorHp] = true
    }
    for _, bn := range baliknama {
        userMap[bn.NomorHP] = true
    }
    for _, m := range mutasi {
        userMap[m.NomorHP] = true
    }

    totalPengguna := len(userMap)

    // ===== Response =====
    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data": gin.H{
            "total_kendaraan":   totalKendaraan,
            "total_pembayaran":  totalPembayaran,
            "total_pengguna":    totalPengguna,
        },
    })
}