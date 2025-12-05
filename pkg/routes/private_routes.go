package routes

import (
	"os"

	"github.com/KimNattanan/go-user-service/internal/handler/rest"
	"github.com/KimNattanan/go-user-service/internal/middleware"
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

func RegisterPrivateRoutes(r *mux.Router, db *gorm.DB, rdb *redis.Client, sessionStore sessions.Store) {
	api := r.PathPrefix("/api/v2").Subrouter()

	jwtMaker := token.NewJWTMaker(os.Getenv("JWT_SECRET"))
	googleOauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	userRepo := userRepo.NewUserRepo(db)
	sessionRepo := sessionRepo.NewSessionRepo(rdb)
	preferenceRepo := preferenceRepo.NewPreferenceRepo(db)

	userUsecase := userUsecase.NewUserUsecase(userRepo)
	sessionUsecase := sessionUsecase.NewSessionUsecase(sessionRepo)
	preferenceUsecase := preferenceUsecase.NewPreferenceUsecase(preferenceRepo)

	userHandler := rest.NewHttpUserHandler(userUsecase, sessionUsecase, sessionStore, googleOauthConfig, jwtMaker)
	preferenceHandler := rest.NewHttpPreferenceHandler(preferenceUsecase)

	authMiddleware := middleware.NewAuthMiddleware(userUsecase, sessionUsecase, sessionStore, jwtMaker, googleOauthConfig)
	api.Use(authMiddleware.Handle)

	userGroup := api.PathPrefix("/user").Subrouter()
	userGroup.HandleFunc("/me", userHandler.GetUser)
	userGroup.HandleFunc("/me", userHandler.Update).Methods("PATCH")
	userGroup.HandleFunc("/me/logout", userHandler.Logout).Methods("POST")
	userGroup.HandleFunc("/me", userHandler.Delete).Methods("DELETE")

	preferenceGroup := api.PathPrefix("/preference").Subrouter()
	preferenceGroup.HandleFunc("/", preferenceHandler.GetPreference)
	preferenceGroup.HandleFunc("/", preferenceHandler.Update).Methods("PATCH")
}
