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

func AddEmailAdmin(c *gin.Context) {
	var body struct {
		Email string `form:"email" binding:"required"`
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

	newEmailAdmin := models.EmailAdmin{
		Email: body.Email,
	}

	var existingUser models.EmailAdmin
	if err := tx.Where("email = ?", newEmailAdmin.Email).First(&existingUser).Error; err == nil {
		// Email already exists, return error
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already in use"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Other errors, return internal server error
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email existence", "details": err.Error()})
		return
	}

	// Save new email admin to database
	if err := tx.Create(&newEmailAdmin).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add Email Admin", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Email Admin added successfully", "data": newEmailAdmin})
}

func GetEmailAdmin(c *gin.Context) {
	var email []models.EmailAdmin

	if err := initializers.DB.Order("email").Find(&email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get email data successfully", "data": email})
}

func DeleteEmailAdmin(c *gin.Context) {
	emailAdminId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	// Start a transaction
	tx := initializers.DB.Begin()

	var email models.EmailAdmin
	err = tx.First(&email, emailAdminId).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "email not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve email"})
		return
	}
	// Delete the email
	err = tx.Delete(&email).Error
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Email Admin"})
		return
	}

	// Commit the transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Email Admin deleted successfully"})
}
