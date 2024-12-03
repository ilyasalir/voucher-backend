package models

import "gorm.io/gorm"

type Inquiry struct {
	gorm.Model
	CarBrand string `gorm:"not null" json:"cars_brand"`
	CarYear  string `gorm:"not null" json:"cars_year"`
	Problem  string `gorm:"not null" json:"problem"`
	Phone    string `gorm:"not null" json:"phone"`
	Status   *bool  `gorm:"default:false" json:"status"`
}
