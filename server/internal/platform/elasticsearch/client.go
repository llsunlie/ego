package elasticsearch

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
)

type Config struct {
	URL      string
	Username string
	Password string
}

type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient(cfg Config, logger *slog.Logger) *Client {
	return &Client{
		baseURL:  strings.TrimRight(cfg.URL, "/"),
		username: cfg.Username,
		password: cfg.Password,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) DoJSON(ctx context.Context, method string, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}
	return c.Do(ctx, method, path, "application/json", reader, out)
}

func (c *Client) DoNDJSON(ctx context.Context, method string, path string, body []byte, out any) error {
	return c.Do(ctx, method, path, "application/x-ndjson", bytes.NewReader(body), out)
}

func (c *Client) Do(ctx context.Context, method string, path string, contentType string, body io.Reader, out any) error {
	if c == nil {
		return fmt.Errorf("elasticsearch client is nil")
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("elasticsearch %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("elasticsearch %s %s: status %d: %s", method, path, resp.StatusCode, string(data))
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
