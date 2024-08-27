package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddCategory(c *gin.Context) {
	var body struct {
		Category string `form:"category" binding:"required"`
	}

	// Bind data JSON ke struct AddCarRequest
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert the category name to uppercase
	upperCategory := strings.ToUpper(body.Category)

	// Init transaction
	tx := initializers.DB.Begin()

	newCategory := models.Category{
		Name: upperCategory,
	}

	// Simpan car baru ke database
	if err := tx.Create(&newCategory).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add Category", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Category added successfully", "data": newCategory})
}

func GetCategory(c *gin.Context) {
	var category []models.Category

	if err := initializers.DB.Order("name").Find(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get category", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get category data successfully", "data": category})
}

func DeleteCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	// Start a transaction
	tx := initializers.DB.Begin()

	var category models.Category
	err = tx.First(&category, categoryID).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve category"})
		return
	}

	var article models.Article
	// Set the CategoryID to NULL for all related articles before deleting the category
	err = tx.Model(article).Where("category_id = ?", categoryID).Update("category_id", nil).Error
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update articles"})
		return
	}

	// Delete the category
	err = tx.Delete(&category).Error
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	// Commit the transaction
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

func GetCategoryIdByName(c *gin.Context) {

	// Ambil ID mobil dari parameter URL
	name := c.Param("name")

	// Query detail mobil berdasarkan ID pengguna dan ID mobil
	var category models.Category
	if err := initializers.DB.Where("name = ?", name).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get category details", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get category details successfully", "data": category})
}
