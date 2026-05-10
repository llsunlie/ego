package id

import "github.com/google/uuid"

type UUIDGenerator struct{}

func NewUUIDGenerator() UUIDGenerator {
	return UUIDGenerator{}
}

func (UUIDGenerator) New() string {
	return uuid.New().String()
}
