CREATE TABLE traces (
  id         UUID PRIMARY KEY,
  user_id    UUID NOT NULL,
  motivation VARCHAR(50) NOT NULL,
  stashed    BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_traces_user_created ON traces(user_id, created_at DESC);
