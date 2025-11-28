package database

import (
	"fmt"
	"os"
	"time"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	var (
		host     = os.Getenv("DB_HOST")
		port     = os.Getenv("DB_PORT")
		user     = os.Getenv("DB_USER")
		password = os.Getenv("DB_PASSWORD")
		dbname   = os.Getenv("DB_NAME")
	)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var db *gorm.DB
	var err error
	for i := 1; i <= 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			fmt.Println("Connected to database")
			break
		}
		fmt.Printf("Database not ready, retrying in 2 seconds... (%d/10)\n", i)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	if err := db.Migrator().AutoMigrate(
		&entity.User{},
		&entity.Preference{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
