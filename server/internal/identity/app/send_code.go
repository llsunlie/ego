package app

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"ego-server/internal/identity/domain"
)

type SendCodeUseCase struct {
	users     domain.UserRepository
	smsSender SmsService
}

func NewSendCodeUseCase(users domain.UserRepository, smsSender SmsService) *SendCodeUseCase {
	return &SendCodeUseCase{users: users, smsSender: smsSender}
}

type SendCodeResult struct {
	Registered bool
}

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

func (uc *SendCodeUseCase) SendCode(ctx context.Context, phone string) (*SendCodeResult, error) {
	phone = strings.TrimSpace(phone)
	if !phonePattern.MatchString(phone) {
		return nil, domain.ErrInvalidPhone
	}

	_, err := uc.users.FindByPhone(ctx, phone)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}
	registered := err == nil

	// Only send SMS for new phones (registration). Registered users use password login.
	if !registered {
		if err := uc.smsSender.Send(ctx, phone); err != nil {
			return nil, err
		}
	}

	return &SendCodeResult{Registered: registered}, nil
}
