package models

import (
	"gorm.io/gorm"
)

type Car struct {
	gorm.Model
	LicensePlat  string  `json:"license_plat" gorm:"unique"`
	UserID       uint    `json:"user_id" gorm:"not null"`
	User         User    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
	CarTypeID    uint    `json:"car_type_id" gorm:"not null"`
	ColorID      uint    `json:"color_id" gorm:"not null"`
	FrameNumber  string  `json:"frame_number"`
	EngineNumber string  `json:"engine_number"`
	Kilometer    uint64  `json:"kilometer"`
	PhotoUrl     string  `json:"photo_url"`
	CarType      CarType `json:"car_type" gorm:"foreignKey:CarTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Color        Color   `json:"color" gorm:"foreignKey:ColorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
