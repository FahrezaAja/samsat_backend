package controllers

import (
	"fmt"
	"net/http"
	
	"samsat_backend/config"
	"samsat_backend/models"
	"samsat_backend/utils"

	"github.com/gin-gonic/gin"
)

// Ambil user berdasarkan bearer token di header
func getUserFromBearer(c *gin.Context) (*models.User, error) {
	token := getBearerToken(c)
	if token == "" {
		return nil, fmt.Errorf("Authorization header wajib diisi")
	}

	var user models.User
	if err := config.DB.Where("token = ?", token).First(&user).Error; err != nil {
		return nil, fmt.Errorf("Token tidak valid atau user tidak ditemukan")
	}

	return &user, nil
}

// ======================================================
// PROFIL USER BERDASARKAN BEARER TOKEN (HALAMAN PROFIL)
// ======================================================
func Me(c *gin.Context) {
	user, err := getUserFromBearer(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Data profil berhasil diambil",
		"data": gin.H{
			"id":       user.ID,
			"nama":     user.Nama,
			"nomor_hp": user.NomorHP,
		},
	})
}

// ======================================================
// LOGOUT USER (Bearer Token)
// ======================================================
func Logout(c *gin.Context) {
	user, err := getUserFromBearer(c)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	user.Token = ""
	if err := config.DB.Save(user).Error; err != nil {
		utils.Error(c, "Gagal logout, silakan coba lagi")
		return
	}

	utils.Success(c, gin.H{
		"message": "Logout berhasil",
	})
}

// ======================================================
// UPDATE PROFIL USER (NAMA & NOMOR HP)
// ======================================================
func UpdateProfile(c *gin.Context) {
	// Ambil user dari Bearer token
	user, err := getUserFromBearer(c)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	// Struktur input dari request body
	var input struct {
		Nama    string `json:"nama" binding:"required"`
		NomorHP string `json:"nomor_hp" binding:"required"`
	}

	// Validasi JSON yang dikirim client
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, "Input tidak valid: "+err.Error())
		return
	}

	// Cek apakah nomor HP baru sudah dipakai user lain
	var existing models.User
	if err := config.DB.
		Where("nomor_hp = ? AND id <> ?", input.NomorHP, user.ID).
		First(&existing).Error; err == nil {
		// Kalau err == nil berarti ada user lain dengan nomor HP itu
		utils.Error(c, "Nomor HP sudah digunakan oleh pengguna lain")
		return
	}

	// Update field di struct user
	user.Nama = input.Nama
	user.NomorHP = input.NomorHP

	// Simpan ke database
	if err := config.DB.Save(user).Error; err != nil {
		utils.Error(c, "Gagal mengupdate profil: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message": "Profil berhasil diperbarui",
		"data": gin.H{
			"id":       user.ID,
			"nama":     user.Nama,
			"nomor_hp": user.NomorHP,
		},
	})
}
