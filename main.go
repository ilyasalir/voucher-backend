package main

import (
	"os"
	"strings"
	"voucher-backend/controllers"
	"voucher-backend/initializers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {
	// Read the environment variable containing the allowed origins
	allowedOrigins := os.Getenv("FRONTEND")

	// Split the string into a slice of origins
	origins := strings.Split(allowedOrigins, ",")

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Konfigurasi CORS dengan withCredentials
	config := cors.DefaultConfig()
	config.AllowOrigins = origins // Ganti dengan origin aplikasi frontend Anda
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"Content-Type", "Authorization", "Cookie", "Set-Cookie"}
	config.AllowCredentials = true // Mengizinkan penggunaan withCredentials

	r.Use(cors.New(config))

	brandRoutes := r.Group("/brand")
	{
		brandRoutes.POST("", controllers.AddBrand)
	}

	voucherRoutes := r.Group("/voucher")
	{
		voucherRoutes.POST("", controllers.AddVoucher)
		voucherRoutes.GET("/:id", controllers.GetVoucherByID)
		voucherRoutes.GET("/brand/:id", controllers.GetVoucherByBrandID)
	}

	transactionRoutes := r.Group("/transaction")
	{
		transactionRoutes.POST("/redemption", controllers.RedeemVouchers)
		transactionRoutes.GET("/redemption/:id", controllers.GetTransactionByID)
	}

	r.Run()
}
