package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"samsat_backend/config"
	"samsat_backend/models"
	"samsat_backend/utils"
	"strings"
	"sync"
	"time"

	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

// ======================================================
// FCM CONFIG & STORAGE TOKEN SEMENTARA
// ======================================================
var FCMServerKey = "997598832883"

// userTokensByPhone menyimpan token FCM per nomor HP user.
// key: normalized nomor HP (hilangkan spasi), value: map[token]bool untuk deduplikasi
var userTokensByPhone = map[string]map[string]bool{}
var userTokensMu sync.RWMutex

// Fungsi untuk menambahkan token ke nomor HP tertentu (deduplikasi)
func addTokenForPhone(nomorHP, token string) {
	norm := normalizePhone(nomorHP)
	if norm == "" || token == "" {
		return
	}
	userTokensMu.Lock()
	defer userTokensMu.Unlock()
	if _, ok := userTokensByPhone[norm]; !ok {
		userTokensByPhone[norm] = map[string]bool{}
	}
	userTokensByPhone[norm][token] = true
	fmt.Println("Token tersimpan untuk", norm, "=>", token)
}

// Ambil list token untuk nomor HP tertentu
func getTokensForPhone(nomorHP string) []string {
	norm := normalizePhone(nomorHP)
	if norm == "" {
		return []string{}
	}
	userTokensMu.RLock()
	defer userTokensMu.RUnlock()
	m, ok := userTokensByPhone[norm]
	if !ok {
		return []string{}
	}
	tokens := []string{}
	for t := range m {
		tokens = append(tokens, t)
	}
	return tokens
}

// Normalisasi nomor HP: hapus spasi dan trim
func normalizePhone(ph string) string {
	return strings.TrimSpace(strings.ReplaceAll(ph, " ", ""))
}

