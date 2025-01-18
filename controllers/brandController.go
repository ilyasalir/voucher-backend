package controllers

import (
	"net/http"
	"voucher-backend/initializers"
	"voucher-backend/models"

	"github.com/gin-gonic/gin"
)

func AddBrand(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	brand := models.Brand{
		Name: body.Name,
	}

	tx := initializers.DB.Begin()

	if result := tx.Create(&brand); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Brand added successfully", "data": brand})
}
