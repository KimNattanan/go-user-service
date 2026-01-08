package routes

import (
	"github.com/KimNattanan/go-user-service/internal/handler/rest"
	"github.com/KimNattanan/go-user-service/pkg/config"
	"github.com/KimNattanan/go-user-service/pkg/token"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	userRepo "github.com/KimNattanan/go-user-service/internal/repo/user"
	userUsecase "github.com/KimNattanan/go-user-service/internal/usecase/user"

	sessionRepo "github.com/KimNattanan/go-user-service/internal/repo/session"
	sessionUsecase "github.com/KimNattanan/go-user-service/internal/usecase/session"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"
)

func RegisterPublicRoutes(r *mux.Router, db *gorm.DB, rdb *redis.Client, sessionStore sessions.Store, cfg *config.Config) {
	api := r.PathPrefix("/api/v1").Subrouter()

	jwtMaker := token.NewJWTMaker(cfg.JWTSecret)

	userRepo := userRepo.NewUserRepo(db)
	sessionRepo := sessionRepo.NewSessionRepo(rdb)

	userUsecase := userUsecase.NewUserUsecase(userRepo)
	sessionUsecase := sessionUsecase.NewSessionUsecase(sessionRepo)

	googleOauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	userHandler := rest.NewHttpUserHandler(userUsecase, sessionUsecase, sessionStore, googleOauthConfig, jwtMaker, cfg.JWTExpiration)

	authGroup := api.PathPrefix("/auth").Subrouter()
	authGroup.HandleFunc("/register", userHandler.Register).Methods("POST")
	authGroup.HandleFunc("/login", userHandler.Login).Methods("POST")
	authGroup.HandleFunc("/google/login", userHandler.GoogleLogin).Methods("GET")
	authGroup.HandleFunc("/google/callback", userHandler.GoogleCallback).Methods("GET")

	userGroup := api.PathPrefix("/users").Subrouter()
	userGroup.HandleFunc("", userHandler.FindAllUsers).Methods("GET")
	userGroup.HandleFunc("/{id}", userHandler.FindUser).Methods("GET")
}
