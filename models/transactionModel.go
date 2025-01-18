package models

import "gorm.io/gorm"

type VoucherRedeem struct {
	gorm.Model
	VoucherID     uint    `gorm:"not null" json:"voucher_id"`
	Voucher       Voucher `gorm:"foreignKey:VoucherID;constraint:OnUpdate:CASCADE" json:"voucher,omitempty"`
	TransactionID uint    `gorm:"not null" json:"transaction_id"`
	Quantity      uint    `gorm:"not null" json:"quantity"`
	TotalPoints   uint    `gorm:"not null" json:"total_points"`
}

type Transaction struct {
	gorm.Model
	CustomerID   uint            `gorm:"not null" json:"customer_id"`
	TotalPoints  uint            `gorm:"not null" json:"total_points"`
	VoucherItems []VoucherRedeem `gorm:"foreignKey:TransactionID;constraint:OnUpdate:CASCADE" json:"voucher_items,omitempty"`
}
