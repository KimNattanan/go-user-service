package repo

import (
	"context"

	"github.com/KimNattanan/go-user-service/internal/entity"
)

type (
	UserRepo interface {
		Save(ctx context.Context, u *entity.User) error
		FindByID(ctx context.Context, id string) (*entity.User, error)
		FindByEmail(ctx context.Context, email string) (*entity.User, error)
		Update(ctx context.Context, u *entity.User) error
		Delete(ctx context.Context, id string) error
	}
	PreferenceRepo interface {
		Save(ctx context.Context, p *entity.Preference) error
		FindByUserID(ctx context.Context, userID string) (*entity.Preference, error)
		Update(ctx context.Context, p *entity.Preference) error
	}
	SessionRepo interface {
		Save(ctx context.Context, s *entity.Session) error
		FindByID(ctx context.Context, id string) (*entity.Session, error)
		FindByUserID(ctx context.Context, userID string) ([]*entity.Session, error)
		Revoke(ctx context.Context, id string) error
		Delete(ctx context.Context, id string) error
	}
)
