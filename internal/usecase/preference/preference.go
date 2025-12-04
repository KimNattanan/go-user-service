package preference

import (
	"context"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/repo"
)

type PreferenceUsecase struct {
	repo repo.PreferenceRepo
}

func NewPreferenceUsecase(repo repo.PreferenceRepo) *PreferenceUsecase {
	return &PreferenceUsecase{repo: repo}
}

func (u *PreferenceUsecase) FindByUserID(ctx context.Context, userID string) (*entity.Preference, error) {
	return u.repo.FindByUserID(ctx, userID)
}

func (u *PreferenceUsecase) Update(ctx context.Context, userID string, fields map[string]interface{}) (*entity.Preference, error) {
	return u.repo.Update(ctx, userID, fields)
}
