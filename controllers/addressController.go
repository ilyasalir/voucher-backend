package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddAddress(c *gin.Context) {
	var body struct {
		Title    string `json:"title" binding:"required"`
		Location string `json:"location" binding:"required"`
	}

	// Bind data JSON ke struct AddAddressRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Buat instance Address dari data yang di-bind user id diambil dari auth
	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)
	newAddress := models.Address{
		UserID:   user.ID,
		Title:    body.Title,
		Location: body.Location,
	}

	// Simpan alamat baru ke database
	result := initializers.DB.Create(&newAddress)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add address", "details": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Address added successfully", "data": newAddress})
}

func DeleteAddress(c *gin.Context) {
	addressID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	// Cari alamat berdasarkan ID
	var address models.Address
	err = initializers.DB.First(&address, addressID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve address"})
		return
	}

	// Hapus alamat dari database
	err = initializers.DB.Unscoped().Delete(&address).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete address"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Address deleted successfully"})
}

func EditAddress(c *gin.Context) {
	var body struct {
		Title    string `json:"title" binding:"required"`
		Location string `json:"location" binding:"required"`
	}

	// Bind data JSON ke struct AddAddressRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	addressID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
		return
	}

	// Cari alamat berdasarkan ID
	var address models.Address
	err = initializers.DB.First(&address, addressID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve address"})
		return
	}

	// Update data alamat dengan nilai yang baru
	address.Title = body.Title
	address.Location = body.Location

	// Simpan perubahan ke database
	err = initializers.DB.Save(&address).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit address", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Address edited successfully", "data": address})
}
