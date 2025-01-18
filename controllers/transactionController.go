package controllers

import (
	"errors"
	"net/http"
	"voucher-backend/initializers"
	"voucher-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RedeemVouchers(c *gin.Context) {
	var body struct {
		CustomerID uint `json:"customer_id" binding:"required"`
		Vouchers   []struct {
			VoucherID uint `json:"voucher_id" binding:"required"`
			Quantity  uint `json:"quantity" binding:"required,min=1"`
		} `json:"vouchers" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := initializers.DB.Begin()

	transaction := models.Transaction{
		CustomerID: body.CustomerID,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction", "details": err.Error()})
		return
	}

	var voucherRedeems []models.VoucherRedeem
	totalPoints := uint(0)

	for _, v := range body.Vouchers {
		var voucher models.Voucher
		if err := tx.Where("id = ?", v.VoucherID).First(&voucher).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				c.JSON(http.StatusNotFound, gin.H{"error": "Voucher not found", "voucher_id": v.VoucherID})
				return
			}
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Voucher details", "details": err.Error()})
			return
		}

		if voucher.Quantity < v.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough quantity for voucher", "voucher_id": v.VoucherID})
			return
		}

		voucherRedeem := models.VoucherRedeem{
			VoucherID:     v.VoucherID,
			TransactionID: transaction.ID,
			Quantity:      v.Quantity,
			TotalPoints:   voucher.Point * v.Quantity,
		}

		voucherRedeems = append(voucherRedeems, voucherRedeem)

		totalPoints += voucherRedeem.TotalPoints

		voucher.Quantity -= v.Quantity
		if err := tx.Save(&voucher).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update voucher quantity", "details": err.Error()})
			return
		}
	}

	if err := tx.Create(&voucherRedeems).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create voucher redeem records", "details": err.Error()})
		return
	}

	transaction.TotalPoints = totalPoints
	if err := tx.Save(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction total points", "details": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Vouchers redeemed successfully", "data": voucherRedeems, "transaction": transaction})
}
