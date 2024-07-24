package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Phone     string    `gorm:"not null" json:"phone"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"default:USER" json:"role"`
	Addresses []Address `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"addresses,omitempty"`
	Orders    []Order   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"orders,omitempty"`
	Cars      []Car     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"cars,omitempty"`
	Stnks     []Stnk    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"stnks,omitempty"`
}
