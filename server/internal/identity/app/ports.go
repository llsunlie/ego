package app

import "context"

type PasswordHasher interface {
	Hash(plaintext string) (string, error)
	Verify(hash, plaintext string) error
}

type TokenIssuer interface {
	Issue(userID string) (string, error)
	IssueRefresh(userID string) (string, error)
}

type RefreshTokenVerifier interface {
	Verify(tokenStr, expectedType string) (userID string, err error)
}

type IDGenerator interface {
	New() string
}

type SmsService interface {
	Send(ctx context.Context, phone string) error
	Verify(ctx context.Context, phone, code string) (bool, error)
}
