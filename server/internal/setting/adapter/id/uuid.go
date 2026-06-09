package id

import "github.com/google/uuid"

// UUIDGenerator generates UUID v4 strings.
type UUIDGenerator struct{}

func NewUUIDGenerator() UUIDGenerator {
	return UUIDGenerator{}
}

func (UUIDGenerator) New() string {
	return uuid.New().String()
}
