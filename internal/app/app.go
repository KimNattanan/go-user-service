package app

import (
	"log"
	"os"

	"github.com/KimNattanan/go-user-service/internal/middleware"
	"github.com/KimNattanan/go-user-service/pkg/database"
	"github.com/KimNattanan/go-user-service/pkg/redisclient"
	"github.com/KimNattanan/go-user-service/pkg/routes"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	_ "github.com/KimNattanan/go-user-service/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func setupDependencies(env string) (*gorm.DB, *redis.Client, sessions.Store, error) {
	envFile := ".env"
	if env != "" {
		envFile = ".env." + env
	}
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	}

	db, err := database.Connect()
	if err != nil {
		return nil, nil, nil, err
	}

	rdb := redisclient.Connect()

	sessionStore := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

	return db, rdb, sessionStore, nil
}

func setupRestServer(db *gorm.DB, rdb *redis.Client, sessionStore sessions.Store) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORS)
	routes.RegisterPublicRoutes(r, db, rdb, sessionStore)
	routes.RegisterPrivateRoutes(r, db, rdb, sessionStore)
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	return r
}
