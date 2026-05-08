package domain

import "errors"

var (
	ErrTraceNotFound   = errors.New("trace not found")
	ErrMomentNotFound  = errors.New("moment not found")
	ErrEchoNotFound    = errors.New("echo not found")
	ErrInsightNotFound = errors.New("insight not found")
	ErrEmptyContent    = errors.New("moment content must not be empty")
)
