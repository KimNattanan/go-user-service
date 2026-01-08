package routes

import (
	"github.com/KimNattanan/go-user-service/internal/handler/rest"
	"github.com/KimNattanan/go-user-service/internal/middleware"
	"github.com/KimNattanan/go-user-service/pkg/config"
	"github.com/KimNattanan/go-user-service/pkg/token"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	userRepo "github.com/KimNattanan/go-user-service/internal/repo/user"
	userUsecase "github.com/KimNattanan/go-user-service/internal/usecase/user"

	sessionRepo "github.com/KimNattanan/go-user-service/internal/repo/session"
	sessionUsecase "github.com/KimNattanan/go-user-service/internal/usecase/session"

	preferenceRepo "github.com/KimNattanan/go-user-service/internal/repo/preference"
	preferenceUsecase "github.com/KimNattanan/go-user-service/internal/usecase/preference"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"
)

func RegisterPrivateRoutes(r *mux.Router, db *gorm.DB, rdb *redis.Client, sessionStore sessions.Store, cfg *config.Config) {
	api := r.PathPrefix("/api/v1").Subrouter()

	jwtMaker := token.NewJWTMaker(cfg.JWTSecret)
	googleOauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	userRepo := userRepo.NewUserRepo(db)
	sessionRepo := sessionRepo.NewSessionRepo(rdb)
	preferenceRepo := preferenceRepo.NewPreferenceRepo(db)

	userUsecase := userUsecase.NewUserUsecase(userRepo)
	sessionUsecase := sessionUsecase.NewSessionUsecase(sessionRepo)
	preferenceUsecase := preferenceUsecase.NewPreferenceUsecase(preferenceRepo)

	userHandler := rest.NewHttpUserHandler(userUsecase, sessionUsecase, sessionStore, googleOauthConfig, jwtMaker, cfg.JWTExpiration)
	preferenceHandler := rest.NewHttpPreferenceHandler(preferenceUsecase)

	authMiddleware := middleware.NewAuthMiddleware(userUsecase, sessionUsecase, sessionStore, jwtMaker, googleOauthConfig, cfg.JWTExpiration)
	api.Use(authMiddleware.Handle)

	authGroup := api.PathPrefix("/auth").Subrouter()
	authGroup.HandleFunc("/logout", userHandler.Logout).Methods("POST")

	meGroup := api.PathPrefix("/me").Subrouter()
	meGroup.HandleFunc("", userHandler.GetUser).Methods("GET")
	meGroup.HandleFunc("", userHandler.Update).Methods("PATCH")
	meGroup.HandleFunc("", userHandler.Delete).Methods("DELETE")

	preferencesGroup := meGroup.PathPrefix("/preferences").Subrouter()
	preferencesGroup.HandleFunc("", preferenceHandler.GetPreference).Methods("GET")
	preferencesGroup.HandleFunc("", preferenceHandler.Update).Methods("PATCH")
}
