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

func AddInquiry(c *gin.Context) {
	var body struct {
		CarBrand string `json:"cars_brand" binding:"required"`
		CarYear  string `json:"cars_year" binding:"required"`
		Problem  string `json:"problem" binding:"required"`
		Phone    string `json:"phone" binding:"required"`
	}

	// Bind data JSON to struct
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := initializers.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
	}()

	newInquiry := models.Inquiry{
		CarBrand: body.CarBrand,
		CarYear:  body.CarYear,
		Problem:  body.Problem,
		Phone:    body.Phone,
	}

	// Save new item to database
	if err := tx.Create(&newInquiry).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit Inquiry", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Inquiry successfully Submitted", "data": newInquiry})
}

func GetAllInquiries(c *gin.Context) {

	var inquiry []models.Inquiry
	if err := initializers.DB.Order("created_at DESC").Find(&inquiry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get inquiries", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get inquiries successfully",
		"data":    inquiry,
	})
}

func UpdateStatusInquiry(c *gin.Context) {
	inquiryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Inquiry ID"})
		return
	}

	tx := initializers.DB.Begin()

	var existingInquiry models.Inquiry
	// Cari Inquiry berdasarkan ID
	if err = tx.First(&existingInquiry, inquiryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Inquiry not found"})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Inquiry", "details": err.Error()})
		return
	}

	// Toggle the status
	if existingInquiry.Status != nil && *existingInquiry.Status {
		existingInquiry.Status = new(bool)
		*existingInquiry.Status = false
	} else {
		existingInquiry.Status = new(bool)
		*existingInquiry.Status = true
	}

	// Simpan perubahan ke database
	if err := tx.Save(&existingInquiry).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Inquiry", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Status Changed", "data": existingInquiry})
}

func DeleteInquiry(c *gin.Context) {
	inquiryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	var inquiry models.Inquiry
	err = initializers.DB.First(&inquiry, inquiryID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "inquiry not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve inquiry"})
		return
	}

	err = initializers.DB.Unscoped().Delete(&inquiry).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete inquiry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inquiry successfully deleted"})
}

func EditInquiry(c *gin.Context) {
	var body struct {
		CarBrand string `json:"cars_brand" binding:"required"`
		CarYear  string `json:"cars_year" binding:"required"`
		Problem  string `json:"problem" binding:"required"`
		Phone    string `json:"phone" binding:"required"`
	}

	// Bind data form ke struct EditArticleRequest
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inventory_id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	// Init transaction
	tx := initializers.DB.Begin()

	var existingInventory models.Inquiry
	// Cari artikel berdasarkan ID
	if err = tx.First(&existingInventory, inventory_id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Inventory", "details": err.Error()})
		return
	}

	if body.CarBrand != "" {
		existingInventory.CarBrand = body.CarBrand
	}
	if body.CarYear != "" {
		existingInventory.CarYear = body.CarYear
	}
	if body.Problem != "" {
		existingInventory.Problem = body.Problem
	}
	if body.Phone != "" {
		existingInventory.Phone = body.Phone
	}

	// Simpan perubahan ke database
	if err := tx.Save(&existingInventory).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit inquiry", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Inquiry Successfully Edited", "data": existingInventory})
}
