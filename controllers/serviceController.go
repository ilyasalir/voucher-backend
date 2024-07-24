package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetServices(c *gin.Context) {
	// Buat instance Car dari data yang di-bind user id diambil dari auth
	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)

	var services []models.Service
	if err := initializers.DB.
		Model(&services).
		Joins("JOIN order_services ON order_services.service_id = services.id").
		Joins("JOIN orders ON order_services.order_id = orders.id").
		Where("orders.user_id = ?", user.ID).
		Select("DISTINCT services.id, services.name").
		Order("services.name").
		Find(&services).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve services", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get services data successfully", "data": services})
}
