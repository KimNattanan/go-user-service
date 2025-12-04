package repo

import (
	"context"

	"github.com/KimNattanan/go-user-service/internal/entity"
)

type (
	UserRepo interface {
		Create(ctx context.Context, user *entity.User) error
		FindByID(ctx context.Context, id string) (*entity.User, error)
		FindByEmail(ctx context.Context, email string) (*entity.User, error)
		Update(ctx context.Context, id string, fields map[string]interface{}) (*entity.User, error)
		Delete(ctx context.Context, id string) error
	}
	PreferenceRepo interface {
		FindByUserID(ctx context.Context, userID string) (*entity.Preference, error)
		Update(ctx context.Context, userID string, fields map[string]interface{}) (*entity.Preference, error)
	}
	SessionRepo interface {
		Create(ctx context.Context, session *entity.Session) error
		FindByID(ctx context.Context, id string) (*entity.Session, error)
		FindByUserID(ctx context.Context, userID string) ([]*entity.Session, error)
		Revoke(ctx context.Context, id string) error
		Delete(ctx context.Context, id string) error
	}
)
