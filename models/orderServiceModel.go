package models

import "gorm.io/gorm"

type OrderService struct {
	gorm.Model
	OrderID   uint    `json:"order_id"`
	ServiceID uint    `json:"service_id"`
	Order     Order   `json:"-" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Service   Service `json:"-" gorm:"foreignKey:ServiceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
