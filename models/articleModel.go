package models

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	ID          uint       `gorm:"primarykey"`
	UserID      uint       `json:"user_id" gorm:"not null"`
	CategoryID  *uint      `json:"category_id" gorm:"default:null"`
	User        User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
	PhotoUrl    string     `json:"photo_url"`
	Status      *bool      `json:"status"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	PublishDate *time.Time `json:"publish_date"`
	Category    Category   `json:"category" gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE"`
	Tags        []Tag      `json:"tags" gorm:"many2many:article_tags;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
