package entity

import "time"

type Session struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	GoogleRefreshToken string    `json:"google_refresh_token"`
	IsRevoked          bool      `json:"is_revoked"`
	CreatedAt          time.Time `json:"created_at"`
	ExpiresAt          time.Time `json:"expires_at"`
}
