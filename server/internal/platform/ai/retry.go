package ai

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type RetryOptions struct {
	MaxAttempts int
	Backoff     func(attempt int) time.Duration
	Operation   string
}

func DefaultEmbeddingRetryOptions() RetryOptions {
	return RetryOptions{
		MaxAttempts: 3,
		Backoff:     defaultAIBackoff,
		Operation:   "embedding",
	}
}

func DefaultChatRetryOptions() RetryOptions {
	return RetryOptions{
		MaxAttempts: 2,
		Backoff:     defaultAIBackoff,
		Operation:   "chat",
	}
}

func defaultAIBackoff(attempt int) time.Duration {
	switch attempt {
	case 1:
		return 200 * time.Millisecond
	default:
		return 600 * time.Millisecond
	}
}

func (c *Client) CreateEmbeddingWithRetry(ctx context.Context, input string, opts RetryOptions) (*EmbeddingResult, error) {
	opts = normalizeRetryOptions(opts, "embedding")
	var lastErr error
	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		result, err := c.CreateEmbedding(ctx, input)
		if err == nil {
			if attempt > 1 {
				c.logger.InfoContext(ctx, "ai.Embed: retry succeeded",
					"operation", opts.Operation,
					"attempt", attempt,
				)
			}
			return result, nil
		}
		lastErr = err
		if ctx.Err() != nil || !isRetryableAIError(err, "ai.Embed") || attempt == opts.MaxAttempts {
			break
		}
		delay := opts.Backoff(attempt)
		c.logger.WarnContext(ctx, "ai.Embed: retryable attempt failed",
			"operation", opts.Operation,
			"attempt", attempt,
			"max_attempts", opts.MaxAttempts,
			"retry_after_ms", delay.Milliseconds(),
			"error", err,
		)
		if err := sleepWithContext(ctx, delay); err != nil {
			return nil, err
		}
	}
	return nil, lastErr
}

func (c *Client) ChatWithRetry(ctx context.Context, messages []ChatMessage, opts RetryOptions) (string, error) {
	opts = normalizeRetryOptions(opts, "chat")
	var lastErr error
	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		result, err := c.Chat(ctx, messages)
		if err == nil {
			if attempt > 1 {
				c.logger.InfoContext(ctx, "ai.Chat: retry succeeded",
					"operation", opts.Operation,
					"attempt", attempt,
				)
			}
			return result, nil
		}
		lastErr = err
		if ctx.Err() != nil || !isRetryableAIError(err, "ai.Chat") || attempt == opts.MaxAttempts {
			break
		}
		delay := opts.Backoff(attempt)
		c.logger.WarnContext(ctx, "ai.Chat: retryable attempt failed",
			"operation", opts.Operation,
			"attempt", attempt,
			"max_attempts", opts.MaxAttempts,
			"retry_after_ms", delay.Milliseconds(),
			"error", err,
		)
		if err := sleepWithContext(ctx, delay); err != nil {
			return "", err
		}
	}
	return "", lastErr
}

func normalizeRetryOptions(opts RetryOptions, operation string) RetryOptions {
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 1
	}
	if opts.Backoff == nil {
		opts.Backoff = func(int) time.Duration { return 0 }
	}
	if strings.TrimSpace(opts.Operation) == "" {
		opts.Operation = operation
	}
	return opts
}

func isRetryableAIError(err error, tag string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, status := range []string{"400", "401", "403", "404"} {
		if strings.Contains(msg, fmt.Sprintf("%s: %s", tag, status)) {
			return false
		}
	}
	return true
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
