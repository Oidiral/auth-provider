package service

import "context"

type UserSignUpInput struct {
	Email     string
	Phone     string
	Username  string
	FirstName string
	LastName  string
	Password  string
}

type UserSignInInput struct {
	Email    string
	Password string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Users interface {
	SignUp(ctx context.Context, input UserSignUpInput) error
	SignIn(ctx context.Context, input UserSignInInput) (Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (Tokens, error)
	Verify(ctx context.Context, userId string, hash string) error
}
