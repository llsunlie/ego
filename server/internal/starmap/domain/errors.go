package domain

import "errors"

var (
	ErrTraceNotFound       = errors.New("trace not found")
	ErrTraceAlreadyStashed = errors.New("trace already stashed")
	ErrStarNotFound        = errors.New("star not found")
	ErrConstellationNotFound = errors.New("constellation not found")
)
