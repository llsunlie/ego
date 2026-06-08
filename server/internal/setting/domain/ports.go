package domain

import (
	"context"
	"time"
)

// UserInfo is the read-only view of a user for the setting module.
type UserInfo struct {
	Phone     string
	CreatedAt time.Time
}

// UserReader reads user data by ID.
type UserReader interface {
	FindByID(ctx context.Context, id string) (*UserInfo, error)
}
