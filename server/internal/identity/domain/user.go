package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string
	Phone        string
	PasswordHash string
	CreatedAt    time.Time
}

type UserRepository interface {
	FindByPhone(ctx context.Context, phone string) (*User, error)
	Create(ctx context.Context, user *User) error
}
