package models

type Color struct {
	ID   uint   `gorm:"primarykey"`
	Name string `json:"name" gorm:"not null"`
}
