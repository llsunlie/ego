package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is a minimal OpenAI-compatible HTTP client providing embedding
// and chat completion. It holds no business logic and no domain knowledge.
type Client struct {
	cfg       Config
	http      *http.Client
	embedBase string // trailing-slash-free embedding base URL
	chatBase  string // trailing-slash-free chat base URL
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:       cfg,
		http:      &http.Client{},
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
	body := embeddingRequest{
		Model: c.cfg.EmbeddingModel,
		Input: input,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("ai.Embed: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.embedBase+"/embeddings", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("ai.Embed: request: %w", err)
	}
	c.setHeaders(req, c.cfg.EmbeddingAPIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai.Embed: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.readError("ai.Embed", resp)
	}

	var parsed embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("ai.Embed: decode: %w", err)
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("ai.Embed: empty data")
	}
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
}

// Chat sends messages to the chat completions endpoint and returns the
// first choice's content. Callers are responsible for constructing their
// own system and user prompts.
func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	body := chatRequest{
		Model:    c.cfg.ChatModel,
		Messages: messages,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("ai.Chat: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatBase+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("ai.Chat: request: %w", err)
	}
	c.setHeaders(req, c.cfg.ChatAPIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai.Chat: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.readError("ai.Chat", resp)
	}

	var parsed chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("ai.Chat: decode: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("ai.Chat: empty choices")
	}
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
