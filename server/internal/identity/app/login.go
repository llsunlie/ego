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
	ids    IDGenerator
}

func NewLoginUseCase(users domain.UserRepository, hasher PasswordHasher, tokens TokenIssuer, ids IDGenerator) *LoginUseCase {
	return &LoginUseCase{users: users, hasher: hasher, tokens: tokens, ids: ids}
}

type LoginResult struct {
	Token   string
	Created bool
}

func (uc *LoginUseCase) Login(ctx context.Context, account, password string) (*LoginResult, error) {
	user, err := uc.users.FindByAccount(ctx, account)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("find user by account: %w", err)
	}

	if errors.Is(err, domain.ErrUserNotFound) {
		return uc.register(ctx, account, password)
	}

	if err := uc.hasher.Verify(user.PasswordHash, password); err != nil {
		return nil, domain.ErrInvalidPassword
	}

	token, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	return &LoginResult{Token: token, Created: false}, nil
}

func (uc *LoginUseCase) register(ctx context.Context, account, password string) (*LoginResult, error) {
	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		ID:           uc.ids.New(),
		Account:      account,
		PasswordHash: hash,
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	return &LoginResult{Token: token, Created: true}, nil
}
