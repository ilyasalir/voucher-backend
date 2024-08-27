package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"carport-backend/utils"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddArticle(c *gin.Context) {
	var body struct {
		Category *uint  `form:"category" binding:"required"`
		Title    string `form:"title" binding:"required"`
		Content  string `form:"content" binding:"required"`
		Status   *bool  `form:"status" binding:"required"`
	}

	// Bind data JSON ke struct AddCarRequest
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Bind the tags separately as JSON
	// Parse the tags field manually from the form data
	tagsStr := c.PostForm("tags")
	var tags []string
	if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse tags", "details": err.Error()})
		return
	}

	// Init transaction
	tx := initializers.DB.Begin()

	// Mendapatkan file dari form
	file, err := c.FormFile("photo")
	var photoUrl string

	// Menyimpan file dan mendapatkan URL-nya
	// photoUrl, err := utils.HandleFileUpload(c, file)
	// if err != nil {
	// 	tx.Rollback()
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
	// 	return
	// }

	// Jika file foto ada, maka lakukan proses upload
	if err == nil && file != nil {
		// Proses upload file ke Firebase
		photoUrl, err = utils.UploadFileToFirebase(*file)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload photo failed"})
			return
		}
	}

	// Buat instance Car dari data yang di-bind user id diambil dari auth
	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)
	newArticle := models.Article{
		UserID:     user.ID,
		PhotoUrl:   photoUrl,
		CategoryID: body.Category,
		Title:      body.Title,
		Content:    body.Content,
		Status:     body.Status,
	}

	if body.Status != nil && *body.Status {
		now := time.Now()
		newArticle.PublishDate = &now
	}

	var articleTags []models.Tag
	for _, tagName := range tags {
		// Gunakan title case saat memeriksa keberadaan layanan
		titleCaseName := strings.Title(tagName)
		tag := models.Tag{Name: titleCaseName}

		if result := tx.FirstOrCreate(&tag, models.Tag{Name: titleCaseName}); result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add service", "details": result.Error.Error()})
			return
		}
		articleTags = append(articleTags, tag)
	}

	newArticle.Tags = articleTags

	// Simpan car baru ke database
	if err := tx.Create(&newArticle).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add article", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Article added successfully", "data": newArticle})
}

func GetAllArticle(c *gin.Context) {

	var article []models.Article
	if err := initializers.DB.Preload("User").Preload("Category").Preload("Tags").Find(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get articles", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get articles successfully",
		"data":    article,
	})
}

func GetArticleDetail(c *gin.Context) {

	// Ambil ID pengguna dari konteks (biasanya diambil dari token otentikasi)
	var user models.User
	userInterface, _ := c.Get("user")
	user, _ = userInterface.(models.User)

	// Ambil ID mobil dari parameter URL
	articleId := c.Param("id")

	// Query detail mobil berdasarkan ID pengguna dan ID mobil
	var article models.Article
	if err := initializers.DB.Where("user_id = ? AND ID = ?", user.ID, articleId).Preload("User").Preload("Category").Preload("Tags").First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Article details", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get Article details successfully", "data": article})
}

func UpdateStatus(c *gin.Context) {

	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid car ID"})
		return
	}

	tx := initializers.DB.Begin()

	var existingArticle models.Article
	// Cari article berdasarkan ID
	if err = tx.First(&existingArticle, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Article", "details": err.Error()})
		return
	}

	status := !*existingArticle.Status
	existingArticle.Status = &status

	if status {
		now := time.Now()
		existingArticle.PublishDate = &now
	} else {
		existingArticle.PublishDate = nil
	}

	// Simpan perubahan ke database
	if err := tx.Save(&existingArticle).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Article", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Verified successfully", "data": existingArticle})
}

func DeleteArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var article models.Article
	err = initializers.DB.First(&article, articleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve article"})
		return
	}

	err = initializers.DB.Unscoped().Delete(&article).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "article and associated orders and cars deleted successfully"})
}

