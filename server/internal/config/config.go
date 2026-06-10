package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DatabaseURL                string
	JWTSecret                  string
	WebPort                    string
	WebTLSPort                 string
	GRPCPort                   string
	WebDir                     string
	JWTExpHours                string
	LogLevel                   string
	LogFormat                  string
	LogOutput                  string
	AIAPIKey                   string
	AIBaseURL                  string
	AIEmbeddingModel           string
	AIEmbeddingDim             string
	AIEmbeddingAPIKey          string
	AIEmbeddingBaseURL         string
	AIChatModel                string
	AIChatAPIKey               string
	AIChatBaseURL              string
	EchoRecallTopK             string
	ElasticsearchURL           string
	ElasticsearchUser          string
	ElasticsearchPass          string
	EchoSparseEnabled          string
	EchoSparseTopK             string
	EchoHybridRRFK             string
	ConstellationSparseEnabled string
	ConstellationSparseTopK    string
	ConstellationHybridRRFK    string
	AliyunAccessKeyID          string
	AliyunAccessKeySecret      string
	AliyunSmsSignName          string
	AliyunSmsTemplateCode      string
	AliyunSmsCodeLength        string
	AliyunSmsValidTime         string
	AliyunSmsInterval          string
	TLSDomain                  string
}

// getEnvWithFallback returns os.Getenv(key), or os.Getenv(fallback) if empty.
func getEnvWithFallback(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return os.Getenv(fallback)
}

func Load() *Config {
	// .env is the single source of configuration defaults.
	// Copy .env.example to .env and fill in your values.
	// OS environment variables take precedence over .env values.
	loadEnvFile()

	return &Config{
		DatabaseURL:                os.Getenv("DATABASE_URL"),
		JWTSecret:                  os.Getenv("JWT_SECRET"),
		WebPort:                    os.Getenv("WEB_PORT"),
		WebTLSPort:                 os.Getenv("WEB_TLS_PORT"),
		GRPCPort:                   os.Getenv("GRPC_PORT"),
		WebDir:                     os.Getenv("WEB_DIR"),
		JWTExpHours:                os.Getenv("JWT_EXP_HOURS"),
		LogLevel:                   os.Getenv("LOG_LEVEL"),
		LogFormat:                  os.Getenv("LOG_FORMAT"),
		LogOutput:                  os.Getenv("LOG_OUTPUT"),
		AIAPIKey:                   os.Getenv("AI_API_KEY"),
		AIBaseURL:                  os.Getenv("AI_BASE_URL"),
		AIEmbeddingModel:           os.Getenv("AI_EMBEDDING_MODEL"),
		AIEmbeddingDim:             getEnvDefault("AI_EMBEDDING_DIM", "4096"),
		AIEmbeddingAPIKey:          getEnvWithFallback("AI_EMBEDDING_API_KEY", "AI_API_KEY"),
		AIEmbeddingBaseURL:         getEnvWithFallback("AI_EMBEDDING_BASE_URL", "AI_BASE_URL"),
		AIChatModel:                os.Getenv("AI_CHAT_MODEL"),
		AIChatAPIKey:               getEnvWithFallback("AI_CHAT_API_KEY", "AI_API_KEY"),
		AIChatBaseURL:              getEnvWithFallback("AI_CHAT_BASE_URL", "AI_BASE_URL"),
		EchoRecallTopK:             getEnvDefault("ECHO_RECALL_TOP_K", "10"),
		ElasticsearchURL:           getEnvDefault("ELASTICSEARCH_URL", "http://localhost:9200"),
		ElasticsearchUser:          os.Getenv("ELASTICSEARCH_USERNAME"),
		ElasticsearchPass:          os.Getenv("ELASTICSEARCH_PASSWORD"),
		EchoSparseEnabled:          getEnvDefault("ECHO_SPARSE_RECALL_ENABLED", "true"),
		EchoSparseTopK:             getEnvDefault("ECHO_SPARSE_RECALL_TOP_K", "10"),
		EchoHybridRRFK:             getEnvDefault("ECHO_HYBRID_RRF_K", "60"),
		ConstellationSparseEnabled: getEnvDefault("CONSTELLATION_SPARSE_RECALL_ENABLED", "true"),
		ConstellationSparseTopK:    getEnvDefault("CONSTELLATION_SPARSE_RECALL_TOP_K", "10"),
		ConstellationHybridRRFK:    getEnvDefault("CONSTELLATION_HYBRID_RRF_K", "60"),
		AliyunAccessKeyID:          os.Getenv("ALIYUN_ACCESS_KEY_ID"),
		AliyunAccessKeySecret:      os.Getenv("ALIYUN_ACCESS_KEY_SECRET"),
		AliyunSmsSignName:          os.Getenv("ALIYUN_SMS_SIGN_NAME"),
		AliyunSmsTemplateCode:      os.Getenv("ALIYUN_SMS_TEMPLATE_CODE"),
		AliyunSmsCodeLength:        os.Getenv("ALIYUN_SMS_CODE_LENGTH"),
		AliyunSmsValidTime:         os.Getenv("ALIYUN_SMS_VALID_TIME"),
		AliyunSmsInterval:          os.Getenv("ALIYUN_SMS_INTERVAL"),
		TLSDomain:                  os.Getenv("TLS_DOMAIN"),
	}
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadEnvFile searches upward from the current working directory for a
// .env file and sets KEY=VALUE pairs into the environment. OS env vars
// (already set) are never overwritten. Malformed lines are silently skipped.
func loadEnvFile() {
	f := openEnvFile()
	if f == nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		// OS env takes precedence.
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		os.Setenv(key, val)
	}
	_ = scanner.Err()
}

func openEnvFile() *os.File {
	// Start from CWD and walk up to find .env.
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	for {
		p := filepath.Join(cwd, ".env")
		if f, err := os.Open(p); err == nil {
			return f
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return nil
		}
		cwd = parent
	}
}
