package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/usecase"
	"github.com/KimNattanan/go-user-service/pkg/apperror"
	"github.com/KimNattanan/go-user-service/pkg/token"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type AuthMiddleware struct {
	userUsecase       usecase.UserUsecase
	sessionUsecase    usecase.SessionUsecase
	sessionStore      sessions.Store
	jwtMaker          *token.JWTMaker
	googleOauthConfig *oauth2.Config
	jwtExpiration     time.Duration
}

func NewAuthMiddleware(userUsecase usecase.UserUsecase, sessionUsecase usecase.SessionUsecase, sessionStore sessions.Store, jwtMaker *token.JWTMaker, googleOauthConfig *oauth2.Config, jwtExpiration int) *AuthMiddleware {
	return &AuthMiddleware{
		userUsecase:       userUsecase,
		sessionUsecase:    sessionUsecase,
		sessionStore:      sessionStore,
		jwtMaker:          jwtMaker,
		googleOauthConfig: googleOauthConfig,
		jwtExpiration:     time.Duration(jwtExpiration),
	}
}

func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieSession, err := m.sessionStore.Get(r, "session")
		if err != nil {
			http.Error(w, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}
		accessToken, _ := cookieSession.Values["access_token"].(string)
		accessClaims, err := m.jwtMaker.VerfiyToken(accessToken)
		if err == nil {
			ctx := context.WithValue(r.Context(), "userID", accessClaims.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		refreshToken, _ := cookieSession.Values["refresh_token"].(string)
		refreshClaims, err := m.jwtMaker.VerfiyToken(refreshToken)
		if err != nil {
			http.Error(w, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}
		user, err := m.userUsecase.FindByID(r.Context(), refreshClaims.ID)
		if err != nil || user == nil {
			http.Error(w, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}
		session, err := m.sessionUsecase.FindByID(r.Context(), refreshClaims.RegisteredClaims.ID)
		if err != nil || session == nil || session.IsRevoked {
			http.Error(w, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}
		if err := m.sessionUsecase.Revoke(r.Context(), session.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		go func() { // update user's info
			client := m.googleOauthConfig.Client(r.Context(), &oauth2.Token{
				RefreshToken: session.GoogleRefreshToken,
			})
			clientRes, _ := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
			if err != nil {
				return
			}
			defer clientRes.Body.Close()
			var userInfo map[string]interface{}
			if err := json.NewDecoder(clientRes.Body).Decode(&userInfo); err != nil {
				return
			}
			firstName, _ := userInfo["given_name"].(string)
			lastName, _ := userInfo["family_name"].(string)
			pictureURL, _ := userInfo["picture"].(string)
			m.userUsecase.Update(r.Context(), user.ID, map[string]interface{}{
				"first_name":  firstName,
				"last_name":   lastName,
				"picture_url": pictureURL,
			})
		}()

		refreshToken, refreshClaims, err = m.jwtMaker.CreateToken(user.ID, time.Second*m.jwtExpiration)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		accessToken, accessClaims, err = m.jwtMaker.CreateToken(user.ID, time.Hour)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newSession := &entity.Session{
			ID:                 refreshClaims.RegisteredClaims.ID,
			UserID:             user.ID,
			GoogleRefreshToken: session.GoogleRefreshToken,
			IsRevoked:          false,
			CreatedAt:          session.CreatedAt,
			ExpiresAt:          refreshClaims.RegisteredClaims.ExpiresAt.Time,
		}
		if err := m.sessionUsecase.Create(r.Context(), newSession); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cookieSession.Values["refresh_token"] = refreshToken
		cookieSession.Values["access_token"] = accessToken
		if err := cookieSession.Save(r, w); err != nil {
			http.Error(w, err.Error(), apperror.StatusCode(err))
			return
		}
		ctx := context.WithValue(r.Context(), "userID", user.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
