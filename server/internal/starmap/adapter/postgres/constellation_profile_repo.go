package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/starmap/domain"
)

type ConstellationProfileRepository struct {
	db           sqlc.DBTX
	embeddingDim int
}

func NewConstellationProfileRepository(db sqlc.DBTX, embeddingDim int) *ConstellationProfileRepository {
	return &ConstellationProfileRepository{db: db, embeddingDim: embeddingDim}
}

const findConstellationProfileCandidatesSQL = `
SELECT
  cp.constellation_id,
  cp.user_id,
  cp.topic,
  cp.summary,
  cp.keywords,
  cp.emotions,
  cp.scenes,
  cp.central_pattern,
  cp.pattern_tags,
  cp.theme_code,
  cp.theme_label,
  cp.theme_description,
  cp.theme_examples,
  cp.profile_text,
  cp.trace_count,
  cp.moment_count,
  cp.status,
  cp.last_error,
  cp.created_at,
  cp.updated_at,
  cv.model,
  cv.dim,
  cv.profile_embedding::text,
  cv.centroid_embedding::text
FROM constellation_profiles cp
JOIN constellation_profile_vectors cv ON cv.constellation_id = cp.constellation_id
WHERE cp.user_id = $1
ORDER BY cv.profile_embedding <=> $2::vector
LIMIT $3
`

const upsertConstellationProfileSQL = `
INSERT INTO constellation_profiles (
  constellation_id, user_id, topic, summary, keywords, emotions, scenes,
  central_pattern, pattern_tags, theme_code, theme_label, theme_description,
  theme_examples, profile_text, trace_count, moment_count, status, last_error,
  created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11, $12,
  $13, $14, $15, $16, $17,
  $18, $19, $20
)
ON CONFLICT (constellation_id) DO UPDATE SET
  user_id = EXCLUDED.user_id,
  topic = EXCLUDED.topic,
  summary = EXCLUDED.summary,
  keywords = EXCLUDED.keywords,
  emotions = EXCLUDED.emotions,
  scenes = EXCLUDED.scenes,
  central_pattern = EXCLUDED.central_pattern,
  pattern_tags = EXCLUDED.pattern_tags,
  theme_code = EXCLUDED.theme_code,
  theme_label = EXCLUDED.theme_label,
  theme_description = EXCLUDED.theme_description,
  theme_examples = EXCLUDED.theme_examples,
  profile_text = EXCLUDED.profile_text,
  trace_count = EXCLUDED.trace_count,
  moment_count = EXCLUDED.moment_count,
  status = EXCLUDED.status,
  last_error = EXCLUDED.last_error,
  updated_at = EXCLUDED.updated_at
`

const upsertConstellationProfileVectorSQL = `
INSERT INTO constellation_profile_vectors (
  constellation_id, user_id, model, dim, profile_embedding, centroid_embedding, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5::vector, $6::vector, $7, $8
)
ON CONFLICT (constellation_id) DO UPDATE SET
  user_id = EXCLUDED.user_id,
  model = EXCLUDED.model,
  dim = EXCLUDED.dim,
  profile_embedding = EXCLUDED.profile_embedding,
  centroid_embedding = EXCLUDED.centroid_embedding,
  updated_at = EXCLUDED.updated_at
`

const upsertConstellationMembershipSQL = `
INSERT INTO constellation_stars (
  constellation_id, star_id, trace_id, user_id, match_score, match_type,
  match_dimensions, match_reason, weight, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10
)
ON CONFLICT (constellation_id, star_id) DO UPDATE SET
  trace_id = EXCLUDED.trace_id,
  user_id = EXCLUDED.user_id,
  match_score = EXCLUDED.match_score,
  match_type = EXCLUDED.match_type,
  match_dimensions = EXCLUDED.match_dimensions,
  match_reason = EXCLUDED.match_reason,
  weight = EXCLUDED.weight
`

