package models

type Service struct {
	ID   uint   `gorm:"primarykey"`
	Name string `json:"name" gorm:"not null;unique"`
}
