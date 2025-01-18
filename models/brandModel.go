package models

import "gorm.io/gorm"

type Brand struct {
	gorm.Model
	Name     string    `gorm:"not null" json:"name"`
	Vouchers []Voucher `gorm:"foreignKey:BrandID;constraint:OnUpdate:CASCADE" json:"vouchers,omitempty"`
}