func (r *ConstellationProfileRepository) FindCandidates(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db is required for constellation profile candidates")
	}
	if limit <= 0 {
		return nil, nil
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user_id: %w", err)
	}
	literal, err := vectorLiteral(embedding, r.embeddingDim)
	if err != nil {
		return nil, fmt.Errorf("constellation candidate vector literal: %w", err)
	}

	rows, err := r.db.Query(ctx, findConstellationProfileCandidatesSQL,
		pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		literal,
		int32(limit),
	)
	if err != nil {
		return nil, fmt.Errorf("find constellation profile candidates: %w", err)
	}
	defer rows.Close()

	var candidates []domain.ConstellationProfileCandidate
	for rows.Next() {
		var (
			constellationID pgtype.UUID
			rowUserID       pgtype.UUID
			profile         domain.ConstellationProfile
			keywords        []byte
			emotions        []byte
			scenes          []byte
			patternTags     []byte
			themeExamples   []byte
			model           string
			dim             int32
			profileVector   string
			centroidVector  string
			createdAt       pgtype.Timestamptz
			updatedAt       pgtype.Timestamptz
		)
		if err := rows.Scan(
			&constellationID,
			&rowUserID,
			&profile.Topic,
			&profile.Summary,
			&keywords,
			&emotions,
			&scenes,
			&profile.CentralPattern,
			&patternTags,
			&profile.ThemeCode,
			&profile.ThemeLabel,
			&profile.ThemeDescription,
			&themeExamples,
			&profile.ProfileText,
			&profile.TraceCount,
			&profile.MomentCount,
			&profile.Status,
			&profile.LastError,
			&createdAt,
			&updatedAt,
			&model,
			&dim,
			&profileVector,
			&centroidVector,
		); err != nil {
			return nil, fmt.Errorf("scan constellation profile candidate: %w", err)
		}
		cid, _ := uuid.FromBytes(constellationID.Bytes[:])
		ruid, _ := uuid.FromBytes(rowUserID.Bytes[:])
		profile.ConstellationID = cid.String()
		profile.UserID = ruid.String()
		profile.CreatedAt = createdAt.Time
		profile.UpdatedAt = updatedAt.Time
		if err := unmarshalStringList(keywords, &profile.Keywords); err != nil {
			return nil, fmt.Errorf("unmarshal keywords: %w", err)
		}
		if err := unmarshalStringList(emotions, &profile.Emotions); err != nil {
			return nil, fmt.Errorf("unmarshal emotions: %w", err)
		}
		if err := unmarshalStringList(scenes, &profile.Scenes); err != nil {
			return nil, fmt.Errorf("unmarshal scenes: %w", err)
		}
		if err := unmarshalStringList(patternTags, &profile.PatternTags); err != nil {
			return nil, fmt.Errorf("unmarshal pattern tags: %w", err)
		}
		if err := unmarshalStringList(themeExamples, &profile.ThemeExamples); err != nil {
			return nil, fmt.Errorf("unmarshal theme examples: %w", err)
		}

		pv, err := parseVectorText(profileVector)
		if err != nil {
			return nil, fmt.Errorf("parse profile embedding: %w", err)
		}
		cv, err := parseVectorText(centroidVector)
		if err != nil {
			return nil, fmt.Errorf("parse centroid embedding: %w", err)
		}
		candidates = append(candidates, domain.ConstellationProfileCandidate{
			Profile: profile,
			Vector: domain.ConstellationProfileVector{
				ConstellationID:   profile.ConstellationID,
				UserID:            profile.UserID,
				Model:             model,
				Dim:               int(dim),
				ProfileEmbedding:  pv,
				CentroidEmbedding: cv,
				CreatedAt:         profile.CreatedAt,
				UpdatedAt:         profile.UpdatedAt,
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate constellation profile candidates: %w", err)
	}
	return candidates, nil
}

func (r *ConstellationProfileRepository) Upsert(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
	if r.db == nil {
		return fmt.Errorf("db is required for constellation profile upsert")
	}
	if profile == nil {
		return fmt.Errorf("constellation profile is nil")
	}
	constellationID, err := uuid.Parse(profile.ConstellationID)
	if err != nil {
		return fmt.Errorf("parse constellation_id: %w", err)
	}
	userID, err := uuid.Parse(profile.UserID)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}
	keywords, err := json.Marshal(profile.Keywords)
	if err != nil {
		return fmt.Errorf("marshal keywords: %w", err)
	}
	emotions, err := json.Marshal(profile.Emotions)
	if err != nil {
		return fmt.Errorf("marshal emotions: %w", err)
	}
	scenes, err := json.Marshal(profile.Scenes)
	if err != nil {
		return fmt.Errorf("marshal scenes: %w", err)
	}
	patternTags, err := json.Marshal(profile.PatternTags)
	if err != nil {
		return fmt.Errorf("marshal pattern tags: %w", err)
	}
	themeExamples, err := json.Marshal(profile.ThemeExamples)
	if err != nil {
		return fmt.Errorf("marshal theme examples: %w", err)
	}
	createdAt := profile.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := profile.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	if _, err := r.db.Exec(ctx, upsertConstellationProfileSQL,
		pgtype.UUID{Bytes: [16]byte(constellationID), Valid: true},
		pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		profile.Topic,
		profile.Summary,
		keywords,
		emotions,
		scenes,
		profile.CentralPattern,
		patternTags,
		profile.ThemeCode,
		profile.ThemeLabel,
		profile.ThemeDescription,
		themeExamples,
		profile.ProfileText,
		profile.TraceCount,
		profile.MomentCount,
		profile.Status,
		profile.LastError,
		pgtype.Timestamptz{Time: createdAt, Valid: true},
		pgtype.Timestamptz{Time: updatedAt, Valid: true},
	); err != nil {
		return fmt.Errorf("upsert constellation profile: %w", err)
	}

	if vector == nil {
		return nil
	}
	profileLiteral, err := vectorLiteral(vector.ProfileEmbedding, r.embeddingDim)
	if err != nil {
		return fmt.Errorf("constellation profile vector literal: %w", err)
	}
	centroidLiteral, err := vectorLiteral(vector.CentroidEmbedding, r.embeddingDim)
	if err != nil {
		return fmt.Errorf("constellation centroid vector literal: %w", err)
	}
	if vector.Model == "" {
		return fmt.Errorf("constellation profile vector model is empty")
	}
	if _, err := r.db.Exec(ctx, upsertConstellationProfileVectorSQL,
		pgtype.UUID{Bytes: [16]byte(constellationID), Valid: true},
		pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		vector.Model,
		int32(vector.Dim),
		profileLiteral,
		centroidLiteral,
		pgtype.Timestamptz{Time: createdAt, Valid: true},
		pgtype.Timestamptz{Time: updatedAt, Valid: true},
	); err != nil {
		return fmt.Errorf("upsert constellation profile vector: %w", err)
	}
	return nil
}

func (r *ConstellationProfileRepository) AddMembership(ctx context.Context, membership domain.ConstellationMembership) error {
	if r.db == nil {
		return fmt.Errorf("db is required for constellation membership")
	}
	constellationID, err := uuid.Parse(membership.ConstellationID)
	if err != nil {
		return fmt.Errorf("parse constellation_id: %w", err)
	}
	starID, err := uuid.Parse(membership.StarID)
	if err != nil {
		return fmt.Errorf("parse star_id: %w", err)
	}
	traceID, err := uuid.Parse(membership.TraceID)
	if err != nil {
		return fmt.Errorf("parse trace_id: %w", err)
	}
	userID, err := uuid.Parse(membership.UserID)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}
	dimensions, err := json.Marshal(membership.MatchDimensions)
	if err != nil {
		return fmt.Errorf("marshal match dimensions: %w", err)
	}
	createdAt := membership.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	if _, err := r.db.Exec(ctx, upsertConstellationMembershipSQL,
		pgtype.UUID{Bytes: [16]byte(constellationID), Valid: true},
		pgtype.UUID{Bytes: [16]byte(starID), Valid: true},
		pgtype.UUID{Bytes: [16]byte(traceID), Valid: true},
		pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		membership.MatchScore,
		membership.MatchType,
		dimensions,
		membership.MatchReason,
		membership.Weight,
		pgtype.Timestamptz{Time: createdAt, Valid: true},
	); err != nil {
		return fmt.Errorf("upsert constellation membership: %w", err)
	}
	return nil
}

func unmarshalStringList(raw []byte, dest *[]string) error {
	if len(raw) == 0 {
		*dest = nil
		return nil
	}
	return json.Unmarshal(raw, dest)
}

func parseVectorText(value string) ([]float32, error) {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	result := make([]float32, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parsed, err := strconv.ParseFloat(part, 32)
		if err != nil {
			return nil, err
		}
		result = append(result, float32(parsed))
	}
	return result, nil
}

var _ domain.ConstellationProfileRepository = (*ConstellationProfileRepository)(nil)
