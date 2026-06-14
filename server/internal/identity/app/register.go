package app

import (
	"context"
	"errors"
	"fmt"

	"ego-server/internal/identity/domain"
)

type RegisterUseCase struct {
	users     domain.UserRepository
	hasher    PasswordHasher
	tokens    TokenIssuer
	ids       IDGenerator
	smsSender SmsService
}

func NewRegisterUseCase(
	users domain.UserRepository,
	hasher PasswordHasher,
	tokens TokenIssuer,
	ids IDGenerator,
	smsSender SmsService,
) *RegisterUseCase {
	return &RegisterUseCase{
		users:     users,
		hasher:    hasher,
		tokens:    tokens,
		ids:       ids,
		smsSender: smsSender,
	}
}

type RegisterResult struct {
	AccessToken  string
	RefreshToken string
}

func (uc *RegisterUseCase) Register(ctx context.Context, phone, code, password string) (*RegisterResult, error) {
	if !phonePattern.MatchString(phone) {
		return nil, domain.ErrInvalidPhone
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password too short")
	}

	ok, err := uc.smsSender.Verify(ctx, phone, code)
	if err != nil {
		return nil, fmt.Errorf("verify code: %w", err)
	}
	if !ok {
		return nil, domain.ErrInvalidVerificationCode
	}

	_, err = uc.users.FindByPhone(ctx, phone)
	if err == nil {
		return nil, domain.ErrPhoneAlreadyRegistered
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		ID:           uc.ids.New(),
		Phone:        phone,
		PasswordHash: hash,
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	accessToken, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err := uc.tokens.IssueRefresh(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	return &RegisterResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
