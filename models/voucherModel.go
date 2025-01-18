package models

import "gorm.io/gorm"

type Voucher struct {
	gorm.Model
	Name     string `gorm:"not null" json:"name"`
	Discount uint   `gorm:"not null" json:"discount"`
	Quantity uint   `gorm:"" json:"quantity"`
	Point    uint   `gorm:"" json:"point"`
	BrandID  uint   `gorm:"not null" json:"brand_id"`
	Brand    Brand  `gorm:"foreignKey:BrandID;constraint:OnUpdate:CASCADE" json:"brand,omitempty"`
}
