package application

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func (a *App) connectToDB() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("env variable not provided")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	a.DB = db

	fmt.Println("Connected to db...")

	return nil
}
