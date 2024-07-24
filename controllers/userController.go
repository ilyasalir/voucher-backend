package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"carport-backend/utils"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(c *gin.Context) {
	// Get the email/password of req body
	var body struct {
		Name            string `json:"name" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		Phone           string `json:"phone" binding:"required,min=10,max=15"`
		Password        string `json:"password" binding:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
		Role            string `json:"role" binding:"omitempty,oneof=USER ADMIN"`
		Title           string `json:"title" binding:"omitempty"`
		Location        string `json:"location" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Manual validation for ConfirmPassword
	if body.Password != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password and Confirm Password do not match"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash the password"})
		return
	}

	// Create the user
	user := models.User{
		Name:     body.Name,
		Email:    body.Email,
		Phone:    body.Phone,
		Password: string(hashedPassword),
		Role:     body.Role,
	}

	tx := initializers.DB.Begin()

	// Cek apakah email sudah ada di database sebelumnya
	var existingUser models.User
	if result := tx.Where("email = ?", user.Email).First(&existingUser); result.Error == nil {
		// User dengan email tersebut sudah ada, kembalikan error
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already in use"})
		return
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Terjadi error lain selain ErrRecordNotFound, kembalikan error
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Store to DB
	if result := tx.Create(&user); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Check if address not empty
	if body.Location != "" && body.Title != "" {
		address := models.Address{
			UserID:   user.ID,
			Title:    body.Title,
			Location: body.Location,
		}

		// Check if the address creation is successful
		if err := tx.Create(&address).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address", "details": err.Error()})
			return
		}
	}

	// commit transaction
	tx.Commit()

	// Respond
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "data": user})
}

// IsEmailExist checks whether the email exists in the database.
func IsEmailExist(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	result := initializers.DB.Where("email = ?", body.Email).First(&existingUser)
	if result.Error == nil {
		// User with the email exists
		c.JSON(http.StatusOK, gin.H{"exists": true})
		return
	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Email does not exist in the database
		c.JSON(http.StatusOK, gin.H{"exists": false})
		return
	} else {
		// Some other error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
}

func EditProfile(c *gin.Context) {
	// Get the email/password of req body
	var body struct {
		Name            string `json:"name" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		Phone           string `json:"phone" binding:"required,min=10,max=15"`
		Password        string `json:"password" binding:"omitempty,min=8"`
		ConfirmPassword string `json:"confirm_password" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)

	tx := initializers.DB.Begin()

	if user.Email != body.Email {
		// Cek apakah email sudah ada di database sebelumnya
		var existingUser models.User
		if result := tx.Where("email = ?", body.Email).First(&existingUser); result.Error == nil && existingUser.ID != user.ID {
			// User dengan email tersebut sudah ada, kembalikan error
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already in use"})
			return
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Terjadi error lain selain ErrRecordNotFound, kembalikan error
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		user.Email = body.Email
	}

	if body.Password != "" {
		// Manual validation for ConfirmPassword
		if body.Password != body.ConfirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password and Confirm Password do not match"})
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash the password"})
			return
		}

		user.Password = string(hashedPassword)
	}

	user.Name = body.Name
	user.Phone = body.Phone

	// Simpan perubahan ke database
	if err := initializers.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit user", "details": err.Error()})
		return
	}

	// commit transaction
	tx.Commit()

	// Respond
	c.JSON(http.StatusOK, gin.H{"message": "User edited successfully", "data": user})
}

func Login(c *gin.Context) {
	// Get the email/password of req body
	var body struct {
		Email    string `binding:"required"`
		Password string `binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up requested user
	var user models.User
	err := initializers.DB.First(&user, "email = ?", body.Email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Compare sent in pass with saved user pass hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Check your password again"})
		return
	}

	// Generate a jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed create token"})
		return
	}

	// Send it back
	// c.SetSameSite(http.SameSiteNoneMode)
	// c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", true, true)
	// c.SetCookie("Authorization", tokenString, 3600*24*30, "/", "carport-backend.onrender.com", true, true)

	c.JSON(http.StatusOK, gin.H{"message": "User logged successfully", "token": tokenString})
}

func GetCurrentUser(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"message": "Get current user successfully",
		"data":    user,
	})
}

