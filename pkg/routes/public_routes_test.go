package routes_test

import (
	"bytes"
	"encoding/json"
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
		body       map[string]string
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
		{
			name:   "register new user",
			method: http.MethodPost,
			path:   "/api/v1/auth/register",
			body: map[string]string{
				"email":       "test@gmail.com",
				"password":    "password123",
				"name":        "user name",
				"first_name":  "first",
				"last_name":   "last",
				"picture_url": "http://example.com/pic.jpg",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "login the created user",
			method: http.MethodPost,
			path:   "/api/v1/auth/login",
			body: map[string]string{
				"email":    "test@gmail.com",
				"password": "password123",
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "login with wrong password",
			method: http.MethodPost,
			path:   "/api/v1/auth/login",
			body: map[string]string{
				"email":    "test@gmail.com",
				"password": "password1234",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "login with wrong email",
			method: http.MethodPost,
			path:   "/api/v1/auth/login",
			body: map[string]string{
				"email":    "taste@gmail.com",
				"password": "password123",
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(body))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}
