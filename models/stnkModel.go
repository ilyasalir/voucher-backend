package models

import (
	"gorm.io/gorm"
)

type Stnk struct {
	gorm.Model
	ID          uint   `gorm:"primarykey"`
	UserID      uint   `json:"user_id" gorm:"not null"`
	User        User   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
	PhotoUrl    string `json:"photo_url"`
	Status      *bool  `json:"status" gorm:"default:false"`
	Description string `json:"desc"`
}
