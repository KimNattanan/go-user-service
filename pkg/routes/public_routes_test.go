package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KimNattanan/go-user-service/internal/app"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func setupTestApp(t *testing.T) *mux.Router {
	err := godotenv.Load("../../.env.test")
	if err != nil {
		t.Fatalf("Failed to load .env.test: %v", err)
	}

	cfg, db, rdb, sessionStore, err := app.SetupDependencies("test")
	if err != nil {
		t.Fatalf("failed to setup dependencies: %v", err)
	}

	r := app.SetupRestServer(db, rdb, sessionStore, cfg)

	return r
}

func TestPublicRoutes(t *testing.T) {
	r := setupTestApp(t)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "GET users",
			method:     http.MethodGet,
			path:       "/api/v1/users",
			wantStatus: http.StatusOK,
		},
		{
			name:       "unknown route",
			method:     http.MethodGet,
			path:       "/not-found",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}