// ======================================================
// FUNGSI KIRIM NOTIFIKASI via FCM
// ======================================================
func sendNotification(token, title, body string) {
	payload := map[string]interface{}{
		"to": token,
		"notification": map[string]string{
			"title": title,
			"body":  body,
		},
		"data": map[string]string{
			// Bisa tambahkan payload data kustom di sini
			"click_action": "FLUTTER_NOTIFICATION_CLICK",
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Authorization", "key="+FCMServerKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ Gagal kirim FCM:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("✅ Notifikasi terkirim ke token:", token, "| Status:", resp.Status)
}

// ======================================================
// REGISTER TOKEN DARI FLUTTER
// Endpoint: POST /register-token
// Body JSON: { "token": "...", "nomor_hp": "0812..." }
// ======================================================
func RegisterToken(c *gin.Context) {
	var body struct {
		Token   string `json:"token"`
		NomorHP string `json:"nomor_hp"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Error(c, "Payload tidak valid: "+err.Error())
		return
	}

	if body.Token == "" {
		utils.Error(c, "Token tidak boleh kosong")
		return
	}

	phone := strings.TrimSpace(body.NomorHP)
	if phone == "" {
		utils.Error(c, "Nomor HP tidak boleh kosong")
		return
	}

	// Simpan ke database
	userToken := models.UserToken{
		NomorHP:  phone,
		FCMToken: body.Token,
	}
	if err := config.DB.Where("nomor_hp = ?", phone).Assign(models.UserToken{FCMToken: body.Token}).FirstOrCreate(&userToken).Error; err != nil {
		utils.Error(c, "Gagal menyimpan token: "+err.Error())
		return
	}

	// Juga simpan ke in-memory untuk akses cepat
	addTokenForPhone(phone, body.Token)
	utils.Success(c, "Token berhasil disimpan")
}

type BookingInput struct {
	Nama           string `json:"nama"`
	NomorKendaraan string `json:"nomor_kendaraan"`
	NomorHP        string `json:"nomor_hp"`
	JenisLayanan   string `json:"jenis_layanan"`
	Kampung        string `json:"kampung"`
	Catatan        string `json:"catatan"`
	UserID         uint   `json:"user_id"`
}

func normalizeService(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// ubah beberapa variasi umum supaya match
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "  ", " ")
	return s
}
func CreateBooking(c *gin.Context) {
	// Verifikasi token untuk mendapatkan user ID
	token := getBearerToken(c)
	if token == "" {
		utils.Error(c, "Authorization header wajib diisi")
		return
	}

	var user models.User
	if err := config.DB.Where("token = ?", token).First(&user).Error; err != nil {
		utils.Error(c, "Token tidak valid atau user tidak ditemukan")
		return
	}

	var input BookingInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jenis := normalizeService(input.JenisLayanan)

	if strings.Contains(jenis, "pajak") {
		data := models.PajakKendaraanBermotor{
			Nama:           input.Nama,
			NomorKendaraan: input.NomorKendaraan,
			NomorHP:        input.NomorHP,
			// Tanggal:     JANGAN DIISI DI SINI
			Kampung: input.Kampung,
			Catatan: input.Catatan,
			Status:  "Pending",
			UserID:  user.ID,
		}

		if err := config.DB.Create(&data).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, gin.H{"message": "Booking berhasil", "data": data})
		return
	} else if strings.Contains(jenis, "balik nama") {
		data := models.BalikNamaKendaraan{
			Nama:           input.Nama,
			NomorKendaraan: input.NomorKendaraan,
			NomorHP:        input.NomorHP,
			// Tanggal:     JANGAN DIISI DI SINI
			Kampung: input.Kampung,
			Catatan: input.Catatan,
			Status:  "Pending",
			UserID:  user.ID,
		}

		if err := config.DB.Create(&data).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, gin.H{"message": "Booking berhasil", "data": data})
		return
	} else if strings.Contains(jenis, "mutasi") {
		data := models.MutasiKendaraan{
			Nama:           input.Nama,
			NomorKendaraan: input.NomorKendaraan,
			NomorHP:        input.NomorHP,
			// Tanggal:     JANGAN DIISI DI SINI
			Kampung: input.Kampung,
			Catatan: input.Catatan,
			Status:  "Pending",
			UserID:  user.ID,
		}

		if err := config.DB.Create(&data).Error; err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, gin.H{"message": "Booking berhasil", "data": data})
		return
	} else {
		utils.Error(c, "Jenis layanan tidak valid")
		return
	}
}

// ======================================================
// SEARCH BOOKING
// ======================================================
func SearchBooking(c *gin.Context) {
	nomorKendaraan := strings.TrimSpace(c.Query("nomor_kendaraan"))
	nama := strings.TrimSpace(c.Query("nama"))
	nomorHP := strings.TrimSpace(c.Query("nomor_hp"))

	if nomorKendaraan == "" && nomorHP == "" {
		utils.Error(c, "Harap isi nomor kendaraan atau nomor HP untuk pencarian")
		return
	}

	normNomorKendaraan := strings.ToLower(strings.ReplaceAll(nomorKendaraan, " ", ""))
	normNomorHP := strings.ToLower(strings.ReplaceAll(nomorHP, " ", ""))

	if len(normNomorKendaraan) > 0 && len(normNomorKendaraan) < 5 {
		utils.Error(c, "Nomor kendaraan harus diisi lengkap (tidak boleh sebagian)")
		return
	}
	if len(normNomorHP) > 0 && len(normNomorHP) < 5 {
		utils.Error(c, "Nomor HP harus diisi lengkap (tidak boleh sebagian)")
		return
	}

	var pajak []models.PajakKendaraanBermotor
	var balik []models.BalikNamaKendaraan
	var mutasi []models.MutasiKendaraan

	queryPajak := config.DB.Model(&models.PajakKendaraanBermotor{})
	if normNomorKendaraan != "" {
		queryPajak = queryPajak.Where("LOWER(REPLACE(nomor_kendaraan, ' ', '')) = ?", normNomorKendaraan)
	}
	if normNomorHP != "" {
		queryPajak = queryPajak.Where("LOWER(REPLACE(nomor_hp, ' ', '')) = ?", normNomorHP)
	}
	if nama != "" {
		queryPajak = queryPajak.Where("nama LIKE ?", "%"+nama+"%")
	}
	queryPajak.Find(&pajak)

	queryBalik := config.DB.Model(&models.BalikNamaKendaraan{})
	if normNomorKendaraan != "" {
		queryBalik = queryBalik.Where("LOWER(REPLACE(nomor_kendaraan, ' ', '')) = ?", normNomorKendaraan)
	}
	if normNomorHP != "" {
		queryBalik = queryBalik.Where("LOWER(REPLACE(nomor_hp, ' ', '')) = ?", normNomorHP)
	}
	if nama != "" {
		queryBalik = queryBalik.Where("nama LIKE ?", "%"+nama+"%")
	}
	queryBalik.Find(&balik)

	queryMutasi := config.DB.Model(&models.MutasiKendaraan{})
	if normNomorKendaraan != "" {
		queryMutasi = queryMutasi.Where("LOWER(REPLACE(nomor_kendaraan, ' ', '')) = ?", normNomorKendaraan)
	}
	if normNomorHP != "" {
		queryMutasi = queryMutasi.Where("LOWER(REPLACE(nomor_hp, ' ', '')) = ?", normNomorHP)
	}
	if nama != "" {
		queryMutasi = queryMutasi.Where("nama LIKE ?", "%"+nama+"%")
	}
	queryMutasi.Find(&mutasi)

	results := []map[string]interface{}{}
	appendResults := func(data interface{}, jenis string) {
		switch v := data.(type) {
		case []models.PajakKendaraanBermotor:
			for _, b := range v {
				tanggalDisplay := "-"
				if !b.Tanggal.IsZero() {
					tanggalDisplay = b.Tanggal.Format("2006-01-02")
				}
				results = append(results, map[string]interface{}{
					"id":              b.ID,
					"nama":            b.Nama,
					"nomor_kendaraan": b.NomorKendaraan,
					"nomor_hp":        b.NomorHP,
					"tanggal":         tanggalDisplay,
					"kampung":         b.Kampung,
					"catatan":         b.Catatan,
					"status":          b.Status,
					"tim":             b.Tim,
					"catatan_admin":   b.CatatanAdmin,
					"harga":           b.Harga,
					"jenis_layanan":   jenis,
				})
			}
		case []models.BalikNamaKendaraan:
			for _, b := range v {
				tanggalDisplay := "-"
				if !b.Tanggal.IsZero() {
					tanggalDisplay = b.Tanggal.Format("2006-01-02")
				}
				results = append(results, map[string]interface{}{
					"id":              b.ID,
					"nama":            b.Nama,
					"nomor_kendaraan": b.NomorKendaraan,
					"nomor_hp":        b.NomorHP,
					"tanggal":         tanggalDisplay,
					"kampung":         b.Kampung,
					"catatan":         b.Catatan,
					"status":          b.Status,
					"tim":             b.Tim,
					"catatan_admin":   b.CatatanAdmin,
					"harga":           b.Harga,
					"jenis_layanan":   jenis,
				})
			}
		case []models.MutasiKendaraan:
			for _, b := range v {
				tanggalDisplay := "-"
				if !b.Tanggal.IsZero() {
					tanggalDisplay = b.Tanggal.Format("2006-01-02")
				}
				results = append(results, map[string]interface{}{
					"id":              b.ID,
					"nama":            b.Nama,
					"nomor_kendaraan": b.NomorKendaraan,
					"nomor_hp":        b.NomorHP,
					"tanggal":         tanggalDisplay,
					"kampung":         b.Kampung,
					"catatan":         b.Catatan,
					"status":          b.Status,
					"tim":             b.Tim,
					"catatan_admin":   b.CatatanAdmin,
					"harga":           b.Harga,
					"jenis_layanan":   jenis,
				})
			}
		}
	}

	appendResults(pajak, "Pajak Kendaraan Bermotor")
	appendResults(balik, "Balik Nama Kendaraan")
	appendResults(mutasi, "Mutasi Kendaraan")

	if len(results) == 0 {
		utils.Error(c, "Data tidak ditemukan. Pastikan nomor kendaraan atau nomor HP diisi lengkap dan benar (tanpa salah ketik).")
		return
	}

	c.JSON(200, gin.H{
		"status": "success",
		"data":   results,
	})
}

// ======================================================
// Fungsi Inisialisasi Firebase Client
// ======================================================
func initializeFirebase() (*messaging.Client, error) {
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile("config/firebase-adminsdk.json"))
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error initializing FCM client: %v", err)
	}

	return fcmClient, nil
}

// Fungsi untuk mendapatkan token FCM berdasarkan nomor HP
func getFCMTokenByPhone(phoneNumber string) string {
	var userToken models.UserToken

	// pakai config.DB bukan db yang nil
	result := config.DB.Where("nomor_hp = ?", phoneNumber).First(&userToken)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Println("Token FCM tidak ditemukan untuk nomor HP:", phoneNumber)
		} else {
			log.Println("Error querying database:", result.Error)
		}
		return ""
	}

	return userToken.FCMToken
}

// ======================================================
// Fungsi Kirim Notifikasi FCM
// ======================================================
func sendNotificationToPhone(phoneNumber string, title string, body string) {
	// Mendapatkan token FCM untuk nomor HP yang diberikan
	userToken := getFCMTokenByPhone(phoneNumber)

	if userToken == "" {
		log.Println("No FCM token found for phone number:", phoneNumber)
		return
	}

	// Membuat pesan FCM
	message := &messaging.Message{
		Token: userToken,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	// Kirim notifikasi
	fcmClient, err := initializeFirebase()
	if err != nil {
		log.Println("Error initializing FCM:", err)
		return
	}

	_, err = fcmClient.Send(context.Background(), message)
	if err != nil {
		log.Println("Error sending FCM notification:", err)
		return
	}

	log.Println("FCM notification sent to:", phoneNumber)
}

// ======================================================
// UPDATE BOOKING PAJAK + KIRIM NOTIFIKASI KE USER TERKAIT
// ======================================================
func UpdateBooking(c *gin.Context) {
	id := c.Param("id")

	var booking models.PajakKendaraanBermotor
	if err := config.DB.First(&booking, id).Error; err != nil {
		utils.Error(c, "Data booking tidak ditemukan")
		return
	}

	var input struct {
		Tanggal      string `json:"tanggal"`
		Tim          string `json:"tim"`
		CatatanAdmin string `json:"catatan_admin"`
		Harga        string `json:"harga"`
		Status       string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, "Input tidak valid: "+err.Error())
		return
	}

	// Parsing tanggal
	if input.Tanggal != "" {
		parsedDate, err := time.Parse("2006-01-02", input.Tanggal)
		if err != nil {
			utils.Error(c, "Format tanggal harus YYYY-MM-DD")
			return
		}
		booking.Tanggal = parsedDate
	} else if booking.Tanggal.IsZero() {
		booking.Tanggal = time.Now()
	}

	if input.Tim != "" {
		booking.Tim = input.Tim
	}
	if input.CatatanAdmin != "" {
		booking.CatatanAdmin = input.CatatanAdmin
	}
	if input.Harga != "" {
		booking.Harga = input.Harga
	}
	if input.Status != "" {
		booking.Status = input.Status
	}

	if err := config.DB.Save(&booking).Error; err != nil {
		utils.Error(c, "Gagal menyimpan perubahan: "+err.Error())
		return
	}

	notification := models.Notification{
		UserID:  fmt.Sprintf("%d", booking.UserID),
		Title:   "Status Diperbarui",
		Message: fmt.Sprintf("Pesanan atas nama %s kini berstatus: %s", booking.Nama, booking.Status),
		Type:    "",
		IsRead:  false,
		Tanggal: &booking.Tanggal,
	}

	if err := config.DB.Create(&notification).Error; err != nil {
		fmt.Println("Error creating notification:", err)
	}

	sendNotificationToPhone(booking.NomorHP, "Status Diperbarui", fmt.Sprintf("Pesanan atas nama %s kini berstatus: %s", booking.Nama, booking.Status))

	utils.Success(c, gin.H{
		"message": "Booking berhasil diperbarui dan notifikasi dikirim",
		"data":    booking,
	})
}

// ======================================================
// UPDATE BOOKING BALIK NAMA + KIRIM NOTIFIKASI KE USER TERKAIT
// ======================================================
func UpdateBalikNama(c *gin.Context) {
	id := c.Param("id")

	var booking models.BalikNamaKendaraan
	if err := config.DB.First(&booking, id).Error; err != nil {
		utils.Error(c, "Data booking tidak ditemukan")
		return
	}

	var input struct {
		Tanggal      string `json:"tanggal"`
		Tim          string `json:"tim"`
		CatatanAdmin string `json:"catatan_admin"`
		Harga        string `json:"harga"`
		Status       string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, "Input tidak valid: "+err.Error())
		return
	}

	if input.Tanggal != "" {
		parsedDate, err := time.Parse("2006-01-02", input.Tanggal)
		if err != nil {
			utils.Error(c, "Format tanggal harus YYYY-MM-DD")
			return
		}
		booking.Tanggal = parsedDate
	} else if booking.Tanggal.IsZero() {
		booking.Tanggal = time.Now()
	}

	if input.Tim != "" {
		booking.Tim = input.Tim
	}
	if input.CatatanAdmin != "" {
		booking.CatatanAdmin = input.CatatanAdmin
	}
	if input.Harga != "" {
		booking.Harga = input.Harga
	}
	if input.Status != "" {
		booking.Status = input.Status
	}

	if err := config.DB.Save(&booking).Error; err != nil {
		utils.Error(c, "Gagal menyimpan perubahan: "+err.Error())
		return
	}

	notification := models.Notification{
		UserID:  fmt.Sprintf("%d", booking.UserID),
		Title:   "Status Diperbarui",
		Message: fmt.Sprintf("Pesanan atas nama %s kini berstatus: %s", booking.Nama, booking.Status),
		Type:    "",
		IsRead:  false,
		Tanggal: &booking.Tanggal,
	}

	if err := config.DB.Create(&notification).Error; err != nil {
		fmt.Println("Error creating notification:", err)
	}

	sendNotificationToPhone(booking.NomorHP, "Status Diperbarui", fmt.Sprintf("Pesanan atas nama %s kini berstatus: %s", booking.Nama, booking.Status))

	utils.Success(c, gin.H{
		"message": "Booking berhasil diperbarui dan notifikasi dikirim",
		"data":    booking,
	})
}

// ======================================================
// GET SEMUA DATA BALIK NAMA
// ======================================================
func GetAllBalikNama(c *gin.Context) {
	var bookings []models.BalikNamaKendaraan

	if err := config.DB.Find(&bookings).Error; err != nil {
		utils.Error(c, "Gagal mengambil data dari database: "+err.Error())
		return
	}

	results := []map[string]interface{}{}
	for _, b := range bookings {
		tanggalDisplay := "-"
		if !b.Tanggal.IsZero() {
			tanggalDisplay = b.Tanggal.Format("2006-01-02")
		}

		results = append(results, map[string]interface{}{
			"id":              b.ID,
			"nama":            b.Nama,
			"nomor_kendaraan": b.NomorKendaraan,
			"nomor_hp":        b.NomorHP,
			"tanggal":         tanggalDisplay,
			"kampung":         b.Kampung,
			"catatan":         b.Catatan,
			"status":          b.Status,
			"tim":             b.Tim,
			"catatan_admin":   b.CatatanAdmin,
			"harga":           b.Harga,
			"jenis_layanan":   "Balik Nama Kendaraan",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   results,
	})
}

// ======================================================
// GET SEMUA DATA PAJAK
// ======================================================
func GetAllBookings(c *gin.Context) {
	var bookings []models.PajakKendaraanBermotor

	if err := config.DB.Find(&bookings).Error; err != nil {
		utils.Error(c, "Gagal mengambil data dari database: "+err.Error())
		return
	}

	results := []map[string]interface{}{}
	for _, b := range bookings {
		tanggalDisplay := "-"
		if !b.Tanggal.IsZero() {
			tanggalDisplay = b.Tanggal.Format("2006-01-02")
		}

		results = append(results, map[string]interface{}{
			"id":              b.ID,
			"nama":            b.Nama,
			"nomor_kendaraan": b.NomorKendaraan,
			"nomor_hp":        b.NomorHP,
			"tanggal":         tanggalDisplay,
			"kampung":         b.Kampung,
			"catatan":         b.Catatan,
			"status":          b.Status,
			"tim":             b.Tim,
			"catatan_admin":   b.CatatanAdmin,
			"harga":           b.Harga,
			"jenis_layanan":   "Pajak Kendaraan Bermotor",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   results,
	})
}

// ======================================================
// UPDATE BOOKING MUTASI + KIRIM NOTIFIKASI KE USER TERKAIT
// ======================================================
func UpdateMutasi(c *gin.Context) {
	id := c.Param("id")

	var booking models.MutasiKendaraan
	if err := config.DB.First(&booking, id).Error; err != nil {
		utils.Error(c, "Data booking tidak ditemukan")
		return
	}

	var input struct {
		Tanggal      string `json:"tanggal"`
		Tim          string `json:"tim"`
		CatatanAdmin string `json:"catatan_admin"`
		Harga        string `json:"harga"`
		Status       string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, "Input tidak valid: "+err.Error())
		return
	}

	if input.Tanggal != "" {
		parsedDate, err := time.Parse("2006-01-02", input.Tanggal)
		if err != nil {
			utils.Error(c, "Format tanggal harus YYYY-MM-DD")
			return
		}
		booking.Tanggal = parsedDate
	} else if booking.Tanggal.IsZero() {
		booking.Tanggal = time.Now()
	}

	if input.Tim != "" {
		booking.Tim = input.Tim
	}
	if input.CatatanAdmin != "" {
		booking.CatatanAdmin = input.CatatanAdmin
	}
	if input.Harga != "" {
		booking.Harga = input.Harga
	}
	if input.Status != "" {
		booking.Status = input.Status
	}

	if err := config.DB.Save(&booking).Error; err != nil {
		utils.Error(c, "Gagal menyimpan perubahan: "+err.Error())
		return
	}

	notification := models.Notification{
		UserID:  fmt.Sprintf("%d", booking.UserID),
		Title:   "Status Diperbarui",
		Message: fmt.Sprintf("Pesanan atas nama %s kini berstatus: %s", booking.Nama, booking.Status),
		Type:    "",
		IsRead:  false,
		Tanggal: &booking.Tanggal,
	}

	if err := config.DB.Create(&notification).Error; err != nil {
		fmt.Println("Error creating notification:", err)
	}

	sendNotificationToPhone(booking.NomorHP, "Status Diperbarui", fmt.Sprintf("Pesanan atas nama %s kini berstatus: %s", booking.Nama, booking.Status))

	utils.Success(c, gin.H{
		"message": "Booking berhasil diperbarui dan notifikasi dikirim",
		"data":    booking,
	})
}

// ======================================================
// GET SEMUA DATA MUTASI
// ======================================================
func GetAllMutasi(c *gin.Context) {
	var bookings []models.MutasiKendaraan

	// Ambil semua data dari tabel Mutasi Kendaraan
	if err := config.DB.Find(&bookings).Error; err != nil {
		utils.Error(c, "Gagal mengambil data dari database: "+err.Error())
		return
	}

	// Format data untuk dikirim ke frontend
	results := []map[string]interface{}{}
	for _, b := range bookings {
		tanggalDisplay := "-"
		if !b.Tanggal.IsZero() {
			tanggalDisplay = b.Tanggal.Format("2006-01-02")
		}

		results = append(results, map[string]interface{}{
			"id":              b.ID,
			"nama":            b.Nama,
			"nomor_kendaraan": b.NomorKendaraan,
			"nomor_hp":        b.NomorHP,
			"tanggal":         tanggalDisplay,
			"kampung":         b.Kampung,
			"catatan":         b.Catatan,
			"status":          b.Status,
			"tim":             b.Tim,
			"catatan_admin":   b.CatatanAdmin,
			"harga":           b.Harga,
			"jenis_layanan":   "Mutasi Kendaraan",
		})
	}

	// Kirim response JSON
	c.JSON(200, gin.H{
		"status": "success",
		"data":   results,
	})
}

// ======================================================
// CEK UPDATE UNTUK NOTIFIKASI (MASIH ADA - POLLING OPTIONAL)
// ======================================================
func CheckUpdate(c *gin.Context) {
	var pajak models.PajakKendaraanBermotor
	var balik models.BalikNamaKendaraan
	var mutasi models.MutasiKendaraan

	// Ambil data terbaru dari masing-masing layanan
	config.DB.Order("updated_at desc").First(&pajak)
	config.DB.Order("updated_at desc").First(&balik)
	config.DB.Order("updated_at desc").First(&mutasi)

	latestTime := time.Time{}
	latestType := ""
	var latestData interface{}

	// Tentukan data terbaru berdasarkan waktu update
	if pajak.UpdatedAt.After(latestTime) {
		latestTime = pajak.UpdatedAt
		latestType = "Pajak Kendaraan Bermotor"
		latestData = pajak
	}
	if balik.UpdatedAt.After(latestTime) {
		latestTime = balik.UpdatedAt
		latestType = "Balik Nama Kendaraan"
		latestData = balik
	}
	if mutasi.UpdatedAt.After(latestTime) {
		latestTime = mutasi.UpdatedAt
		latestType = "Mutasi Kendaraan"
		latestData = mutasi
	}

	// Jika tidak ada perubahan terbaru
	if latestType == "" {
		c.JSON(http.StatusOK, gin.H{
			"update":  false,
			"message": "Belum ada data yang diperbarui.",
		})
		return
	}

	// Jika ada update terbaru, kirimkan pesan
	c.JSON(http.StatusOK, gin.H{
		"update":        true,
		"message":       "Ada update Status nih, apakah ini punya kamu?",
		"jenis_layanan": latestType,
		"data":          latestData,
	})
}

