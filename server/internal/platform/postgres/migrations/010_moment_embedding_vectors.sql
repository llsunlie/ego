CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE moment_embedding_vectors (
  moment_id  UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL,
  trace_id   UUID NOT NULL,
  model      TEXT NOT NULL,
  dim        INT NOT NULL,
  embedding  VECTOR(4096) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (moment_id, model)
);

CREATE INDEX idx_moment_embedding_vectors_user_model
ON moment_embedding_vectors(user_id, model);

CREATE INDEX idx_moment_embedding_vectors_trace
ON moment_embedding_vectors(trace_id);

CREATE INDEX idx_moment_embedding_vectors_embedding_hnsw
ON moment_embedding_vectors
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
