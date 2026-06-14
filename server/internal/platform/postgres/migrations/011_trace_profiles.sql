CREATE TABLE trace_profiles (
  trace_id                 UUID PRIMARY KEY REFERENCES traces(id) ON DELETE CASCADE,
  user_id                  UUID NOT NULL,
  topic                    TEXT NOT NULL,
  summary                  TEXT NOT NULL,
  keywords                 JSONB NOT NULL DEFAULT '[]'::JSONB,
  emotions                 JSONB NOT NULL DEFAULT '[]'::JSONB,
  scenes                   JSONB NOT NULL DEFAULT '[]'::JSONB,
  central_pattern          TEXT NOT NULL,
  pattern_tags             JSONB NOT NULL DEFAULT '[]'::JSONB,
  representative_moment_id UUID,
  profile_text             TEXT NOT NULL,
  status                   TEXT NOT NULL,
  retry_count              INT NOT NULL,
  last_error               TEXT NOT NULL,
  created_at               TIMESTAMPTZ NOT NULL,
  updated_at               TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_trace_profiles_user_status
ON trace_profiles(user_id, status);

CREATE TABLE trace_profile_vectors (
  trace_id   UUID PRIMARY KEY REFERENCES trace_profiles(trace_id) ON DELETE CASCADE,
  user_id    UUID NOT NULL,
  model      TEXT NOT NULL,
  dim        INT NOT NULL,
  embedding  VECTOR(1024) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_trace_profile_vectors_user_model
ON trace_profile_vectors(user_id, model);

CREATE INDEX idx_trace_profile_vectors_embedding_hnsw
ON trace_profile_vectors
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
