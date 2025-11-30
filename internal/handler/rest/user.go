package rest

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/usecase"
	"github.com/KimNattanan/go-user-service/pkg/apperror"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type HttpUserHandler struct {
	userUsecase       usecase.UserUsecase
	sessionUsecase    usecase.SessionUsecase
	googleOauthConfig *oauth2.Config
}

func NewHttpUserHandler(userUsecase usecase.UserUsecase) *HttpUserHandler {
	return &HttpUserHandler{
		userUsecase: userUsecase,
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

	session := &entity.Session{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		RefreshToken: token.RefreshToken,
		IsRevoked:    false,
		ExpiresAt:    token.Expiry,
	}
	if err := h.sessionUsecase.Create(r.Context(), session); err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Expires:  time.Now(),
		HttpOnly: true,
		Secure:   false,
	})

}
