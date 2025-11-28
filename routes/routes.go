package routes

import (
	"samsat_backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	// ======================================================
	// ROUTES TANPA LOGIN
	// ======================================================
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	// ======================================================
	// ROUTES LOGIN MENGGUNAKAN BEARER TOKEN DARI DATABASE
	// ======================================================
	// Tidak menggunakan middleware JWTAuthMiddleware lagi
	// Semua pengecekan token ada di masing-masing controller

	// Booking
	r.POST("/booking", controllers.CreateBooking)
	r.GET("/booking/search", controllers.SearchBooking)
	r.GET("/booking/all", controllers.GetAllBookings)
	r.PUT("/booking/update/:id", controllers.UpdateBooking)

	// Balik Nama
	r.PUT("/baliknama/update/:id", controllers.UpdateBalikNama)
	r.GET("/baliknama/all", controllers.GetAllBalikNama)

	// Mutasi Kendaraan
	r.PUT("/mutasikendaraan/update/:id", controllers.UpdateMutasi)
	r.GET("/mutasikendaraan/all", controllers.GetAllMutasi)

	// Logout & Profil
	r.GET("/me", controllers.Me) // halaman profil (nama & nomor HP)
	r.POST("/logout", controllers.Logout)
	r.PUT("/me", controllers.UpdateProfile)

	// Notifikasi
	r.GET("/api/notifications/:user_id", controllers.GetNotifications)
	r.POST("/api/notifications/mark-read", controllers.MarkNotificationRead)
	r.GET("/api/notifications/recent/:user_id", controllers.GetUserNotificationsRecent)
	r.DELETE("/api/notifications/:id", controllers.DeleteNotification)
	r.POST("/save-fcm", controllers.SaveFCMToken)

	// ======================================================
	// ROUTES ADMIN
	// (tidak pakai JWT user biasa / bebas login method)
	// ======================================================
	r.POST("/admin/create", controllers.CreateAdmin)
	r.GET("/admin/all", controllers.GetAllAdmin)
	r.POST("/admin/login", controllers.AdminLogin)

	// Dashboard Admin
	r.GET("/dashboard/stats", controllers.DashboardStats)

	//history
	r.GET("/api/requests/history/:user_id", controllers.GetRequestHistory)

}
