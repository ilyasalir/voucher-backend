package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"not null" json:"name"`
	Email    string `gorm:"unique;not null" json:"email"`
	Phone    string `gorm:"not null" json:"phone"`
	Password string `gorm:"not null" json:"-"`
}
