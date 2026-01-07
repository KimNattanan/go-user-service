package database

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db    *gorm.DB
	sqlDB *sql.DB
)

func Connect(dsn string) (*gorm.DB, error) {
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
	sqlDB, err = db.DB()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Close() error {
	fmt.Println("Closing Database connection...")
	if sqlDB != nil {
		return sqlDB.Close()
	}
	return nil
}
