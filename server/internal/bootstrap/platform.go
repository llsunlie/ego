package bootstrap

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"ego-server/internal/config"
	"ego-server/internal/platform/ai"
	"ego-server/internal/platform/auth"
	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Platform struct {
	Pool     *pgxpool.Pool
	JWTKey   []byte
	JWTExp   time.Duration
	Hasher   auth.BcryptHasher
	Tokens   auth.JWTIssuer
	Logger   *slog.Logger
	AIClient *ai.Client
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

	aiClient := ai.NewClient(ai.Config{
		APIKey:         cfg.AIAPIKey,
		BaseURL:        cfg.AIBaseURL,
		EmbeddingModel: cfg.AIEmbeddingModel,
		ChatModel:      cfg.AIChatModel,
	})

	return &Platform{
		Pool:     pool,
		JWTKey:   jwtKey,
		JWTExp:   jwtExp,
		Hasher:   auth.BcryptHasher{},
		Tokens:   auth.JWTIssuer{Secret: jwtKey, Exp: jwtExp},
		Logger:   logger,
		AIClient: aiClient,
	}, nil
}

func (p *Platform) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
