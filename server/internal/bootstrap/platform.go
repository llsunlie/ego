package bootstrap

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"ego-server/internal/config"
	"ego-server/internal/platform/auth"
	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Platform struct {
	Pool   *pgxpool.Pool
	JWTKey []byte
	JWTExp time.Duration
	Hasher auth.BcryptHasher
	Tokens auth.JWTIssuer
	Logger *slog.Logger
}

func InitPlatform(cfg *config.Config) (*Platform, error) {
	logger, err := logging.New(logging.Config{
		Level:      cfg.LogLevel,
		Format:     cfg.LogFormat,
		OutputPath: cfg.LogOutput,
	})
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

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
		Logger: logger,
	}, nil
}

func (p *Platform) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
