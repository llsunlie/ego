CREATE TABLE echos (
  id                 UUID PRIMARY KEY,
  moment_id          UUID NOT NULL,
  user_id            UUID NOT NULL,
  matched_moment_ids UUID[] NOT NULL,
  similarities       FLOAT[] NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_echos_moment ON echos(moment_id);
