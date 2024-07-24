package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"carport-backend/utils"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddStnk(c *gin.Context) {
	tx := initializers.DB.Begin()

	file, err := c.FormFile("photo")
	var photoUrl string

	// Check if file exists
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload error"})
		return
	}

	if file != nil {
		// Process upload
		photoUrl, err = utils.UploadFileToFirebase(*file)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload photo failed"})
			return
		}
	}

	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)

	// Log user ID for debugging
	log.Println("User ID for STNK creation:", user.ID)

	if user.ID == 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	newStnk := models.Stnk{
		UserID:   user.ID,
		PhotoUrl: photoUrl,
	}

	if err := tx.Create(&newStnk).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create STNK", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "STNK added successfully", "data": newStnk})
}

func GetStnkByUser(c *gin.Context) {
	// Ambil ID pengguna dari konteks (biasanya diambil dari token otentikasi)
	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)

	// Query mobil berdasarkan ID pengguna
	var stnk []models.Stnk
	err := initializers.DB.Preload("User").Where("user_id = ?", user.ID).Find(&stnk).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get STNK", "details": err.Error()})
		return
	}
	if len(stnk) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No STNK found for the user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Get STNK successfully", "data": stnk})
}

func GetAllStnk(c *gin.Context) {

	var stnk []models.Stnk
	if err := initializers.DB.Preload("User").Find(&stnk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stnks", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get stnks successfully",
		"data":    stnk,
	})
}

func AccStnk(c *gin.Context) {

	stnkID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	tx := initializers.DB.Begin()

	var existingStnk models.Stnk
	// Cari stnk berdasarkan ID
	if err = tx.First(&existingStnk, stnkID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve car", "details": err.Error()})
		return
	}

	status := true
	existingStnk.Status = &status

	// Simpan perubahan ke database
	if err := tx.Save(&existingStnk).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit car", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Verified successfully", "data": existingStnk})
}

func AddDesc(c *gin.Context) {
	var body struct {
		Description string `json:"description"`
	}

	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stnkID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	tx := initializers.DB.Begin()

	var existingDesc models.Stnk
	// Cari stnk berdasarkan ID
	if err = tx.First(&existingDesc, stnkID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve car", "details": err.Error()})
		return
	}
	existingDesc.Description = body.Description

	// Simpan perubahan ke database
	if err := tx.Save(&existingDesc).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit car", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Verified successfully", "data": existingDesc})
}

func UpdateStnk(c *gin.Context) {
	stnkID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stnk ID"})
		return
	}

	tx := initializers.DB.Begin()

	var existingStnk models.Stnk
	// Cari stnk berdasarkan ID
	if err = tx.First(&existingStnk, stnkID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Stnk not found"})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Stnk", "details": err.Error()})
		return
	}

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
		if existingStnk.PhotoUrl != "" {
			err = utils.DeleteFileFromFirebase(existingStnk.PhotoUrl)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		existingStnk.PhotoUrl = photoUrl
	}

	// Simpan perubahan ke database
	if err := tx.Save(&existingStnk).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Upload Stnk", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Upload successfully", "data": existingStnk})
}

func DeleteDesc(c *gin.Context) {
	stnkID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stnk ID"})
		return
	}

	tx := initializers.DB.Begin()

	var existingStnk models.Stnk
	// Cari stnk berdasarkan ID
	if err = tx.First(&existingStnk, stnkID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Stnk not found"})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Stnk", "details": err.Error()})
		return
	}

	// Set the description to an empty string
	existingStnk.Description = ""

	// Simpan perubahan ke database
	if err := tx.Save(&existingStnk).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete description", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Description deleted successfully", "data": existingStnk})
}
