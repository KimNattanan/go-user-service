package rest

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/usecase"
	"github.com/KimNattanan/go-user-service/pkg/apperror"
	"github.com/KimNattanan/go-user-service/pkg/token"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type HttpUserHandler struct {
	userUsecase       usecase.UserUsecase
	sessionUsecase    usecase.SessionUsecase
	sessionStore      sessions.Store
	googleOauthConfig *oauth2.Config
	jwtMaker          *token.JWTMaker
}

func NewHttpUserHandler(userUsecase usecase.UserUsecase, sessionUsecase usecase.SessionUsecase, sessionStore sessions.Store, googleOauthConfig *oauth2.Config, jwtMaker *token.JWTMaker) *HttpUserHandler {
	return &HttpUserHandler{
		userUsecase:       userUsecase,
		sessionUsecase:    sessionUsecase,
		sessionStore:      sessionStore,
		googleOauthConfig: googleOauthConfig,
		jwtMaker:          jwtMaker,
	}
}

func (h *HttpUserHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	})
	url := h.googleOauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "select_account"))

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusFound)
}

func (h *HttpUserHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if state, err := r.Cookie("oauthstate"); err != nil || state.Value != query.Get("state") {
		http.Error(w, "invalid oauth state", http.StatusUnauthorized)
		return
	}
	code := query.Get("code")
	if code == "" {
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}
	token, err := h.googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusUnauthorized)
		return
	}

	client := h.googleOauthConfig.Client(r.Context(), token)
	clientRes, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "failed to get user info", apperror.StatusCode(err))
		return
	}
	defer clientRes.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(clientRes.Body).Decode(&userInfo); err != nil {
		http.Error(w, "failed to decode user info", apperror.StatusCode(err))
		return
	}

	user, err := h.userUsecase.LoginOrRegisterWithGoogle(r.Context(), userInfo)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	refreshToken, refreshClaims, err := h.jwtMaker.CreateToken(user.ID, time.Hour*24*30)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	accessToken, _, err := h.jwtMaker.CreateToken(user.ID, time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session := &entity.Session{
		ID:                 refreshClaims.RegisteredClaims.ID,
		UserID:             user.ID,
		GoogleRefreshToken: token.RefreshToken,
		IsRevoked:          false,
		CreatedAt:          time.Now(),
		ExpiresAt:          refreshClaims.RegisteredClaims.ExpiresAt.Time,
	}
	if err := h.sessionUsecase.Create(r.Context(), session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   false,
	})
	cookieSession, _ := h.sessionStore.Get(r, "session")
	cookieSession.Values["access_token"] = accessToken
	cookieSession.Values["refresh_token"] = refreshToken
	if err := cookieSession.Save(r, w); err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	w.Header().Set("Location", os.Getenv("FRONTEND_OAUTH_REDIRECT_URL"))
	w.WriteHeader(http.StatusSeeOther)
}

func (h *HttpUserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	userID, _ := ctx.Value("userID").(string)
	user, err := h.userUsecase.FindByID(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}
	json.NewEncoder(w).Encode(user)
}
