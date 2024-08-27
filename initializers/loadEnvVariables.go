package initializers

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	// Menggunakan path absolut sesuai dengan petunjuk Render
	secretFilePath := ".env"

	err := godotenv.Load(secretFilePath)
	// err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file" + err.Error())
	}
}
