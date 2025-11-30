package app

import (
	"log"

	"github.com/KimNattanan/go-user-service/pkg/database"
	"github.com/KimNattanan/go-user-service/pkg/routes"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func setupDependencies(env string) (*gorm.DB, error) {
	envFile := ".env"
	if env != "" {
		envFile = ".env." + env
	}
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	}

	db, err := database.Connect()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupRestServer(db *gorm.DB) *mux.Router {
	r := mux.NewRouter()
	routes.RegisterPublicRoutes(r, db)
	return r
}
