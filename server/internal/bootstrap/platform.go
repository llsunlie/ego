package bootstrap

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"ego-server/internal/config"
	"ego-server/internal/identity/adapter/sms"
	"ego-server/internal/platform/ai"
	"ego-server/internal/platform/auth"
	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Platform struct {
	Pool          *pgxpool.Pool
	JWTKey        []byte
	JWTExp        time.Duration
	Hasher        auth.BcryptHasher
	Tokens        auth.JWTIssuer
	Logger        *slog.Logger
	AIClient      *ai.Client
	SmsService    *sms.AliyunSmsService
	TokenVerifier auth.RefreshTokenVerifier
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

	accessExpHours, err := strconv.Atoi(cfg.JwtAccessExpHours)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXP_HOURS: %w", err)
	}
	accessExp := time.Duration(accessExpHours) * time.Hour

	refreshExpDays, err := strconv.Atoi(cfg.JwtRefreshExpDays)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXP_DAYS: %w", err)
	}
	refreshExp := time.Duration(refreshExpDays) * 24 * time.Hour

	aiClient := ai.NewClient(ai.Config{
		EmbeddingAPIKey:  cfg.AIEmbeddingAPIKey,
		EmbeddingBaseURL: cfg.AIEmbeddingBaseURL,
		EmbeddingModel:   cfg.AIEmbeddingModel,
		ChatAPIKey:       cfg.AIChatAPIKey,
		ChatBaseURL:      cfg.AIChatBaseURL,
		ChatModel:        cfg.AIChatModel,
	}, logger)

	return &Platform{
		Pool:       pool,
		JWTKey:     jwtKey,
		JWTExp:     accessExp,
		Hasher:     auth.BcryptHasher{},
		Tokens:     auth.JWTIssuer{Secret: jwtKey, Exp: accessExp, RefreshExp: refreshExp},
		Logger:     logger,
		AIClient:   aiClient,
		SmsService:    newSmsService(cfg, pool),
		TokenVerifier: auth.RefreshTokenVerifier{Secret: jwtKey},
	}, nil
}

func newSmsService(cfg *config.Config, pool *pgxpool.Pool) *sms.AliyunSmsService {
	s, err := sms.NewAliyunSmsService(
		cfg.AliyunAccessKeyID,
		cfg.AliyunAccessKeySecret,
		cfg.AliyunSmsSignName,
		cfg.AliyunSmsTemplateCode,
		cfg.AliyunSmsCodeLength,
		cfg.AliyunSmsValidTime,
		cfg.AliyunSmsInterval,
	)
	if err != nil {
		pool.Close()
		panic("init sms service: " + err.Error())
	}
	return s
}

func (p *Platform) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
