package app

import (
	"context"
	"errors"
	"fmt"

	"ego-server/internal/identity/domain"
)

type ResetPasswordUseCase struct {
	users     domain.UserRepository
	hasher    PasswordHasher
	tokens    TokenIssuer
	smsSender SmsService
}

func NewResetPasswordUseCase(
	users domain.UserRepository,
	hasher PasswordHasher,
	tokens TokenIssuer,
	smsSender SmsService,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		users:     users,
		hasher:    hasher,
		tokens:    tokens,
		smsSender: smsSender,
	}
}

type ResetPasswordResult struct {
	AccessToken  string
	RefreshToken string
}

func (uc *ResetPasswordUseCase) ResetPassword(ctx context.Context, phone, code, newPassword string) (*ResetPasswordResult, error) {
	if !phonePattern.MatchString(phone) {
		return nil, domain.ErrInvalidPhone
	}
	if len(newPassword) < 6 {
		return nil, fmt.Errorf("password too short")
	}

	ok, err := uc.smsSender.Verify(ctx, phone, code)
	if err != nil {
		return nil, fmt.Errorf("verify code: %w", err)
	}
	if !ok {
		return nil, domain.ErrInvalidVerificationCode
	}

	user, err := uc.users.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	hash, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := uc.users.UpdatePassword(ctx, user.ID, hash); err != nil {
		return nil, fmt.Errorf("update password: %w", err)
	}

	accessToken, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err := uc.tokens.IssueRefresh(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	return &ResetPasswordResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
