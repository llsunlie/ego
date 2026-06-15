CREATE TABLE constellation_profiles (
  constellation_id UUID PRIMARY KEY REFERENCES constellations(id) ON DELETE CASCADE,
  user_id          UUID NOT NULL,
  topic            TEXT NOT NULL,
  summary          TEXT NOT NULL,
  keywords         JSONB NOT NULL DEFAULT '[]'::JSONB,
  emotions         JSONB NOT NULL DEFAULT '[]'::JSONB,
  scenes           JSONB NOT NULL DEFAULT '[]'::JSONB,
  central_pattern  TEXT NOT NULL,
  pattern_tags     JSONB NOT NULL DEFAULT '[]'::JSONB,
  theme_code       TEXT NOT NULL DEFAULT '',
  theme_label      TEXT NOT NULL DEFAULT '',
  theme_description TEXT NOT NULL DEFAULT '',
  theme_examples   JSONB NOT NULL DEFAULT '[]'::JSONB,
  profile_text     TEXT NOT NULL,
  trace_count      DOUBLE PRECISION NOT NULL,
  moment_count     DOUBLE PRECISION NOT NULL,
  status           TEXT NOT NULL,
  last_error       TEXT NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL,
  updated_at       TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_constellation_profiles_user_status
ON constellation_profiles(user_id, status);

CREATE TABLE constellation_profile_vectors (
  constellation_id    UUID PRIMARY KEY REFERENCES constellation_profiles(constellation_id) ON DELETE CASCADE,
  user_id             UUID NOT NULL,
  model               TEXT NOT NULL,
  dim                 INT NOT NULL,
  profile_embedding   VECTOR(1024) NOT NULL,
  centroid_embedding  VECTOR(1024) NOT NULL,
  created_at          TIMESTAMPTZ NOT NULL,
  updated_at          TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_constellation_profile_vectors_user_model
ON constellation_profile_vectors(user_id, model);

CREATE INDEX idx_constellation_profile_vectors_profile_embedding_hnsw
ON constellation_profile_vectors
USING hnsw (profile_embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

CREATE TABLE constellation_stars (
  constellation_id UUID NOT NULL REFERENCES constellations(id) ON DELETE CASCADE,
  star_id          UUID NOT NULL REFERENCES stars(id) ON DELETE CASCADE,
  trace_id         UUID NOT NULL REFERENCES traces(id) ON DELETE CASCADE,
  user_id          UUID NOT NULL,
  match_score      DOUBLE PRECISION NOT NULL,
  match_type       TEXT NOT NULL,
  match_dimensions JSONB NOT NULL DEFAULT '[]'::JSONB,
  match_reason     TEXT NOT NULL,
  weight           DOUBLE PRECISION NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (constellation_id, star_id)
);

CREATE INDEX idx_constellation_stars_star
ON constellation_stars(star_id);

CREATE INDEX idx_constellation_stars_user_constellation
ON constellation_stars(user_id, constellation_id);
