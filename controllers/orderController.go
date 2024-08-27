package controllers

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddOrder(c *gin.Context) {
	var body struct {
		CarID       uint      `json:"car_id" binding:"required"`
		ServiceType string    `json:"service_type" binding:"required"`
		Address     string    `json:"address" binding:"omitempty"`
		OrderTime   time.Time `json:"order_time" binding:"required"`
		Services    string    `json:"services" binding:"required"`
	}

	// Bind data JSON ke struct AddOrderRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi jam operasional bengkel (8.00 hingga 18.00)
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load location", "details": err.Error()})
		return
	}

	operationalStartTime := time.Date(body.OrderTime.Year(), body.OrderTime.Month(), body.OrderTime.Day(), 8, 0, 0, 0, location)
	operationalEndTime := time.Date(body.OrderTime.Year(), body.OrderTime.Month(), body.OrderTime.Day(), 18, 0, 0, 0, location)

	if body.OrderTime.Before(operationalStartTime) || body.OrderTime.Add(time.Duration(1)*time.Hour).After(operationalEndTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order time or duration, outside of operational hours"})
		return
	}

	// Init transaction
	tx := initializers.DB.Begin()

	// Cek apakah CarID valid
	var existingCar models.Car
	if result := tx.First(&existingCar, body.CarID); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CarID"})
		return
	}

	// Cek ketersediaan waktu
	interval := 1 // Interval dalam jam
	newOrderStartTime := body.OrderTime
	newOrderEndTime := body.OrderTime.Add(time.Duration(interval) * time.Hour)

	var overlappingOrdersCount int64
	var existingOrders []models.Order
	if err := tx.Model(&models.Order{}).Find(&existingOrders).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve existing orders", "details": err.Error()})
		return
	}

	for _, existingOrder := range existingOrders {
		existingOrderStartTime := existingOrder.OrderTime
		existingOrderEndTime := existingOrder.OrderTime.Add(time.Duration(existingOrder.Duration) * time.Hour)

		if ((newOrderStartTime.Before(existingOrderEndTime) && existingOrderStartTime.Before(newOrderEndTime)) ||
			(newOrderEndTime.After(existingOrderStartTime) && existingOrderStartTime.After(newOrderStartTime))) && existingOrder.Status != "CANCELED" {
			overlappingOrdersCount++
		}
	}

	if overlappingOrdersCount >= 2 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exceeded maximum orders for the given time"})
		return
	}

	// Membuat order baru
	newOrder := models.Order{
		UserID:      existingCar.UserID,
		CarID:       body.CarID,
		ServiceType: body.ServiceType,
		Address:     body.Address,
		OrderTime:   body.OrderTime,
		Services:    body.Services,
	}

	// Simpan order baru ke database
	if result := tx.Create(&newOrder); result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add order", "details": result.Error.Error()})
		return
	}

	// Preload data Car dan Services saat mengembalikan respons
	if err := tx.Preload("Car.CarType.Brand").Preload("Car.Color").First(&newOrder).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preload related data", "details": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Order added successfully", "data": newOrder})
}

func DeleteOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Cari order berdasarkan ID
	var order models.Order
	err = initializers.DB.First(&order, orderID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order"})
		return
	}

	// Hapus order dari database
	err = initializers.DB.Unscoped().Delete(&order).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

func UpdateStatusOrder(c *gin.Context) {
	var body struct {
		Status   string `json:"status" binding:"required"`
		Duration uint64 `json:"duration" binding:"omitempty"`
	}

	// Bind data JSON ke struct AddAddressRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Cari order berdasarkan ID
	var order models.Order
	err = initializers.DB.First(&order, orderID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order"})
		return
	}

	// Validasi jam operasional bengkel (8.00 hingga 18.00)
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load location", "details": err.Error()})
		return
	}

	operationalStartTime := time.Date(order.OrderTime.Year(), order.OrderTime.Month(), order.OrderTime.Day(), 8, 0, 0, 0, location)
	operationalEndTime := time.Date(order.OrderTime.Year(), order.OrderTime.Month(), order.OrderTime.Day(), 18, 0, 0, 0, location)

	if order.OrderTime.Before(operationalStartTime) || order.OrderTime.Add(time.Duration(body.Duration)*time.Hour).After(operationalEndTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order time or duration, outside of operational hours"})
		return
	}

	// Check overlapping orders
	if body.Duration > 0 {
		newOrderStartTime := order.OrderTime
		newOrderEndTime := order.OrderTime.Add(time.Duration(body.Duration) * time.Hour)

		var overlappingOrdersCount int64
		var existingOrders []models.Order
		if err := initializers.DB.Model(&models.Order{}).Where("ID != ?", orderID).Find(&existingOrders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve existing orders", "details": err.Error()})
			return
		}

		for _, existingOrder := range existingOrders {
			existingOrderStartTime := existingOrder.OrderTime
			existingOrderEndTime := existingOrder.OrderTime.Add(time.Duration(existingOrder.Duration) * time.Hour)

			if ((newOrderStartTime.Before(existingOrderEndTime) && existingOrderStartTime.Before(newOrderEndTime)) ||
				(newOrderEndTime.After(existingOrderStartTime) && existingOrderStartTime.After(newOrderStartTime))) && existingOrder.Status != "CANCELED" {
				overlappingOrdersCount++
			}
		}

		if overlappingOrdersCount >= 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Exceeded maximum orders for the given time"})
			return
		}

		order.Duration = body.Duration
	}

	order.Status = body.Status

	// Save order ke database
	if err = initializers.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated to " + body.Status + " successfully"})
}

func UpdatePriceOrder(c *gin.Context) {
	var body struct {
		Status string `json:"status" binding:"required"`
		Price  uint64 `json:"price" binding:"required"`
	}

	// Bind data JSON ke struct AddAddressRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Cari order berdasarkan ID
	var order models.Order
	err = initializers.DB.First(&order, orderID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order"})
		return
	}

	order.Price = body.Price
	order.Status = body.Status

	// Save order ke database
	if err = initializers.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order price"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order price updated successfully"})
}

func GetOrder(c *gin.Context) {
	dateString := c.Query("date")
	userID := c.Query("user_id")
	carId := c.Query("car_id")

	tx := initializers.DB.Begin()
	var orders []models.Order

	// Filter berdasarkan user_id dan date_string
	query := tx.Preload("User").Preload("Car.CarType.Brand").Preload("Car.Color")

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if dateString != "" {
		query = query.Where("CAST(order_time as DATE) = ?", dateString)
	}

	if carId != "" {
		query = query.Where("car_id = ?", carId)
	}

	// Eksekusi query dan urutkan hasil berdasarkan order_time
	if err := query.Order("order_time desc").Find(&orders).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders", "details": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Get orders data successfully", "data": orders})
}

func GetOrderUsed(c *gin.Context) {
	dateString := c.Query("date")

	tx := initializers.DB.Begin()
	var orders []models.Order

	// Filter berdasarkan date_string
	query := tx
	if dateString != "" {
		query = query.Where("CAST(order_time as DATE) = ?", dateString).Select("order_time, duration, status")
	}

	// Eksekusi query dan urutkan hasil berdasarkan order_time
	if err := query.Order("order_time desc").Find(&orders).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders used", "details": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Get orders used successfully", "data": orders})
}
