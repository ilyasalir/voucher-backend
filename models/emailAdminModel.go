package models

type EmailAdmin struct {
	ID    uint   `gorm:"primarykey"`
	Email string `json:"email" gorm:"not null, unique"`
}
