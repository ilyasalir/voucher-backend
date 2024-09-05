package initializers

import "carport-backend/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{}, &models.Address{}, &models.Brand{}, &models.CarType{}, &models.Color{}, &models.Car{}, &models.Service{}, &models.Order{}, &models.OrderService{}, &models.Stnk{}, &models.Article{}, &models.Category{}, &models.Tag{}, &models.ArticleTag{}, &models.EmailAdmin{})
}
