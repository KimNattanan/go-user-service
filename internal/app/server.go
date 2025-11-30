package app

import (
	"log"
	"net/http"
)

func Start() {
	db, err := setupDependencies("development")
	if err != nil {
		log.Fatalf("failed to setup dependencies: %v", err)
	}
	r := setupRestServer(db)

	http.ListenAndServe(":8000", r)
}
