package session

import (
	"context"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/repo"
)

type SessionUsecase struct {
	repo repo.SessionRepo
}

func NewSessionUsecase(repo repo.SessionRepo) *SessionUsecase {
	return &SessionUsecase{repo: repo}
}

func (u *SessionUsecase) Create(ctx context.Context, session *entity.Session) error {
	return u.repo.Create(ctx, session)
}

func (u *SessionUsecase) FindByID(ctx context.Context, id string) (*entity.Session, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *SessionUsecase) FindByUserID(ctx context.Context, userID string) ([]*entity.Session, error) {
	return u.repo.FindByUserID(ctx, userID)
}

func (u *SessionUsecase) Revoke(ctx context.Context, id string) error {
	return u.repo.Revoke(ctx, id)
}

func (u *SessionUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}
