CREATE TABLE stars (
  id         UUID PRIMARY KEY,
  user_id    UUID NOT NULL,
  trace_id   UUID NOT NULL,
  topic      TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX idx_stars_trace ON stars(trace_id);

CREATE TABLE constellations (
  id                   UUID PRIMARY KEY,
  user_id              UUID NOT NULL,
  name                 VARCHAR(100) NOT NULL,
  constellation_insight TEXT NOT NULL,
  star_ids             UUID[] NOT NULL,
  topic_prompts        TEXT[] NOT NULL,
  created_at           TIMESTAMPTZ NOT NULL,
  updated_at           TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_constellations_user ON constellations(user_id);
CREATE INDEX idx_constellations_stars ON constellations USING GIN (star_ids);
