package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"ego-server/internal/platform/metrics"
)

// Client is a minimal OpenAI-compatible HTTP client providing embedding
// and chat completion. It holds no business logic and no domain knowledge.
type Client struct {
	cfg       Config
	http      *http.Client
	logger    *slog.Logger
	embedBase string // trailing-slash-free embedding base URL
	chatBase  string // trailing-slash-free chat base URL
}

func NewClient(cfg Config, logger *slog.Logger) *Client {
	return &Client{
		cfg:       cfg,
		http:      &http.Client{},
		logger:    logger,
		embedBase: strings.TrimRight(cfg.EmbeddingBaseURL, "/"),
		chatBase:  strings.TrimRight(cfg.ChatBaseURL, "/"),
	}
}

// --- Embeddings -----------------------------------------------------------

// EmbeddingResult wraps the embedding vector with metadata so callers can
// record provenance and decide when to refresh cached embeddings.
type EmbeddingResult struct {
	Embedding []float32
	Model     string
	Usage     EmbeddingUsage
}

// EmbeddingUsage records token consumption for an embedding request.
type EmbeddingUsage struct {
	PromptTokens int
	TotalTokens  int
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// CreateEmbedding calls the OpenAI embeddings endpoint and returns the
// embedding vector together with model and usage metadata.
func (c *Client) CreateEmbedding(ctx context.Context, input string) (*EmbeddingResult, error) {
	metrics.AiCallsInFlight.Inc()
	defer metrics.AiCallsInFlight.Dec()

	start := time.Now()
	model := c.cfg.EmbeddingModel

	body := embeddingRequest{
		Model: model,
		Input: input,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		metrics.AiEmbedTotal.WithLabelValues(model, "error").Inc()
		return nil, fmt.Errorf("ai.Embed: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.embedBase+"/embeddings", bytes.NewReader(reqBody))
	if err != nil {
		metrics.AiEmbedTotal.WithLabelValues(model, "error").Inc()
		return nil, fmt.Errorf("ai.Embed: request: %w", err)
	}
	c.setHeaders(req, c.cfg.EmbeddingAPIKey)

	c.logger.InfoContext(ctx, "ai.Embed: request",
		"model", model,
		"input_len", len([]rune(input)),
	)

	resp, err := c.http.Do(req)
	if err != nil {
		metrics.AiEmbedTotal.WithLabelValues(model, "error").Inc()
		c.logger.ErrorContext(ctx, "ai.Embed: error", "error", err, "elapsed_ms", time.Since(start).Milliseconds())
		return nil, fmt.Errorf("ai.Embed: do: %w", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start).Seconds()
	metrics.AiEmbedDuration.WithLabelValues(model).Observe(elapsed)

	if resp.StatusCode != http.StatusOK {
		metrics.AiEmbedTotal.WithLabelValues(model, fmt.Sprintf("%d", resp.StatusCode)).Inc()
		return nil, c.readError("ai.Embed", resp)
	}

	var parsed embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		metrics.AiEmbedTotal.WithLabelValues(model, "error").Inc()
		return nil, fmt.Errorf("ai.Embed: decode: %w", err)
	}
	if len(parsed.Data) == 0 {
		metrics.AiEmbedTotal.WithLabelValues(model, "error").Inc()
		return nil, fmt.Errorf("ai.Embed: empty data")
	}

	metrics.AiEmbedTotal.WithLabelValues(model, "ok").Inc()
	metrics.AiEmbedTokens.WithLabelValues(model).Add(float64(parsed.Usage.TotalTokens))

	c.logger.InfoContext(ctx, "ai.Embed: ok",
		"model", model,
		"tokens", parsed.Usage.TotalTokens,
		"elapsed_ms", time.Since(start).Milliseconds(),
	)

	return &EmbeddingResult{
		Embedding: parsed.Data[0].Embedding,
		Model:     parsed.Model,
		Usage: EmbeddingUsage{
			PromptTokens: parsed.Usage.PromptTokens,
			TotalTokens:  parsed.Usage.TotalTokens,
		},
	}, nil
}

// --- Chat ------------------------------------------------------------------

// ChatMessage represents a single message in a chat completion conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Chat sends messages to the chat completions endpoint and returns the
// first choice's content. Callers are responsible for constructing their
// own system and user prompts.
func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	metrics.AiCallsInFlight.Inc()
	defer metrics.AiCallsInFlight.Dec()

	start := time.Now()
	model := c.cfg.ChatModel

	body := chatRequest{
		Model:    model,
		Messages: messages,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		metrics.AiChatTotal.WithLabelValues(model, "error").Inc()
		return "", fmt.Errorf("ai.Chat: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatBase+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		metrics.AiChatTotal.WithLabelValues(model, "error").Inc()
		return "", fmt.Errorf("ai.Chat: request: %w", err)
	}
	c.setHeaders(req, c.cfg.ChatAPIKey)

	c.logger.InfoContext(ctx, "ai.Chat: request",
		"model", model,
		"messages", len(messages),
		"preview", truncate(lastUserContent(messages), 200),
	)

	resp, err := c.http.Do(req)
	if err != nil {
		metrics.AiChatTotal.WithLabelValues(model, "error").Inc()
		c.logger.ErrorContext(ctx, "ai.Chat: error", "error", err, "elapsed_ms", time.Since(start).Milliseconds())
		return "", fmt.Errorf("ai.Chat: do: %w", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start).Seconds()
	metrics.AiChatDuration.WithLabelValues(model).Observe(elapsed)

	if resp.StatusCode != http.StatusOK {
		metrics.AiChatTotal.WithLabelValues(model, fmt.Sprintf("%d", resp.StatusCode)).Inc()
		return "", c.readError("ai.Chat", resp)
	}

	var parsed chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		metrics.AiChatTotal.WithLabelValues(model, "error").Inc()
		return "", fmt.Errorf("ai.Chat: decode: %w", err)
	}
	if len(parsed.Choices) == 0 {
		metrics.AiChatTotal.WithLabelValues(model, "error").Inc()
		return "", fmt.Errorf("ai.Chat: empty choices")
	}

	metrics.AiChatTotal.WithLabelValues(model, "ok").Inc()
	metrics.AiChatTokens.WithLabelValues(model, "prompt").Add(float64(parsed.Usage.PromptTokens))
	metrics.AiChatTokens.WithLabelValues(model, "completion").Add(float64(parsed.Usage.CompletionTokens))

	c.logger.InfoContext(ctx, "ai.Chat: ok",
		"model", model,
		"prompt_tokens", parsed.Usage.PromptTokens,
		"completion_tokens", parsed.Usage.CompletionTokens,
		"elapsed_ms", time.Since(start).Milliseconds(),
		"preview", truncate(parsed.Choices[0].Message.Content, 200),
	)

	return parsed.Choices[0].Message.Content, nil
}

// --- helpers ---------------------------------------------------------------

func (c *Client) setHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) readError(tag string, resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("%s: %d %s", tag, resp.StatusCode, string(body))
}

func lastUserContent(messages []ChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	if len(messages) > 0 {
		return messages[len(messages)-1].Content
	}
	return ""
}

func truncate(s string, n int) string {
	rs := []rune(s)
	if len(rs) > n {
		return string(rs[:n]) + "..."
	}
	return s
}
