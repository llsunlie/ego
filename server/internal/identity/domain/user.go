package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string
	Account      string
	PasswordHash string
	CreatedAt    time.Time
}

type UserRepository interface {
	FindByAccount(ctx context.Context, account string) (*User, error)
	Create(ctx context.Context, user *User) error
}
