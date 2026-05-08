package domain

import "errors"

var (
	ErrStarNotFound        = errors.New("star not found")
	ErrChatSessionNotFound = errors.New("chat session not found")
	ErrWrongUser           = errors.New("resource does not belong to current user")
)
