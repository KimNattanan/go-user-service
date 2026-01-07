package httpserver

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/KimNattanan/go-user-service/pkg/config"
	"github.com/gorilla/mux"
)

func Start(r *mux.Router, cfg *config.Config) *http.Server {
	log.Println("Starting REST server on port:", cfg.AppPort)
	srv := &http.Server{
		Addr:         "0.0.0.0:" + cfg.AppPort,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("REST server error: %v", err)
	}
	return srv
}

func Shutdown(srv *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
