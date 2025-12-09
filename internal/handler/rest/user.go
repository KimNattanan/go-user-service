package rest

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/KimNattanan/go-user-service/internal/dto"
	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/usecase"
	"github.com/KimNattanan/go-user-service/pkg/apperror"
	"github.com/KimNattanan/go-user-service/pkg/token"
	"github.com/gorilla/mux"
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

// @Summary Redirect to Google OAuth login
// @Tags Auth
// @Description Redirects user to Google OAuth provider
// @Success 302
// @Router /auth/google/login [get]
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

// @Summary OAuth callback from Google
// @Tags Auth
// @Description Handles Google OAuth callback and creates a session
// @Success 303
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Router /auth/google/callback [get]
func (h *HttpUserHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
	token, err := h.googleOauthConfig.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "failed to exchange token", http.StatusUnauthorized)
		return
	}

	client := h.googleOauthConfig.Client(ctx, token)
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

	user, err := h.userUsecase.LoginOrRegisterWithGoogle(ctx, userInfo)
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
	if err := h.sessionUsecase.Create(ctx, session); err != nil {
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

// @Summary Logout user
// @Tags Auth
// @Success 204
// @Router /auth/logout [post]
func (h *HttpUserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	cookieSession, err := h.sessionStore.Get(r, "session")
	if err != nil {
		http.Error(w, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	refreshToken, _ := cookieSession.Values["refresh_token"].(string)
	refreshClaims, err := h.jwtMaker.VerfiyToken(refreshToken)
	if err != nil {
		http.Error(w, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.sessionUsecase.Revoke(ctx, refreshClaims.RegisteredClaims.ID); err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}
	cookieSession.Values["access_token"] = ""
	cookieSession.Save(r, w)

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Delete current user
// @Tags Me
// @Success 204
// @Router /me [delete]
func (h *HttpUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	userID, _ := ctx.Value("userID").(string)

	if err := h.userUsecase.Delete(ctx, userID); err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Get public user profile by ID
// @Tags Users
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 404 {string} string
// @Router /users/{id} [get]
func (h *HttpUserHandler) FindUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := h.userUsecase.FindByID(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(dto.ToUserResponse(user))
}

// @Summary Get current user
// @Tags Me
// @Success 200 {object} dto.UserResponse
// @Router /me [get]
func (h *HttpUserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	userID, _ := ctx.Value("userID").(string)

	user, err := h.userUsecase.FindByID(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(dto.ToUserResponse(user))
}

// @Summary Update current user
// @Tags Me
// @Param request body dto.UserUpdateRequest true "Update user payload"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {string} string
// @Router /me [patch]
func (h *HttpUserHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	userID, _ := ctx.Value("userID").(string)

	var (
		data0 dto.UserUpdateRequest
		data  map[string]interface{}
	)
	if err := json.NewDecoder(r.Body).Decode(&data0); err != nil {
		http.Error(w, apperror.ErrInvalidData.Error(), http.StatusBadRequest)
		return
	}
	dataBytes, err := json.Marshal(data0)
	if err != nil {
		http.Error(w, apperror.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		http.Error(w, apperror.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	user, err := h.userUsecase.Update(ctx, userID, data)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(dto.ToUserResponse(user))
}
