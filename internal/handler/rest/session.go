package rest

import (
	"github.com/KimNattanan/go-user-service/internal/usecase"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type HttpSessionHandler struct {
	sessionUsecase    usecase.SessionUsecase
	userUsecase       usecase.UserUsecase
	sessionStore      sessions.Store
	googleOauthConfig *oauth2.Config
}

func NewHttpSessionHandler(sessionUsecase usecase.SessionUsecase, userUsecase usecase.UserUsecase, sessionStore sessions.Store, googleOauthConfig *oauth2.Config) *HttpSessionHandler {
	return &HttpSessionHandler{
		sessionUsecase:    sessionUsecase,
		userUsecase:       userUsecase,
		sessionStore:      sessionStore,
		googleOauthConfig: googleOauthConfig,
	}
}

// func (h *HttpSessionHandler) RenewToken(w http.ResponseWriter, r *http.Request) {
// 	cookieSession, _ := h.sessionStore.Get(r, "cookie-session")
// 	sessionID, _ := cookieSession.Values["session_id"].(string)
// 	session, err := h.sessionUsecase.FindByID(r.Context(), sessionID)
// 	if err != nil {
// 		cookieSession.Options.MaxAge = -1
// 		cookieSession.Save(r, w)
// 		http.Error(w, apperror.ErrInternalServer.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	if session.IsRevoked {
// 		cookieSession.Options.MaxAge = -1
// 		cookieSession.Save(r, w)
// 		http.Error(w, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
// 		return
// 	}
// 	if user, err := h.userUsecase.FindByID(r.Context(), session.UserID); err != nil || user == nil {
// 		cookieSession.Options.MaxAge = -1
// 		cookieSession.Save(r, w)
// 		http.Error(w, apperror.ErrUnauthorized.Error(), http.StatusUnauthorized)
// 		return
// 	}
// 	token := &oauth2.Token{
// 		RefreshToken: session.RefreshToken,
// 	}
// 	tokenSource := h.googleOauthConfig.TokenSource(r.Context(), token)
// 	newToken, err := tokenSource.Token()
// 	if err != nil {

// 	}
// }
