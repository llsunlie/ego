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
	Pool                    *pgxpool.Pool
	JWTKey                  []byte
	JWTExp                  time.Duration
	Hasher                  auth.BcryptHasher
	Tokens                  auth.JWTIssuer
	Logger                  *slog.Logger
	AIClient                *ai.Client
	SmsService              *sms.AliyunSmsService
	TokenVerifier           auth.RefreshTokenVerifier
	AIEmbeddingDim          int
	EchoRecallTopK          int32
	EchoSparseTopK          int32
	EchoHybridRRFK          int
	EchoSparseOn            bool
	ConstellationSparseTopK int
	ConstellationHybridRRFK int
	ConstellationSparseOn   bool
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

	embeddingDim, err := strconv.Atoi(cfg.AIEmbeddingDim)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid AI_EMBEDDING_DIM: %w", err)
	}
	echoRecallTopK, err := strconv.Atoi(cfg.EchoRecallTopK)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid ECHO_RECALL_TOP_K: %w", err)
	}
	echoSparseTopK, err := strconv.Atoi(cfg.EchoSparseTopK)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid ECHO_SPARSE_RECALL_TOP_K: %w", err)
	}
	echoHybridRRFK, err := strconv.Atoi(cfg.EchoHybridRRFK)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid ECHO_HYBRID_RRF_K: %w", err)
	}
	echoSparseOn, err := strconv.ParseBool(cfg.EchoSparseEnabled)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid ECHO_SPARSE_RECALL_ENABLED: %w", err)
	}
	constellationSparseTopK, err := strconv.Atoi(cfg.ConstellationSparseTopK)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid CONSTELLATION_SPARSE_RECALL_TOP_K: %w", err)
	}
	constellationHybridRRFK, err := strconv.Atoi(cfg.ConstellationHybridRRFK)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid CONSTELLATION_HYBRID_RRF_K: %w", err)
	}
	constellationSparseOn, err := strconv.ParseBool(cfg.ConstellationSparseEnabled)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("invalid CONSTELLATION_SPARSE_RECALL_ENABLED: %w", err)
	}

	aiClient := ai.NewClient(ai.Config{
		EmbeddingAPIKey:  cfg.AIEmbeddingAPIKey,
		EmbeddingBaseURL: cfg.AIEmbeddingBaseURL,
		EmbeddingModel:   cfg.AIEmbeddingModel,
		ChatAPIKey:       cfg.AIChatAPIKey,
		ChatBaseURL:      cfg.AIChatBaseURL,
		ChatModel:        cfg.AIChatModel,
	}, logger)
	logger.Debug("platform config parsed",
		"ai_embedding_model", cfg.AIEmbeddingModel,
		"ai_embedding_dim", embeddingDim,
		"echo_recall_top_k", echoRecallTopK,
		"echo_sparse_enabled", echoSparseOn,
		"echo_sparse_top_k", echoSparseTopK,
		"echo_hybrid_rrf_k", echoHybridRRFK,
		"constellation_sparse_enabled", constellationSparseOn,
		"constellation_sparse_top_k", constellationSparseTopK,
		"constellation_hybrid_rrf_k", constellationHybridRRFK,
	)

	return &Platform{
		Pool:                    pool,
		JWTKey:                  jwtKey,
		JWTExp:                  accessExp,
		Hasher:                  auth.BcryptHasher{},
		Tokens:                  auth.JWTIssuer{Secret: jwtKey, Exp: accessExp, RefreshExp: refreshExp},
		Logger:                  logger,
		AIClient:                aiClient,
		SmsService:              newSmsService(cfg, pool),
		TokenVerifier:           auth.RefreshTokenVerifier{Secret: jwtKey},
		AIEmbeddingDim:          embeddingDim,
		EchoRecallTopK:          int32(echoRecallTopK),
		EchoSparseTopK:          int32(echoSparseTopK),
		EchoHybridRRFK:          echoHybridRRFK,
		EchoSparseOn:            echoSparseOn,
		ConstellationSparseTopK: constellationSparseTopK,
		ConstellationHybridRRFK: constellationHybridRRFK,
		ConstellationSparseOn:   constellationSparseOn,
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
