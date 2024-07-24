package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetColors(c *gin.Context) {
	var colors []models.Color

	if err := initializers.DB.Order("name").Find(&colors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get colors", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get colors data successfully", "data": colors})
}
