package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetCarTypes(c *gin.Context) {
	brandId := c.Query("brand_id")

	tx := initializers.DB.Begin()
	var typeCars []models.CarType

	// Filter berdasarkan brand_id
	if brandId != "" {
		tx = tx.Where("brand_id = ?", brandId)
	}

	// Eksekusi query dan urutkan hasil berdasarkan name
	if err := tx.Order("name").Find(&typeCars).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get type car", "details": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Get type car data successfully", "data": typeCars})
}
