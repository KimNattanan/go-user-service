package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/KimNattanan/go-user-service/pkg/database"
	"github.com/KimNattanan/go-user-service/pkg/httpserver"
	"github.com/KimNattanan/go-user-service/pkg/redisclient"
)

func Start() {
	cfg, db, rdb, sessionStore, err := SetupDependencies("development")
	if err != nil {
		log.Fatalf("failed to setup dependencies: %v", err)
	}

	r := SetupRestServer(db, rdb, sessionStore, cfg)

	srv := httpserver.Start(r, cfg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("Shutting down server...")

	if err := httpserver.Shutdown(srv); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}
	if err := database.Close(); err != nil {
		log.Printf("database close failed: %v", err)
	}
	if err := redisclient.Close(rdb); err != nil {
		log.Printf("redis client close failed: %v", err)
	}

	log.Println("Server shutted down.")
}
