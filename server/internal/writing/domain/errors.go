package domain

import "errors"

var (
	ErrTraceNotFound  = errors.New("trace not found")
	ErrMomentNotFound = errors.New("moment not found")
	ErrEmptyContent   = errors.New("moment content must not be empty")
)
