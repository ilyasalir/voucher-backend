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
	"gorm.io/gorm"
)

func AddCar(c *gin.Context) {
	var body struct {
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
	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)
	newCar := models.Car{
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

	// Preload data CarType, Color, dan Brand saat mengembalikan respons
	if err := tx.Preload("CarType.Brand").Preload("Color").First(&newCar).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preload related data", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Car added successfully", "data": newCar})
}

func GetCarsByUser(c *gin.Context) {
	// Ambil ID pengguna dari konteks (biasanya diambil dari token otentikasi)
	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)

	// Query mobil berdasarkan ID pengguna
	var cars []models.Car
	if err := initializers.DB.Where("user_id = ?", user.ID).Preload("CarType.Brand").Preload("Color").Find(&cars).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cars", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get cars successfully", "data": cars})
}

func GetCarDetails(c *gin.Context) {
	// Ambil ID pengguna dari konteks (biasanya diambil dari token otentikasi)
	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)

	// Ambil ID mobil dari parameter URL
	licensePlat := c.Param("plat")

	// Query detail mobil berdasarkan ID pengguna dan ID mobil
	var car models.Car
	if err := initializers.DB.Where("user_id = ? AND license_plat = ?", user.ID, licensePlat).Preload("CarType.Brand").Preload("Color").First(&car).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get car details", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get car details successfully", "data": car})
}

func EditCar(c *gin.Context) {
	var body struct {
		LicensePlat  string `form:"license_plat"`
		CarTypeName  string `form:"car_type_name"`
		ColorName    string `form:"color_name"`
		FrameNumber  string `form:"frame_number"`
		EngineNumber string `form:"engine_number"`
		Kilometer    uint64 `form:"kilometer"`
		BrandID      uint   `form:"brand_id"`
	}

	// Bind data form ke struct EditCarRequest
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	carID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	// Init transaction
	tx := initializers.DB.Begin()

	var existingCar models.Car
	// Cari mobil berdasarkan ID
	if err = tx.First(&existingCar, carID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve car", "details": err.Error()})
		return
	}

	// Cek apakah license_plat sudah ada di database sebelumnya
	if body.LicensePlat != "" && body.LicensePlat != existingCar.LicensePlat {
		var checkCar models.Car
		if result := tx.Where("license_plat = ?", body.LicensePlat).First(&checkCar); result.Error == nil && checkCar.ID != uint(carID) {
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
	}

	// Cek apakah CarType sudah ada atau gunakan yang sudah ada
	var carType models.CarType
	if body.CarTypeName != "" {
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get car types", "details": err.Error()})
				return
			}
		}
	} else {
		// Gunakan CarType yang sudah ada di existingCar
		carType = existingCar.CarType
	}

	// Cek apakah Color sudah ada atau gunakan yang sudah ada
	var color models.Color
	if body.ColorName != "" {
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get colors", "details": err.Error()})
				return
			}
		}
	} else {
		// Gunakan Color yang sudah ada di existingCar
		color = existingCar.Color
	}

	// Mendapatkan file dari form
	file, err := c.FormFile("photo")
	if err == nil && file != nil { // jika foto ada
		// Jika ada file baru diunggah
		photoUrl, err := utils.UploadFileToFirebase(*file)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Hapus foto yang lama hanya jika foto baru diunggah
		if existingCar.PhotoUrl != "" {
			err = utils.DeleteFileFromFirebase(existingCar.PhotoUrl)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		existingCar.PhotoUrl = photoUrl
	}

	// Update data mobil dengan nilai yang baru
	if body.LicensePlat != "" {
		existingCar.LicensePlat = body.LicensePlat
	}
	if body.FrameNumber != "" {
		existingCar.FrameNumber = body.FrameNumber
	}
	if body.EngineNumber != "" {
		existingCar.EngineNumber = body.EngineNumber
	}
	if body.Kilometer != 0 {
		existingCar.Kilometer = body.Kilometer
	}

	// Assign atau tambahkan CarType baru jika diperlukan
	if carType.ID == 0 {
		existingCar.CarType = carType
	} else {
		existingCar.CarTypeID = carType.ID
	}

	// Assign atau tambahkan Color baru jika diperlukan
	if color.ID == 0 {
		existingCar.Color = color
	} else {
		existingCar.ColorID = color.ID
	}

	// Simpan perubahan ke database
	if err := tx.Save(&existingCar).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit car", "details": err.Error()})
		return
	}

	// Preload data CarType, Color, dan Brand saat mengembalikan respons
	if err := tx.Preload("CarType.Brand").Preload("Color").First(&existingCar).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preload related data", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Car edited successfully", "data": existingCar})
}

func DeleteCar(c *gin.Context) {
	// Mendapatkan ID Car dari path parameter
	carID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	// Init transaction
	tx := initializers.DB.Begin()

	// Cari Car berdasarkan ID
	var car models.Car
	if err := tx.First(&car, carID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve car"})
		return
	}

	// Hapus file/foto dari server
	// if err := utils.DeleteFile(car.PhotoUrl); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete photo", "details": err.Error()})
	// 	return
	// }

	// Jika ada foto
	if car.PhotoUrl != "" {
		err = utils.DeleteFileFromFirebase(car.PhotoUrl)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Hapus Car dari database
	if err := tx.Unscoped().Delete(&car).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete car", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Car deleted successfully"})
}
