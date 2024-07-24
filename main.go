package main

import (
	"carport-backend/controllers"
	"carport-backend/initializers"
	"carport-backend/middleware"
	"os"
	"path/filepath"
	"strings"

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

	r.GET("/uploads/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		filePath := filepath.Join("uploads", filename)
		c.File(filePath)
	})

	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", controllers.RegisterWithCarOrOrder)
		authRoutes.POST("/login", controllers.Login)
		authRoutes.GET("/checkEmail", controllers.IsEmailExist)
		// authRoutes.POST("/logout", middleware.RequireAuth, controllers.Logout)
		authRoutes.PUT("/edit", middleware.RequireAuth, controllers.EditProfile)
		authRoutes.GET("", middleware.RequireAuth, controllers.GetCurrentUser)
	}

	addressRoutes := r.Group("/address")
	{
		addressRoutes.POST("", middleware.RequireAuth, controllers.AddAddress)
		addressRoutes.DELETE("/:id", middleware.RequireAuth, controllers.DeleteAddress)
		addressRoutes.PUT("/:id", middleware.RequireAuth, controllers.EditAddress)
	}

	stnkRoutes := r.Group("/stnk")
	{
		stnkRoutes.POST("", middleware.RequireAuth, controllers.AddStnk)
		stnkRoutes.GET("", middleware.RequireAuth, controllers.GetStnkByUser)
		stnkRoutes.PUT("/update/:id", middleware.RequireAuth, controllers.UpdateStnk)
		stnkRoutes.PUT("/desc/delete/:id", middleware.RequireAuth, controllers.DeleteDesc)
	}

	carRoutes := r.Group("/car")
	{
		carRoutes.POST("", middleware.RequireAuth, controllers.AddCar)
		carRoutes.GET("", middleware.RequireAuth, controllers.GetCarsByUser)
		carRoutes.GET("/:plat", middleware.RequireAuth, controllers.GetCarDetails)
		carRoutes.PUT("/:id", middleware.RequireAuth, controllers.EditCar)
		carRoutes.DELETE("/:id", middleware.RequireAuth, controllers.DeleteCar)
	}

	orderRoutes := r.Group("/order")
	{
		orderRoutes.GET("", middleware.RequireAuth, controllers.GetOrder)
		orderRoutes.GET("/used", controllers.GetOrderUsed)
		orderRoutes.POST("", middleware.RequireAuth, controllers.AddOrder)
		orderRoutes.DELETE("/:id", middleware.RequireAuth, controllers.DeleteOrder)
		orderRoutes.PUT("/status/:id", middleware.RequireAuth, controllers.UpdateStatusOrder)
		orderRoutes.PUT("/price/:id", middleware.RequireAuth, controllers.UpdatePriceOrder)
	}

	colorRoutes := r.Group("/color")
	{
		colorRoutes.GET("", controllers.GetColors)
	}

	typeCarRoutes := r.Group("/cartype")
	{
		typeCarRoutes.GET("", controllers.GetCarTypes)
	}

	serviceRoutes := r.Group("/service")
	{
		serviceRoutes.GET("", middleware.RequireAuth, controllers.GetServices)
	}

	adminRoutes := r.Group("/admin")
	{
		adminRoutes.GET("/cars", middleware.RequireAdmin, controllers.GetCarsByUserId)
		adminRoutes.GET("/users", middleware.RequireAdmin, controllers.GetAllUsers)
		adminRoutes.GET("/usersbyid", middleware.RequireAdmin, controllers.GetUsersByUserId)
		adminRoutes.GET("/stnk", middleware.RequireAdmin, controllers.GetAllStnk)
		adminRoutes.PUT("/user/:id", middleware.RequireAdmin, controllers.EditUser)
		adminRoutes.POST("/car", middleware.RequireAdmin, controllers.AddCarByAdmin)
		adminRoutes.POST("/register", middleware.RequireAdmin, controllers.Register)
		adminRoutes.DELETE("/:id", middleware.RequireAuth, controllers.DeleteUser)
		adminRoutes.PUT("/stnk/acc/:id", middleware.RequireAuth, controllers.AccStnk)
		adminRoutes.PUT("/stnk/desc/:id", middleware.RequireAuth, controllers.AddDesc)
	}

	r.Run()
}
