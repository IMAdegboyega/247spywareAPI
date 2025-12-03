package config

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	dbType := os.Getenv("DB_TYPE")

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	switch dbType {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_SSLMODE"),
		)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	default:
		// Default to SQLite for development (pure Go driver, no CGO needed)
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "blog.db"
		}
		db, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}