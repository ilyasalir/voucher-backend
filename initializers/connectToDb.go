package initializers

import (
	"database/sql"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDb() {
	var err error

	dsn := os.Getenv("DB")
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Connection pool configurations
	sqlDB.SetMaxOpenConns(10)                  // Lower the max open connections
	sqlDB.SetMaxIdleConns(10)                  // Lower the max idle connections
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Reuse connections for 30 minutes
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Idle connections released after 10 minutes

	const maxRetries = 3
	retryCount := 0

	for {
		DB, err = gorm.Open(postgres.New(postgres.Config{
			Conn: sqlDB,
		}), &gorm.Config{})
		if err == nil || retryCount >= maxRetries {
			break
		}
		retryCount++
		log.Printf("Retrying to connect to the database (%d/%d)...", retryCount, maxRetries)
		time.Sleep(2 * time.Second) // Wait before retrying
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", retryCount, err)
	}
}
