package bootstrap

import (
	"fmt"
	"strconv"
	"time"

	"ego-server/internal/config"
	"ego-server/internal/platform/auth"
	"ego-server/internal/platform/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Platform struct {
	Pool   *pgxpool.Pool
	JWTKey []byte
	JWTExp time.Duration
	Hasher auth.BcryptHasher
	Tokens auth.JWTIssuer
}

func InitPlatform(cfg *config.Config) (*Platform, error) {
	pool, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	jwtKey := []byte(cfg.JWTSecret)
	expHours, err := strconv.Atoi(cfg.JWTExpHours)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid JWT_EXP_HOURS: %w", err)
	}
	jwtExp := time.Duration(expHours) * time.Hour

	return &Platform{
		Pool:   pool,
		JWTKey: jwtKey,
		JWTExp: jwtExp,
		Hasher: auth.BcryptHasher{},
		Tokens: auth.JWTIssuer{Secret: jwtKey, Exp: jwtExp},
	}, nil
}

func (p *Platform) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
