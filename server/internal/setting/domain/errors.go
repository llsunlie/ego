package domain

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrFeedbackEmpty  = errors.New("feedback content is empty")
)
