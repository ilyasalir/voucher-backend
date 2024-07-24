package models

import "gorm.io/gorm"

type Brand struct {
	gorm.Model
	Name string `json:"name" gorm:"not null"`
}
