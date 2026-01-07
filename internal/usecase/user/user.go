package user

import (
	"context"
	"errors"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/repo"
	"github.com/KimNattanan/go-user-service/pkg/apperror"
)

type UserUsecase struct {
	repo repo.UserRepo
}

func NewUserUsecase(repo repo.UserRepo) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (u *UserUsecase) Create(ctx context.Context, user *entity.User) error {
	return u.repo.Create(ctx, user)
}

func (u *UserUsecase) FindAll(ctx context.Context) ([]*entity.User, error) {
	return u.repo.FindAll(ctx)
}

func (u *UserUsecase) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *UserUsecase) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return u.repo.FindByEmail(ctx, email)
}

// fields should be json name
func (u *UserUsecase) Update(ctx context.Context, id string, fields map[string]interface{}) (*entity.User, error) {
	return u.repo.Update(ctx, id, fields)
}

func (u *UserUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *UserUsecase) LoginOrRegisterWithGoogle(ctx context.Context, userInfo map[string]interface{}) (*entity.User, error) {
	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		return nil, apperror.ErrInvalidData
	}
	name, _ := userInfo["name"].(string)
	firstName, _ := userInfo["given_name"].(string)
	lastName, _ := userInfo["family_name"].(string)
	pictureURL, _ := userInfo["picture"].(string)

	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, apperror.ErrRecordNotFound) {
		return nil, err
	}
	if user == nil {
		user = &entity.User{
			Email:        email,
			Name:         name,
			FirstName:    firstName,
			LastName:     lastName,
			PasswordHash: "",
			PictureURL:   pictureURL,
		}
		if err := u.repo.Create(ctx, user); err != nil {
			return nil, err
		}
	} else {
		if user, err = u.repo.Update(ctx, user.ID, map[string]interface{}{
			"first_name":  firstName,
			"last_name":   lastName,
			"picture_url": pictureURL,
		}); err != nil {
			return nil, err
		}
	}
	return user, nil
}
