CREATE TABLE traces (
  id         UUID PRIMARY KEY,
  user_id    UUID NOT NULL,
  topic      TEXT,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_traces_user_id ON traces(user_id);
