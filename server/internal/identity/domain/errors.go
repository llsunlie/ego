package domain

import "errors"

var ErrUserNotFound = errors.New("user not found")

var ErrInvalidPassword = errors.New("invalid password")
