package bootstrap

import "github.com/google/uuid"

type uuidGenerator struct{}

func (uuidGenerator) New() string {
	return uuid.New().String()
}
