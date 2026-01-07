package app

import (
	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/middleware"
	"github.com/KimNattanan/go-user-service/pkg/config"
	"github.com/KimNattanan/go-user-service/pkg/database"
	"github.com/KimNattanan/go-user-service/pkg/redisclient"
	"github.com/KimNattanan/go-user-service/pkg/routes"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	_ "github.com/KimNattanan/go-user-service/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func setupDependencies(env string) (*config.Config, *gorm.DB, *redis.Client, sessions.Store, error) {
	cfg := config.LoadConfig(env)

	db, err := database.Connect(cfg.DBDSN)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if env == "test" {
		db.Migrator().DropTable(
			&entity.User{},
			&entity.Preference{},
		)
	}
	if err := db.Migrator().AutoMigrate(
		&entity.User{},
		&entity.Preference{},
	); err != nil {
		return nil, nil, nil, nil, err
	}

	rdb := redisclient.Connect(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionKey))

	return cfg, db, rdb, sessionStore, nil
}

func setupRestServer(db *gorm.DB, rdb *redis.Client, sessionStore sessions.Store, cfg *config.Config) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORS)
	routes.RegisterPublicRoutes(r, db, rdb, sessionStore, cfg)
	routes.RegisterPrivateRoutes(r, db, rdb, sessionStore, cfg)
	routes.RegisterNotFoundRoute(r)
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	return r
}
