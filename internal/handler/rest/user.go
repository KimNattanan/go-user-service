package rest

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/KimNattanan/go-user-service/internal/dto"
	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/usecase"
	"github.com/KimNattanan/go-user-service/pkg/apperror"
	"github.com/KimNattanan/go-user-service/pkg/token"
	"github.com/asaskevich/govalidator"
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
	jwtExpiration     time.Duration
}

func NewHttpUserHandler(userUsecase usecase.UserUsecase, sessionUsecase usecase.SessionUsecase, sessionStore sessions.Store, googleOauthConfig *oauth2.Config, jwtMaker *token.JWTMaker, jwtExpiration int) *HttpUserHandler {
	return &HttpUserHandler{
		userUsecase:       userUsecase,
		sessionUsecase:    sessionUsecase,
		sessionStore:      sessionStore,
		googleOauthConfig: googleOauthConfig,
		jwtMaker:          jwtMaker,
		jwtExpiration:     time.Duration(jwtExpiration),
	}
}

// @Summary Redirect to Google OAuth login
// @Description Redirects user to Google OAuth provider
// @Tags Auth
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
// @Description Handles Google OAuth callback and creates a session
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]interface{} "logged in successfully"
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Router /auth/google/callback [get]
func (h *HttpUserHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "id_token missing", http.StatusUnauthorized)
		return
	}

	payload, err := idtoken.Validate(ctx, rawIDToken, h.googleOauthConfig.ClientID)
	if err != nil {
		http.Error(w, "invalid id token", http.StatusUnauthorized)
		return
	}

	userInfo := map[string]interface{}{
		"sub":            payload.Subject,
		"email":          payload.Claims["email"],
		"email_verified": payload.Claims["email_verified"],
		"name":           payload.Claims["name"],
		"given_name":     payload.Claims["given_name"],
		"family_name":    payload.Claims["family_name"],
		"picture":        payload.Claims["picture"],
	}
	if userInfo["email_verified"] != true {
		http.Error(w, "email not verified", http.StatusUnauthorized)
		return
	}

	user, err := h.userUsecase.LoginOrRegisterWithGoogle(ctx, userInfo)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	refreshToken, refreshClaims, err := h.jwtMaker.CreateToken(user.ID, time.Second*h.jwtExpiration)
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
		Secure:   true,
	})
	cookieSession, _ := h.sessionStore.Get(r, "session")
	cookieSession.Values["access_token"] = accessToken
	cookieSession.Values["refresh_token"] = refreshToken
	if err := cookieSession.Save(r, w); err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"message": "logged in successfully"})
}

// @Summary Register new user
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]interface{} "registered successfully"
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Router /auth/register [post]
func (h *HttpUserHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	req := new(dto.RegisterRequest)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, apperror.ErrInvalidData.Error(), http.StatusBadRequest)
		return
	}
	ok, errr := govalidator.ValidateStruct(req)
	fmt.Println(ok, errr, "!!")
	if !ok {
		http.Error(w, errr.Error(), http.StatusBadRequest)
		return
	}

	user := &entity.User{
		Email:      req.Email,
		Password:   req.Password,
		Name:       req.Name,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		PictureURL: req.PictureURL,
	}

	user, err := h.userUsecase.Register(ctx, user)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	refreshToken, refreshClaims, err := h.jwtMaker.CreateToken(user.ID, time.Second*h.jwtExpiration)
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
		GoogleRefreshToken: "",
		IsRevoked:          false,
		CreatedAt:          time.Now(),
		ExpiresAt:          refreshClaims.RegisteredClaims.ExpiresAt.Time,
	}
	if err := h.sessionUsecase.Create(ctx, session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookieSession, _ := h.sessionStore.Get(r, "session")
	cookieSession.Values["access_token"] = accessToken
	cookieSession.Values["refresh_token"] = refreshToken
	if err := cookieSession.Save(r, w); err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"message": "registered successfully"})
}

// @Summary Login user
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]interface{} "logged in successfully"
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Router /auth/login [post]
func (h *HttpUserHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	req := new(dto.LoginRequest)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, apperror.ErrInvalidData.Error(), http.StatusBadRequest)
		return
	}
	if ok, err := govalidator.ValidateStruct(req); !ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	refreshToken, refreshClaims, err := h.jwtMaker.CreateToken(user.ID, time.Second*h.jwtExpiration)
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
		GoogleRefreshToken: "",
		IsRevoked:          false,
		CreatedAt:          time.Now(),
		ExpiresAt:          refreshClaims.RegisteredClaims.ExpiresAt.Time,
	}
	if err := h.sessionUsecase.Create(ctx, session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cookieSession, _ := h.sessionStore.Get(r, "session")
	cookieSession.Values["access_token"] = accessToken
	cookieSession.Values["refresh_token"] = refreshToken
	if err := cookieSession.Save(r, w); err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"message": "logged in successfully"})
}

// @Summary Logout user
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]interface{} "logged out successfully"
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

	json.NewEncoder(w).Encode(map[string]interface{}{"message": "logged out successfully"})
}

// @Summary Delete current user
// @Tags Me
// @Produce json
// @Success 200 {object} map[string]interface{} "user deleted"
// @Router /me [delete]
func (h *HttpUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	userID, _ := ctx.Value("userID").(string)

	if err := h.userUsecase.Delete(ctx, userID); err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"message": "user deleted"})
}

// @Summary Get user by ID
// @Tags Users
// @Accept json
// @Produce json
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

// @Summary Get all users
// @Tags Users
// @Produce json
// @Success 200 {array} dto.UserResponse
// @Router /users [get]
func (h *HttpUserHandler) FindAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	users, err := h.userUsecase.FindAll(ctx)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(dto.ToUserResponseList(users))
}

// @Summary Get current user
// @Tags Me
// @Produce json
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
// @Accept json
// @Produce json
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
