# platform/ai

OpenAI-compatible HTTP client providing embedding and chat completion.

## Exports

| File | What |
|------|------|
| `config.go` | `Config` struct (APIKey, BaseURL, EmbeddingModel, ChatModel) |
| `client.go` | `Client` with `CreateEmbedding(ctx, input) (*EmbeddingResult, error)` and `Chat(ctx, messages) (string, error)` |
| `retry.go` | Retry wrappers `CreateEmbeddingWithRetry` and `ChatWithRetry` |
| `similarity.go` | `CosineSimilarity(a, b []float32) float64` |

## EmbeddingResult

```go
type EmbeddingResult struct {
    Embedding []float32
    Model     string          // e.g. "text-embedding-3-small"
    Usage     EmbeddingUsage  // token consumption
}
```

Callers can use `Model` to detect model changes and invalidate cached embeddings.

## Usage

`Client` is created in `bootstrap.InitPlatform()` from `config.Config` env vars and stored on `bootstrap.Platform.AIClient`.

Business modules consume the client through their own `adapter/ai/` layer — see `writing/adapter/ai/`, `conversation/adapter/ai/`, `starmap/adapter/ai/`.

## Retry

`CreateEmbedding` and `Chat` are single-attempt primitives. Business adapters should use retry wrappers for non-interactive or critical AI calls:

- `CreateEmbeddingWithRetry(ctx, input, DefaultEmbeddingRetryOptions())`
- `ChatWithRetry(ctx, messages, DefaultChatRetryOptions())`

The default policy retries retryable transport, decode, empty response, rate limit, and 5xx-style errors. It does not retry 400 / 401 / 403 / 404 responses.
