# platform/ai

OpenAI-compatible HTTP client providing embedding and chat completion.

## Exports

| File | What |
|------|------|
| `config.go` | `Config` struct (APIKey, BaseURL, EmbeddingModel, ChatModel) |
| `client.go` | `Client` with `CreateEmbedding(ctx, input) (*EmbeddingResult, error)` and `Chat(ctx, messages) (string, error)` |
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