func EditArticle(c *gin.Context) {
	var body struct {
		Category *uint    `json:"category"`
		Title    string   `json:"title"`
		Content  string   `json:"content"`
		Status   *bool    `json:"status"`
		Tags     []string `json:"tags"`
	}

	// Bind data form ke struct EditArticleRequest
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	articleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	// Init transaction
	tx := initializers.DB.Begin()

	var existingArticle models.Article
	// Cari artikel berdasarkan ID
	if err = tx.Preload("Tags").First(&existingArticle, articleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve article", "details": err.Error()})
		return
	}

	// Mendapatkan file dari form
	file, err := c.FormFile("photo")
	if err == nil && file != nil { // jika foto ada
		// Jika ada file baru diunggah
		photoUrl, err := utils.UploadFileToFirebase(*file)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Hapus foto yang lama hanya jika foto baru diunggah
		if existingArticle.PhotoUrl != "" {
			err = utils.DeleteFileFromFirebase(existingArticle.PhotoUrl)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		existingArticle.PhotoUrl = photoUrl
	}

	// Update data artikel dengan nilai yang baru
	if body.Category != nil {
		existingArticle.CategoryID = body.Category
	}
	if body.Title != "" {
		existingArticle.Title = body.Title
	}
	if body.Content != "" {
		existingArticle.Content = body.Content
	}
	if body.Status != nil {
		existingArticle.Status = body.Status
	}

	// Update tags
	if len(body.Tags) > 0 {
		var articleTags []models.Tag
		for _, tagName := range body.Tags {
			// Gunakan title case saat memeriksa keberadaan tag
			titleCaseName := strings.Title(tagName)
			tag := models.Tag{Name: titleCaseName}

			if result := tx.FirstOrCreate(&tag, models.Tag{Name: titleCaseName}); result.Error != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tag", "details": result.Error.Error()})
				return
			}
			articleTags = append(articleTags, tag)
		}
		existingArticle.Tags = articleTags
	}

	// Simpan perubahan ke database
	if err := tx.Save(&existingArticle).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit article", "details": err.Error()})
		return
	}

	// Preload data User, Category, dan Tags saat mengembalikan respons
	if err := tx.Preload("User").Preload("Category").Preload("Tags").First(&existingArticle).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preload related data", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Article edited successfully", "data": existingArticle})
}

func GetArticleDetails(c *gin.Context) {

	// Ambil ID mobil dari parameter URL
	ID := c.Param("id")

	// Query detail mobil berdasarkan ID pengguna dan ID mobil
	var article models.Article
	if err := initializers.DB.Where("id = ?", ID).Preload("Category").Preload("Tags").First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get article detailssssssssss", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get article details successfully", "data": article})
}

func GetArticleById(c *gin.Context) {

	var body struct {
		ArticleID []uint `json:"article_id" binding:"required"`
	}

	// Bind data JSON ke struct AddCarRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Query detail mobil berdasarkan ID pengguna dan ID mobil
	var article []models.Article
	if err := initializers.DB.Where("id IN ?", body.ArticleID).Preload("Category").Preload("Tags").Find(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get article details", "details": body.ArticleID})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get article details successfully", "data": article})
}

func GetArticleByCategory(c *gin.Context) {

	// Ambil ID mobil dari parameter URL
	ID := c.Param("id")

	// Query detail mobil berdasarkan ID pengguna dan ID mobil
	var article []models.Article
	if err := initializers.DB.Where("category_id = ?", ID).Preload("Category").Preload("Tags").Find(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get article details", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get article details successfully", "data": article})
}

func GetArticleByTag(c *gin.Context) {

	// Ambil ID mobil dari parameter URL
	ID := c.Param("id")

	// Query detail mobil berdasarkan ID pengguna dan ID mobil
	var article []models.ArticleTag
	if err := initializers.DB.Where("tag_id = ?", ID).Preload("Tags").Preload("Article").Find(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get article details", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Get article details successfully", "data": article})
}
