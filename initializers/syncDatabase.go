package initializers

import "voucher-backend/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.Brand{}, &models.Voucher{}, &models.Transaction{}, &models.VoucherRedeem{})
}
