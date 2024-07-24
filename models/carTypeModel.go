package models

type CarType struct {
	ID      uint   `gorm:"primarykey"`
	Name    string `json:"name" gorm:"not null"`
	BrandID uint   `json:"brand_id" gorm:"not null"`
	Brand   Brand  `json:"brand" gorm:"foreignKey:BrandID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
