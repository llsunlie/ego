package ai

// Config holds the configuration for the AI client.
// Embedding and chat each have their own API key / base URL so they can
// target different providers. When the per-endpoint fields are empty the
// caller should fall back to the shared values.
type Config struct {
	EmbeddingAPIKey  string
	EmbeddingBaseURL string
	EmbeddingModel   string
	ChatAPIKey       string
	ChatBaseURL      string
	ChatModel        string
}
