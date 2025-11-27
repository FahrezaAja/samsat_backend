package controllers

import (
	"net/http"
	"strings"

	"samsat_backend/config"
	"samsat_backend/models"
	"samsat_backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ======================================================
// REGISTER USER
// ======================================================
func Register(c *gin.Context) {
	var input struct {
		Nama     string `json:"nama" binding:"required"`
		NomorHP  string `json:"nomor_hp" binding:"required"`
		Password string `json:"password" binding:"required"`
		FCMToken string `json:"fcm_token"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, "Input tidak valid: "+err.Error())
		return
	}

	var existing models.User
	if err := config.DB.Where("nomor_hp = ?", input.NomorHP).First(&existing).Error; err == nil {
		utils.Error(c, "Nomor HP sudah terdaftar, silakan login")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Error(c, "Gagal mengenkripsi password")
		return
	}

	user := models.User{
		Nama:     input.Nama,
		NomorHP:  input.NomorHP,
		Password: string(hashedPassword),
		FCMToken: input.FCMToken,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utils.Error(c, "Gagal menyimpan user: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message": "Registrasi berhasil, silakan login",
		"user":    user,
	})
}

// ======================================================
// LOGIN USER (Bearer Token)
// ======================================================
func Login(c *gin.Context) {
	var input struct {
		NomorHP  string `json:"nomor_hp"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid data"})
		return
	}

	// Cari user
	var user models.User
	if err := config.DB.Where("nomor_hp = ?", input.NomorHP).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	// Verifikasi password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	}

	// Generate Bearer Token menggunakan utils.GenerateSimpleToken
	token := utils.GenerateSimpleToken(user.NomorHP, user.ID)

	// Simpan token ke DB
	user.Token = token
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error saving token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data": gin.H{
			"id":        user.ID,
			"nama":      user.Nama,
			"nomor_hp":  user.NomorHP,
			"token":     user.Token,
			"fcm_token": user.FCMToken,
		},
	})
}



// ======================================================
// PROFIL USER BERDASARKAN TOKEN (UNTUK FLUTTER)
// ======================================================
func Profil(c *gin.Context) {
	token := c.Param("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Token wajib diisi",
		})
		return
	}

	var user models.User
	if err := config.DB.Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Data profil tidak ditemukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"nama":      user.Nama,
		"nomor_hp":  user.NomorHP,
		"createdAt": user.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// ======================================
// Helper ambil Bearer Token
// ======================================
func getBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	// Jika tidak ada "Bearer ", anggap seluruh header adalah token
	return authHeader
}
