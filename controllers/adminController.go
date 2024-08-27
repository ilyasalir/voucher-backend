package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"carport-backend/utils"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetUsersByUserId(c *gin.Context) {
	userId := c.Query("user_id")

	// Check if userId is provided
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Validate if userId is a valid integer
	userIdInt, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	// Query cars based on user ID
	var users []models.User
	if err := initializers.DB.Where("ID = ?", userIdInt).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cars", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get users successfully", "data": users})
}

func GetCarsByUserId(c *gin.Context) {
	userId := c.Query("user_id")

	// Check if userId is provided
	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Validate if userId is a valid integer
	userIdInt, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	// Query cars based on user ID
	var cars []models.Car
	if err := initializers.DB.Where("user_id = ?", userIdInt).Preload("CarType.Brand").Preload("Color").Preload("User").Find(&cars).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cars", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get cars successfully", "data": cars})
}

func GetAllUsers(c *gin.Context) {

	var users []models.User
	if err := initializers.DB.Preload("Addresses").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get users successfully",
		"data":    users,
	})
}

func EditUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

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
	if err := initializers.DB.Preload("Addresses").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user", "details": err.Error()})
		return
	}

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

func AddCarByAdmin(c *gin.Context) {
	var body struct {
		UserID       uint   `form:"user_id" binding:"required"`
		LicensePlat  string `form:"license_plat" binding:"required"`
		CarTypeName  string `form:"car_type_name" binding:"required"`
		ColorName    string `form:"color_name" binding:"required"`
		FrameNumber  string `form:"frame_number"`
		EngineNumber string `form:"engine_number"`
		Kilometer    uint64 `form:"kilometer"`
		BrandID      uint   `form:"brand_id" binding:"required"`
	}

	// Bind data JSON ke struct AddCarRequest
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Init transaction
	tx := initializers.DB.Begin()

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
	newCar := models.Car{
		UserID:       body.UserID,
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

	// Preload data CarType, Color, dan Brand saat mengembalikan respons
	if err := tx.Preload("CarType.Brand").Preload("Color").First(&newCar).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preload related data", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Car added successfully", "data": newCar})
}
