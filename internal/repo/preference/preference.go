package preference

import (
	"context"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"gorm.io/gorm"
)

type PreferenceRepo struct {
	db *gorm.DB
}

func NewPreferenceRepo(db *gorm.DB) *PreferenceRepo {
	return &PreferenceRepo{db: db}
}

func (r *PreferenceRepo) FindByUserID(ctx context.Context, userID string) (*entity.Preference, error) {
	db := r.db.WithContext(ctx)
	var preference entity.Preference
	if err := db.First(&preference, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &preference, nil
}

func (r *PreferenceRepo) Update(ctx context.Context, userID string, fields map[string]interface{}) error {
	db := r.db.WithContext(ctx)
	result := db.Model(&entity.Preference{}).Where("user_id = ?", userID).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
