package user

import (
	"context"
	"errors"

	"github.com/KimNattanan/go-user-service/internal/entity"
	"github.com/KimNattanan/go-user-service/internal/repo"
	"github.com/KimNattanan/go-user-service/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repo repo.UserRepo
}

func NewUserUsecase(repo repo.UserRepo) *UserUsecase {
	return &UserUsecase{repo: repo}
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
			Email:      email,
			Name:       name,
			FirstName:  firstName,
			LastName:   lastName,
			Password:   "",
			PictureURL: pictureURL,
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

func (u *UserUsecase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	existingUser, err := u.repo.FindByEmail(ctx, user.Email)
	if existingUser != nil {
		return nil, apperror.ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, apperror.ErrRecordNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(passwordHash)

	if err := u.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	createdUser, err := u.repo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return createdUser, nil
}

func (u *UserUsecase) Login(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}
	return user, nil
}
