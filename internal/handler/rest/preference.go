package rest

import (
	"encoding/json"
	"net/http"

	"github.com/KimNattanan/go-user-service/internal/dto"
	"github.com/KimNattanan/go-user-service/internal/usecase"
	"github.com/KimNattanan/go-user-service/pkg/apperror"
)

type HttpPreferenceHandler struct {
	preferenceUsecase usecase.PreferenceUsecase
}

func NewHttpPreferenceHandler(preferenceUsecase usecase.PreferenceUsecase) *HttpPreferenceHandler {
	return &HttpPreferenceHandler{preferenceUsecase: preferenceUsecase}
}

func (h *HttpPreferenceHandler) GetPreference(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	userID, _ := ctx.Value("userID").(string)

	preference, err := h.preferenceUsecase.FindByUserID(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(preference)
}

func (h *HttpPreferenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	userID, _ := ctx.Value("userID").(string)

	var (
		data0 dto.PreferenceUpdateRequest
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

	preference, err := h.preferenceUsecase.Update(ctx, userID, data)
	if err != nil {
		http.Error(w, err.Error(), apperror.StatusCode(err))
		return
	}

	json.NewEncoder(w).Encode(dto.ToPreferenceResponse(preference))
}
