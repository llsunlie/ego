package app

import (
	"context"
	"errors"
	"fmt"

	"ego-server/internal/identity/domain"
)

type LoginUseCase struct {
	users  domain.UserRepository
	hasher PasswordHasher
	tokens TokenIssuer
}

func NewLoginUseCase(users domain.UserRepository, hasher PasswordHasher, tokens TokenIssuer) *LoginUseCase {
	return &LoginUseCase{users: users, hasher: hasher, tokens: tokens}
}

type LoginResult struct {
	Token string
}

func (uc *LoginUseCase) Login(ctx context.Context, phone, password string) (*LoginResult, error) {
	user, err := uc.users.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	if err := uc.hasher.Verify(user.PasswordHash, password); err != nil {
		return nil, domain.ErrInvalidPassword
	}

	token, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	return &LoginResult{Token: token}, nil
}
