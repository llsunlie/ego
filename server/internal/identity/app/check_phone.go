package app

import (
	"context"
	"errors"
	"strings"

	"ego-server/internal/identity/domain"
)

type CheckPhoneUseCase struct {
	users domain.UserRepository
}

func NewCheckPhoneUseCase(users domain.UserRepository) *CheckPhoneUseCase {
	return &CheckPhoneUseCase{users: users}
}

type CheckPhoneResult struct {
	Registered bool
}

func (uc *CheckPhoneUseCase) Check(ctx context.Context, phone string) (*CheckPhoneResult, error) {
	phone = strings.TrimSpace(phone)
	if !phonePattern.MatchString(phone) {
		return nil, domain.ErrInvalidPhone
	}

	_, err := uc.users.FindByPhone(ctx, phone)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	return &CheckPhoneResult{Registered: err == nil}, nil
}
