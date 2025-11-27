package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"samsat_backend/config"
	"samsat_backend/models"

	"github.com/gin-gonic/gin"
)

// GET /api/requests/history/:user_id
func GetRequestHistory(c *gin.Context) {
	userIdParam := c.Param("user_id")
	if userIdParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id tidak boleh kosong"})
		return
	}

	// Ambil token dari header Authorization
	token := getBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header wajib diisi"})
		return
	}

	// Cari user berdasarkan token
	var user models.User
	if err := config.DB.Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau user tidak ditemukan"})
		return
	}

	// Pastikan userId dari param = user.ID dari token
	if fmt.Sprintf("%d", user.ID) != userIdParam {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak sesuai dengan user ID"})
		return
	}

	// ================== AMBIL DATA PER TABEL ==================
	var pajak []models.PajakKendaraanBermotor
	var balik []models.BalikNamaKendaraan
	var mutasi []models.MutasiKendaraan

	if err := config.DB.
		Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Find(&pajak).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data pajak"})
		return
	}

	if err := config.DB.
		Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Find(&balik).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data balik nama"})
		return
	}

	if err := config.DB.
		Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Find(&mutasi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil data mutasi"})
		return
	}

	// ================== STRUCT HISTORY ==================
	type HistoryItem struct {
		ID             uint      `json:"id"`
		Nama           string    `json:"nama"`
		NomorKendaraan string    `json:"nomor_kendaraan"`
		NomorHP        string    `json:"nomor_hp"`
		Tanggal        string    `json:"tanggal"` // <- STRING supaya bisa "-" kalau belum selesai
		Kampung        string    `json:"kampung"`
		Catatan        string    `json:"catatan"`
		Status         string    `json:"status"`
		Tim            string    `json:"tim"`
		CatatanAdmin   string    `json:"catatan_admin"`
		Harga          string    `json:"harga"`
		JenisLayanan   string    `json:"jenis_layanan"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	var history []HistoryItem

	formatTanggal := func(status string, tanggal time.Time) string {
		if (status == "Selesai" || status == "Diproses") && !tanggal.IsZero() {
			return tanggal.Format("2006-01-02")
		}
		return "-"
	}

	// ================== PAJAK ==================
	for _, b := range pajak {
		history = append(history, HistoryItem{
			ID:             b.ID,
			Nama:           b.Nama,
			NomorKendaraan: b.NomorKendaraan,
			NomorHP:        b.NomorHP,
			Tanggal:        formatTanggal(b.Status, b.Tanggal),
			Kampung:        b.Kampung,
			Catatan:        b.Catatan,
			Status:         b.Status,
			Tim:            b.Tim,
			CatatanAdmin:   b.CatatanAdmin,
			Harga:          b.Harga,
			JenisLayanan:   "Pajak Kendaraan Bermotor",
			CreatedAt:      b.CreatedAt,
			UpdatedAt:      b.UpdatedAt,
		})
	}

	// ================== BALIK NAMA ==================
	for _, b := range balik {
		history = append(history, HistoryItem{
			ID:             b.ID,
			Nama:           b.Nama,
			NomorKendaraan: b.NomorKendaraan,
			NomorHP:        b.NomorHP,
			Tanggal:        formatTanggal(b.Status, b.Tanggal),
			Kampung:        b.Kampung,
			Catatan:        b.Catatan,
			Status:         b.Status,
			Tim:            b.Tim,
			CatatanAdmin:   b.CatatanAdmin,
			Harga:          b.Harga,
			JenisLayanan:   "Balik Nama Kendaraan",
			CreatedAt:      b.CreatedAt,
			UpdatedAt:      b.UpdatedAt,
		})
	}

	// ================== MUTASI ==================
	for _, b := range mutasi {
		history = append(history, HistoryItem{
			ID:             b.ID,
			Nama:           b.Nama,
			NomorKendaraan: b.NomorKendaraan,
			NomorHP:        b.NomorHP,
			Tanggal:        formatTanggal(b.Status, b.Tanggal),
			Kampung:        b.Kampung,
			Catatan:        b.Catatan,
			Status:         b.Status,
			Tim:            b.Tim,
			CatatanAdmin:   b.CatatanAdmin,
			Harga:          b.Harga,
			JenisLayanan:   "Mutasi Kendaraan",
			CreatedAt:      b.CreatedAt,
			UpdatedAt:      b.UpdatedAt,
		})
	}

	// ================== SORT DESC BY CREATED_AT ==================
	sort.Slice(history, func(i, j int) bool {
		return history[i].CreatedAt.After(history[j].CreatedAt)
	})

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"total":    len(history),
		"requests": history,
	})
}
