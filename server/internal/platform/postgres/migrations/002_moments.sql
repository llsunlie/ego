CREATE TABLE moments (
  id         UUID PRIMARY KEY,
  trace_id   UUID NOT NULL,
  user_id    UUID NOT NULL,
  content    TEXT NOT NULL,
  embeddings JSONB NOT NULL DEFAULT '[]'::JSONB,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_moments_user_id ON moments(user_id);
CREATE INDEX idx_moments_trace_id ON moments(trace_id);
CREATE INDEX idx_moments_created_at ON moments(created_at DESC);