// func Logout(c *gin.Context) {
// 	// c.SetSameSite(http.SameSiteNoneMode)
// 	// c.SetCookie("Authorization", "", 0, "", "", true, true)
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "User logged out successfully",
// 	})
// }

func RegisterWithCarOrOrder(c *gin.Context) {
	// Get the email/password of req body
	var body struct {
		Name            string `form:"name" binding:"required"`
		Email           string `form:"email" binding:"required,email"`
		Phone           string `form:"phone" binding:"required,min=10,max=15"`
		Password        string `form:"password" binding:"required,min=8"`
		ConfirmPassword string `form:"confirm_password" binding:"required"`
		Role            string `form:"role" binding:"omitempty,oneof=USER ADMIN"`
		Title           string `form:"title" binding:"omitempty"`
		Location        string `form:"location" binding:"omitempty"`

		LicensePlat  string `form:"license_plat" binding:"omitempty"`
		CarTypeName  string `form:"car_type_name" binding:"omitempty"`
		ColorName    string `form:"color_name" binding:"omitempty"`
		FrameNumber  string `form:"frame_number" binding:"omitempty"`
		EngineNumber string `form:"engine_number" binding:"omitempty"`
		Kilometer    uint64 `form:"kilometer" binding:"omitempty"`
		BrandID      uint   `form:"brand_id" binding:"omitempty"`

		ServiceType string    `form:"service_type" binding:"omitempty"`
		OrderTime   time.Time `form:"order_time" binding:"omitempty"`
		Services    string    `form:"services" binding:"omitempty"`
	}

	// Bind data JSON ke struct AddCarRequest
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Manual validation for ConfirmPassword
	if body.Password != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password and Confirm Password do not match"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash the password"})
		return
	}

	// Create the user
	user := models.User{
		Name:     body.Name,
		Email:    body.Email,
		Phone:    body.Phone,
		Password: string(hashedPassword),
		Role:     body.Role,
	}

	tx := initializers.DB.Begin()

	// Cek apakah email sudah ada di database sebelumnya
	var existingUser models.User
	if result := tx.Where("email = ?", user.Email).First(&existingUser); result.Error == nil {
		// User dengan email tersebut sudah ada, kembalikan error
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already in use"})
		return
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Terjadi error lain selain ErrRecordNotFound, kembalikan error
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Store to DB
	if result := tx.Create(&user); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Check if address not empty
	if body.Location != "" && body.Title != "" {
		address := models.Address{
			UserID:   user.ID,
			Title:    body.Title,
			Location: body.Location,
		}

		// Check if the address creation is successful
		if err := tx.Create(&address).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address", "details": err.Error()})
			return
		}
	}

	// Add car
	var newCar models.Car
	if body.BrandID != 0 && body.CarTypeName != "" && body.ColorName != "" && body.LicensePlat != "" {
		// Cek apakah license_plat sudah ada di database sebelumnya
		var existingCar models.Car
		if result := tx.Where("license_plat = ?", body.LicensePlat).First(&existingCar); result.Error == nil {
			// Car dengan license_plat tersebut sudah ada, kembalikan error
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "License plat is already in use"})
			return
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Terjadi error lain selain ErrRecordNotFound, kembalikan error
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		// Cek apakah CarType sudah ada
		var carType models.CarType
		if err := tx.Where("LOWER(name) = ? AND brand_id = ?", strings.ToLower(body.CarTypeName), body.BrandID).First(&carType).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// CarType belum ada, tambahkan
				carType = models.CarType{
					Name:    strings.Title(body.CarTypeName), // Huruf kapital pada setiap awal kata
					BrandID: body.BrandID,
				}
				if err := tx.Create(&carType).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add car type", "details": err.Error()})
					return
				}
			} else {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrive car types", "details": err.Error()})
				return
			}
		}

		// Cek apakah Color sudah ada
		var color models.Color
		if err := tx.Where("LOWER(name) = ?", strings.ToLower(body.ColorName)).First(&color).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Color belum ada, tambahkan
				color = models.Color{
					Name: strings.Title(body.ColorName), // Huruf kapital pada setiap awal kata
				}
				if err := tx.Create(&color).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add color", "details": err.Error()})
					return
				}

			} else {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrive colors", "details": err.Error()})
				return
			}
		}

		// Mendapatkan file dari form
		file, err := c.FormFile("photo")
		var photoUrl string

		// Menyimpan file dan mendapatkan URL-nya
		// photoUrl, err := utils.HandleFileUpload(c, file)
		// if err != nil {
		// 	tx.Rollback()
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		// 	return
		// }

		// Jika file foto ada, maka lakukan proses upload
		if err == nil && file != nil {
			// Proses upload file ke Firebase
			photoUrl, err = utils.UploadFileToFirebase(*file)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload photo failed"})
				return
			}
		}

		// Buat instance Car dari data yang di-bind user id diambil dari auth
		newCar = models.Car{
			UserID:       user.ID,
			LicensePlat:  body.LicensePlat,
			CarTypeID:    carType.ID,
			ColorID:      color.ID,
			FrameNumber:  body.FrameNumber,
			EngineNumber: body.EngineNumber,
			Kilometer:    body.Kilometer,
			PhotoUrl:     photoUrl,
		}

		// Simpan car baru ke database
		if err := tx.Create(&newCar).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add car", "details": err.Error()})
			return
		}
	}

	// Add Order
	if body.ServiceType != "" && body.Services != "" {
		// Validasi jam operasional bengkel (8.00 hingga 18.00)
		location, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load location", "details": err.Error()})
			return
		}

		operationalStartTime := time.Date(body.OrderTime.Year(), body.OrderTime.Month(), body.OrderTime.Day(), 8, 0, 0, 0, location)
		operationalEndTime := time.Date(body.OrderTime.Year(), body.OrderTime.Month(), body.OrderTime.Day(), 18, 0, 0, 0, location)

		if body.OrderTime.Before(operationalStartTime) || body.OrderTime.Add(time.Duration(1)*time.Hour).After(operationalEndTime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order time or duration, outside of operational hours"})
			return
		}

		interval := 1 // Interval dalam jam
		newOrderStartTime := body.OrderTime
		newOrderEndTime := body.OrderTime.Add(time.Duration(interval) * time.Hour)

		var overlappingOrdersCount int64
		var existingOrders []models.Order
		if err := tx.Model(&models.Order{}).Find(&existingOrders).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve existing orders", "details": err.Error()})
			return
		}

		for _, existingOrder := range existingOrders {
			existingOrderStartTime := existingOrder.OrderTime
			existingOrderEndTime := existingOrder.OrderTime.Add(time.Duration(existingOrder.Duration) * time.Hour)

			if ((newOrderStartTime.Before(existingOrderEndTime) && existingOrderStartTime.Before(newOrderEndTime)) ||
				(newOrderEndTime.After(existingOrderStartTime) && existingOrderStartTime.After(newOrderStartTime))) && existingOrder.Status != "CANCELED" {
				overlappingOrdersCount++
			}
		}

		if overlappingOrdersCount >= 2 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Exceeded maximum orders for the given time"})
			return
		}

		// Membuat order baru
		newOrder := models.Order{
			UserID:      user.ID,
			CarID:       newCar.ID,
			ServiceType: body.ServiceType,
			Address:     body.Location,
			OrderTime:   body.OrderTime,
		}

		// Menambahkan layanan (services) ke order
		services := strings.Split(body.Services, ",")
		var orderServices []models.Service
		for _, serviceName := range services {
			// Gunakan title case saat memeriksa keberadaan layanan
			titleCaseName := strings.Title(serviceName)
			service := models.Service{Name: titleCaseName}

			if result := tx.FirstOrCreate(&service, models.Service{Name: titleCaseName}); result.Error != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add service", "details": result.Error.Error()})
				return
			}
			orderServices = append(orderServices, service)
		}

		newOrder.Services = orderServices

		// Simpan order baru ke database
		if result := tx.Create(&newOrder); result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add order", "details": result.Error.Error()})
			return
		}
	}

	// commit transaction
	tx.Commit()

	// Generate a jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed create token"})
		return
	}

	// Respond
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "token": tokenString, "data": user})
}

func DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	var user models.User
	err = initializers.DB.First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve User"})
		return
	}

	err = initializers.DB.Unscoped().Delete(&user).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete User"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User and associated orders and cars deleted successfully"})
}
