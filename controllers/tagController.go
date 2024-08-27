package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetTags(c *gin.Context) {
	// Buat instance Car dari data yang di-bind user id diambil dari auth
	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)

	var tags []models.Tag
	if err := initializers.DB.
		Model(&tags).
		Joins("JOIN article_tags ON article_tags.tag_id = tags.id").
		Joins("JOIN articles ON article_tags.article_id = articles.id").
		Where("articles.user_id = ?", user.ID).
		Select("DISTINCT tags.id, tags.name").
		Order("tags.name").
		Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tags", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get tags data successfully", "data": tags})
}

func GetTagById(c *gin.Context) {

	// Ambil ID mobil dari parameter URL
	ID := c.Param("id")

	// Query detail mobil berdasarkan ID pengguna dan ID mobil
	var tag models.Tag
	if err := initializers.DB.Where("id = ?", ID).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get article tag", "tag": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get article tag successfully", "data": tag})
}
