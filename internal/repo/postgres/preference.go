package postgres

import "gorm.io/gorm"

type PreferenceRepo struct {
	db *gorm.DB
}

func NewPreferenceRepo(db *gorm.DB) *PreferenceRepo {
	return &PreferenceRepo{db: db}
}
