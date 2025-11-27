package controllers

import (
    "net/http"
    "samsat_backend/config"
    "samsat_backend/models"
    "golang.org/x/crypto/bcrypt"
    "github.com/gin-gonic/gin"
)

// ✅ Ambil semua admin
func GetAllAdmin(c *gin.Context) {
    var admins []models.Admin
    if err := config.DB.Find(&admins).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data admin"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status": "success",
        "data":   admins,
    })
}
func CreateAdmin(c *gin.Context) {
    var input models.Admin

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid", "detail": err.Error()})
        return
    }

    // ✅ Cek field kosong
    if input.Username == "" || input.Password == "" || input.Nama == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Semua field wajib diisi"})
        return
    }

    // ✅ Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal hash password"})
        return
    }
    input.Password = string(hashedPassword)

    // ✅ Simpan ke database
    if err := config.DB.Create(&input).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan ke database"})
        return
    }

    // ✅ Response sukses
    c.JSON(http.StatusOK, gin.H{
        "status":  "success",
        "message": "User admin baru berhasil ditambahkan",
        "data":    input,
    })
}

func AdminLogin(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input tidak valid"})
		return
	}

	var admin models.Admin
	if err := config.DB.Where("username = ?", input.Username).First(&admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username tidak ditemukan"})
		return
	}

	// 🔹 Normalisasi prefix bcrypt agar Go bisa membaca hash dari Laravel
	hash := []byte(admin.Password)
	if len(hash) > 3 && hash[2] == 'y' {
		hash[2] = 'a' // ubah $2y$ → $2a$
	}

	// 🔹 Bandingkan password hash bcrypt
	if err := bcrypt.CompareHashAndPassword(hash, []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password salah"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login berhasil",
		"data": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
			"nama":     admin.Nama,
			"role":     admin.Role,
		},
	})
}