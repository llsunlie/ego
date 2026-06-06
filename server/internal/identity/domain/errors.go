package domain

import "errors"

var ErrUserNotFound = errors.New("user not found")

var ErrInvalidPassword = errors.New("invalid password")

var ErrPhoneAlreadyRegistered = errors.New("phone already registered")

var ErrInvalidVerificationCode = errors.New("invalid verification code")

var ErrCodeExpired = errors.New("verification code expired")

var ErrInvalidPhone = errors.New("invalid phone number")
