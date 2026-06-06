package app

import (
	"context"
	"regexp"
	"strings"

	"ego-server/internal/identity/domain"
)

type SendCodeUseCase struct {
	smsSender SmsService
}

func NewSendCodeUseCase(smsSender SmsService) *SendCodeUseCase {
	return &SendCodeUseCase{smsSender: smsSender}
}

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

func (uc *SendCodeUseCase) SendCode(ctx context.Context, phone string) error {
	phone = strings.TrimSpace(phone)
	if !phonePattern.MatchString(phone) {
		return domain.ErrInvalidPhone
	}

	return uc.smsSender.Send(ctx, phone)
}
