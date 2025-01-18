package controllers

import (
	"errors"
	"net/http"
	"voucher-backend/initializers"
	"voucher-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddVoucher(c *gin.Context) {
	var body struct {
		Name     string `json:"name" binding:"required"`
		Discount uint   `json:"discount" binding:"required"`
		Point    uint   `json:"point" binding:"omitempty"`
		Quantity uint   `json:"quantity" binding:"required"`
		BrandID  uint   `json:"brand_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	voucher := models.Voucher{
		Name:     body.Name,
		Discount: body.Discount,
		Point:    body.Point,
		Quantity: body.Quantity,
		BrandID:  body.BrandID,
	}

	tx := initializers.DB.Begin()

	if result := tx.Create(&voucher); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Voucher added successfully", "data": voucher})
}

func GetVoucherByID(c *gin.Context) {
	ID := c.Param("id")

	tx := initializers.DB.Begin()

	var voucher models.Voucher
	if err := tx.Where("id = ?", ID).Preload("Brand").Find(&voucher).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Voucher not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Voucher details", "details": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Get voucher details successfully", "data": voucher})
}

func GetVoucherByBrandID(c *gin.Context) {
	ID := c.Param("id")

	tx := initializers.DB.Begin()

	var voucher []models.Voucher
	if err := tx.Where("brand_id = ?", ID).Preload("Brand").Find(&voucher).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Voucher not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Voucher details", "details": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Get voucher details successfully", "data": voucher})
}
