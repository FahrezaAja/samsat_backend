package controllers

import (
	"fmt"
	"net/http"
	"samsat_backend/config"
	"samsat_backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

// GET /api/notifications/:user_id
func GetNotifications(c *gin.Context) {
	userId := c.Param("user_id")
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId tidak boleh kosong"})
		return
	}

	// Verifikasi token untuk mendapatkan user ID
	token := getBearerToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header wajib diisi"})
		return
	}

	var user models.User
	if err := config.DB.Where("token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau user tidak ditemukan"})
		return
	}

	// Pastikan userId dari param sama dengan user ID dari token
	if fmt.Sprintf("%d", user.ID) != userId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak sesuai dengan user ID"})
		return
	}

	var notifications []models.Notification
	err := config.DB.Where("user_id = ? AND is_read = ?", userId, false).Order("created_at desc").Find(&notifications).Error
	if err != nil || len(notifications) == 0 {
		// Tidak ada notifikasi baru
		c.JSON(http.StatusOK, gin.H{
			"has_new":      false,
			"notification": nil,
		})
		return
	}

	latestNotif := notifications[0]

	var tanggalStr string
	if latestNotif.Tanggal != nil {
		tanggalStr = latestNotif.Tanggal.Format("2006-01-02")
	} else {
		tanggalStr = ""
	}

	notifData := gin.H{
		"id":      latestNotif.ID,
		"title":   latestNotif.Title,
		"message": latestNotif.Message,
		"type":    latestNotif.Type,
		"tanggal": tanggalStr, // <-- ini yang dipakai Flutter
	}

	c.JSON(http.StatusOK, gin.H{
		"has_new":      true,
		"notification": notifData,
	})
}

// POST /api/notifications/mark-read
func MarkNotificationRead(c *gin.Context) {
	var body struct {
		UserID         string `json:"id"`
		NotificationID string `json:"notification_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload tidak valid"})
		return
	}

	// Tidak perlu verifikasi token untuk mark-read sesuai dengan implementasi Flutter
	var notification models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", body.NotificationID, body.UserID).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notifikasi tidak ditemukan"})
		return
	}

	notification.IsRead = true
	notification.UpdatedAt = time.Now()
	config.DB.Save(&notification)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Fungsi opsional: buat notifikasi baru (testing)
func CreateNotification(c *gin.Context) {
	var body struct {
		UserID  string `json:"user_id"`
		Title   string `json:"title"`
		Message string `json:"message"`
		Type    string `json:"type"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload tidak valid"})
		return
	}

	notification := models.Notification{
		UserID:  body.UserID,
		Title:   body.Title,
		Message: body.Message,
		Type:    body.Type,
		IsRead:  false,
	}

	config.DB.Create(&notification)
	c.JSON(http.StatusOK, gin.H{"status": "success", "notification_id": notification.ID})
}

// GET /api/notifications
// Ambil SEMUA notifikasi berdasarkan ID user yang login (dari token)
func GetAllNotifications(c *gin.Context) {
	// Ambil token dari header Authorization: Bearer <token>
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

	// Notification.UserID bertipe string, jadi kita samakan formatnya
	userIDStr := fmt.Sprintf("%d", user.ID)

	var notifications []models.Notification
	if err := config.DB.
		Where("user_id = ?", userIDStr).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengambil notifikasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"total":         len(notifications),
		"notifications": notifications,
	})
}

// GET /api/notifications/recent/:user_id
// Mengambil semua riwayat notifikasi milik user tertentu (berdasarkan ID)
// dan tetap memverifikasi bahwa token yang dipakai adalah milik user tersebut.
func GetUserNotificationsRecent(c *gin.Context) {
	userIdParam := c.Param("user_id")
	if userIdParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id tidak boleh kosong"})
		return
	}

	// Ambil token dari header
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

	// Karena Notification.UserID bertipe string, samakan format
	userIDStr := fmt.Sprintf("%d", user.ID)

	// Ambil semua notifikasi milik user ini saja (riwayat)
	var notifications []models.Notification
	if err := config.DB.
		Where("user_id = ?", userIDStr).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil notifikasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"total":         len(notifications),
		"notifications": notifications,
	})
}

// DELETE /api/notifications/:id
// Menghapus satu notifikasi milik user yang sedang login (berdasarkan token)
func DeleteNotification(c *gin.Context) {
	notifID := c.Param("id")
	if notifID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification id tidak boleh kosong"})
		return
	}

	// Ambil token dari header
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

	// Karena Notification.UserID bertipe string, samakan format
	userIDStr := fmt.Sprintf("%d", user.ID)

	// Pastikan notifikasi ini memang milik user tersebut
	var notif models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", notifID, userIDStr).First(&notif).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notifikasi tidak ditemukan atau bukan milik user ini"})
		return
	}

	// Hapus notifikasi
	if err := config.DB.Delete(&notif).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menghapus notifikasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "success",
		"deleted_id":      notifID,
		"deleted_message": "Notifikasi berhasil dihapus",
	})
}
