package models

import "gorm.io/gorm"

type ArticleTag struct {
	gorm.Model
	ArticleID uint    `json:"article_id"`
	TagID     uint    `json:"tag_id"`
	Article   Article `json:"article" gorm:"foreignKey:ArticleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Tags      Tag     `json:"tags" gorm:"foreignKey:TagID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
