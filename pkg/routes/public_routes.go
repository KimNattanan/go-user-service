package routes

import (
	"os"

	"github.com/KimNattanan/go-user-service/internal/handler/rest"
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

func RegisterPublicRoutes(r *mux.Router, db *gorm.DB, rdb *redis.Client, sessionStore sessions.Store) {
	api := r.PathPrefix("/api/v2").Subrouter()

	jwtMaker := token.NewJWTMaker(os.Getenv("JWT_SECRET"))

	userRepo := userRepo.NewUserRepo(db)
	sessionRepo := sessionRepo.NewSessionRepo(rdb)

	userUsecase := userUsecase.NewUserUsecase(userRepo)
	sessionUsecase := sessionUsecase.NewSessionUsecase(sessionRepo)

	googleOauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	userHandler := rest.NewHttpUserHandler(userUsecase, sessionUsecase, sessionStore, googleOauthConfig, jwtMaker)

	authGroup := api.PathPrefix("/auth").Subrouter()
	authGroup.HandleFunc("/google/login", userHandler.GoogleLogin).Methods("GET")
	authGroup.HandleFunc("/google/callback", userHandler.GoogleCallback).Methods("GET")

	userGroup := api.PathPrefix("/users").Subrouter()
	userGroup.HandleFunc("/{id}", userHandler.FindUser).Methods("GET")
}
