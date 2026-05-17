CREATE TABLE insights (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  moment_id          UUID NOT NULL,
  echo_id            UUID,
  text               TEXT NOT NULL,
  related_moment_ids UUID[] NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_insights_moment ON insights(moment_id);
CREATE INDEX idx_insights_echo   ON insights(echo_id);
