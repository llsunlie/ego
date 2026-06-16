CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Moments: GIN trigram index on content for sparse text search
CREATE INDEX IF NOT EXISTS idx_moments_content_trgm
  ON moments USING gin (content gin_trgm_ops);

-- Constellation profiles: generated column consolidating all searchable text fields
ALTER TABLE constellation_profiles
  ADD COLUMN IF NOT EXISTS search_text TEXT
  GENERATED ALWAYS AS (
    coalesce(topic, '') || ' ' ||
    coalesce(summary, '') || ' ' ||
    coalesce(keywords::text, '') || ' ' ||
    coalesce(emotions::text, '') || ' ' ||
    coalesce(scenes::text, '') || ' ' ||
    coalesce(central_pattern, '') || ' ' ||
    coalesce(pattern_tags::text, '') || ' ' ||
    coalesce(theme_label, '') || ' ' ||
    coalesce(theme_description, '') || ' ' ||
    coalesce(theme_examples::text, '') || ' ' ||
    coalesce(profile_text, '')
  ) STORED;

-- GIN trigram index on the consolidated search column
CREATE INDEX IF NOT EXISTS idx_constellation_profiles_search_trgm
  ON constellation_profiles USING gin (search_text gin_trgm_ops);
