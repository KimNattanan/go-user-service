package dto

import "github.com/KimNattanan/go-user-service/internal/entity"

type PreferenceResponse struct {
	Theme  string `json:"theme"`
}

type PreferenceUpdateRequest struct {
	Theme string `json:"theme"`
}

func ToPreferenceResponse(preference *entity.Preference) *PreferenceResponse {
	return &PreferenceResponse{
		Theme: preference.Theme,
	}
}