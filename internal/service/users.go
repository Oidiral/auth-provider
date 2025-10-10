package service

import (
	"context"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/internal/repository"
)

type UserService struct {
	repository repository.Users
}

func NewUserService(repository repository.Users) *UserService {
	return &UserService{
		repository: repository,
	}
}

func (u *UserService) SignUp(ctx context.Context, input UserSignUpInput) error {

	user := domain.User{
		Username:  input.Username,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Phone:     input.Phone,
		Email:     input.Email,
	}
}

func (u *UserService) SignIn(ctx context.Context, input UserSignInInput) (Tokens, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserService) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	//TODO implement me
	panic("implement me")
}

func (u *UserService) Verify(ctx context.Context, userId string, hash string) error {
	//TODO implement me
	panic("implement me")
}
