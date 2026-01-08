package dto

import "github.com/KimNattanan/go-user-service/internal/entity"

type UserResponse struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PictureURL string `json:"picture_url"`
	Preference *PreferenceResponse
}

type UserUpdateRequest struct {
	Name       string `json:"name,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	PictureURL string `json:"picture_url,omitempty"`
}

type RegisterRequest struct {
	Email      string `json:"email" valid:"required,email"`
	Password   string `json:"password" valid:"required"`
	Name       string `json:"name"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PictureURL string `json:"picture_url" valid:"url"`
}

type LoginRequest struct {
	Email    string `json:"email" valid:"required,email"`
	Password string `json:"password" valid:"required"`
}

func ToUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		Email:      user.Email,
		Name:       user.Name,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		PictureURL: user.PictureURL,
		Preference: ToPreferenceResponse(&user.Preference),
	}
}

func ToUserResponseList(users []*entity.User) []*UserResponse {
	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = ToUserResponse(user)
	}
	return userResponses
}
