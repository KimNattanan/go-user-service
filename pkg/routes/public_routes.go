package routes

import (
	"github.com/KimNattanan/go-user-service/internal/handler/rest"
	userRepo "github.com/KimNattanan/go-user-service/internal/repo/user"
	userUsecase "github.com/KimNattanan/go-user-service/internal/usecase/user"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func RegisterPublicRoutes(r *mux.Router, db *gorm.DB) {
	userRepo := userRepo.NewUserRepo(db)
	userUsecase := userUsecase.NewUserUsecase(userRepo)
	userHandler := rest.NewHttpUserHandler(userUsecase)
	r.HandleFunc("/login", userHandler.GoogleLogin)
}
