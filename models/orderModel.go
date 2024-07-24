package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	UserID      uint      `json:"user_id" gorm:"not null;constraint:OnDelete:CASCADE;"`
	CarID       uint      `json:"car_id"`
	ServiceType string    `json:"service_type" gorm:"not null"`
	Address     string    `json:"address"`
	OrderTime   time.Time `json:"order_time" gorm:"not null"`
	Price       uint64    `json:"price" gorm:"default:0"`
	Duration    uint64    `json:"duration" gorm:"default:1"`
	Status      string    `json:"status" gorm:"default:PENDING"`
	User        User      `json:"user" gorm:"foreignKey:UserID;"`
	Car         Car       `json:"car" gorm:"foreignKey:CarID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Services    []Service `json:"services" gorm:"many2many:order_services;"`
}
