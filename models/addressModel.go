package models

import "gorm.io/gorm"

type Address struct {
	gorm.Model
	UserID    uint   `json:"user_id" gorm:"not null"`
	Title     string `json:"title" gorm:"not null"`
	Location  string `json:"location" gorm:"not null"`
}

