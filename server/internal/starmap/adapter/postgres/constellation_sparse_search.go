package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/starmap/domain"
)

const searchConstellationCandidatesSQL = `
SELECT constellation_id, similarity(search_text, $1) AS score
FROM constellation_profiles
WHERE user_id = $2
  AND search_text % $1
ORDER BY similarity(search_text, $1) DESC
LIMIT $3
`

const getConstellationProfilesByIDsSQL = `
SELECT
  constellation_id, topic, summary, keywords, emotions, scenes,
  central_pattern, pattern_tags, theme_code, theme_label,
  theme_description, theme_examples, profile_text
FROM constellation_profiles
WHERE user_id = $1 AND constellation_id = ANY($2::uuid[])
`

// ConstellationSparseSearch implements both ConstellationProfileSearchIndexer
// and ConstellationProfileSparseCandidateReader using PostgreSQL pg_trgm.
type ConstellationSparseSearch struct {
	db sqlc.DBTX
}

func NewConstellationSparseSearch(_ *sqlc.Queries, db sqlc.DBTX) *ConstellationSparseSearch {
	return &ConstellationSparseSearch{db: db}
}

// IndexProfile is a no-op: the data is already stored in constellation_profiles,
// and the GIN trigram index on search_text automatically covers it.
func (s *ConstellationSparseSearch) IndexProfile(_ context.Context, _ domain.ConstellationProfile) error {
	return nil
}

// SearchCandidates returns constellation IDs ranked by pg_trgm similarity
// to the trace profile's concatenated search text.
func (s *ConstellationSparseSearch) SearchCandidates(ctx context.Context, profile domain.TraceProfile, limit int) ([]domain.ConstellationProfileSparseCandidate, error) {
	logger := logging.FromContext(ctx)
	if s == nil || s.db == nil || limit <= 0 {
		return nil, nil
	}

	queryText := sparseQueryText(profile)
	if strings.TrimSpace(queryText) == "" {
		return nil, nil
	}

	uid, err := uuid.Parse(profile.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user_id: %w", err)
	}

	start := time.Now()
	rows, err := s.db.Query(ctx, searchConstellationCandidatesSQL,
		queryText,
		pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("sparse search constellation profiles: %w", err)
	}
	defer rows.Close()

	type scored struct {
		id    string
		score float64
	}
	var scoredIDs []scored
	for rows.Next() {
		var (
			id    pgtype.UUID
			score float64
		)
		if err := rows.Scan(&id, &score); err != nil {
			return nil, fmt.Errorf("scan sparse search row: %w", err)
		}
		parsed, err := uuid.FromBytes(id.Bytes[:])
		if err != nil {
			continue
		}
		scoredIDs = append(scoredIDs, scored{id: parsed.String(), score: score})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sparse search rows: %w", err)
	}
	rows.Close()

	if len(scoredIDs) == 0 {
		return nil, nil
	}

	// Load full profile data to compute matched fields and preview.
	scoreMap := make(map[string]float64, len(scoredIDs))
	uuids := make([]pgtype.UUID, len(scoredIDs))
	for i, sc := range scoredIDs {
		scoreMap[sc.id] = sc.score
		pid, _ := uuid.Parse(sc.id)
		uuids[i] = pgtype.UUID{Bytes: [16]byte(pid), Valid: true}
	}

	profileRows, err := s.db.Query(ctx, getConstellationProfilesByIDsSQL,
		pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		uuids,
	)
	if err != nil {
		return nil, fmt.Errorf("load constellation profile details: %w", err)
	}
	defer profileRows.Close()

	type profileRow struct {
		constellationID  pgtype.UUID
		topic            string
		summary          string
		keywords         []byte
		emotions         []byte
		scenes           []byte
		centralPattern   string
		patternTags      []byte
		themeCode        string
		themeLabel       string
		themeDescription string
		themeExamples    []byte
		profileText      string
	}

	var profiles []profileRow
	for profileRows.Next() {
		var r profileRow
		if err := profileRows.Scan(
			&r.constellationID,
			&r.topic, &r.summary, &r.keywords, &r.emotions, &r.scenes,
			&r.centralPattern, &r.patternTags, &r.themeCode, &r.themeLabel,
			&r.themeDescription, &r.themeExamples, &r.profileText,
		); err != nil {
			return nil, fmt.Errorf("scan profile row: %w", err)
		}
		profiles = append(profiles, r)
	}
	if err := profileRows.Err(); err != nil {
		return nil, fmt.Errorf("profile rows: %w", err)
	}

	result := make([]domain.ConstellationProfileSparseCandidate, 0, len(profiles))
	for _, r := range profiles {
		cid, _ := uuid.FromBytes(r.constellationID.Bytes[:])
		cidStr := cid.String()

		doc := sparseDocument{
			topic:            r.topic,
			keywords:         string(r.keywords),
			scenes:           string(r.scenes),
			patternTags:      string(r.patternTags),
			themeCode:        r.themeCode,
			themeLabel:       r.themeLabel,
			summary:          r.summary,
			profileText:      r.profileText,
			themeDescription: r.themeDescription,
			themeExamples:    string(r.themeExamples),
		}

		result = append(result, domain.ConstellationProfileSparseCandidate{
			ConstellationID: cidStr,
			Score:           scoreMap[cidStr],
			MatchedFields:   matchedSparseFields(profile, doc),
			Preview:         doc.preview(),
		})
	}

	logger.DebugContext(ctx, "ConstellationSparseSearch: pg_trgm candidates loaded",
		"user_id", profile.UserID,
		"candidate_count", len(result),
		"limit", limit,
		"elapsed_ms", time.Since(start).Milliseconds(),
	)
	return result, nil
}

// --- Helpers (pure Go, no ES dependency) ---

type sparseDocument struct {
	topic, keywords, scenes, patternTags, themeCode, themeLabel,
	summary, profileText, themeDescription, themeExamples string
}

func sparseQueryText(profile domain.TraceProfile) string {
	parts := []string{
		profile.Topic,
		profile.Summary,
		strings.Join(profile.Keywords, " "),
		strings.Join(profile.Scenes, " "),
		strings.Join(profile.PatternTags, " "),
		profile.CentralPattern,
		profile.ProfileText,
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func matchedSparseFields(profile domain.TraceProfile, doc sparseDocument) []string {
	var fields []string
	if containsAny(profile.Keywords, doc.keywords) {
		fields = append(fields, "keywords")
	}
	if containsAny(profile.Scenes, doc.scenes) {
		fields = append(fields, "scenes")
	}
	if containsAny(profile.PatternTags, doc.patternTags) {
		fields = append(fields, "pattern_tags")
	}
	if strings.TrimSpace(profile.Topic) != "" && strings.Contains(doc.topic, strings.TrimSpace(profile.Topic)) {
		fields = append(fields, "topic")
	}
	return fields
}

func containsAny(values []string, text string) bool {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && strings.Contains(text, value) {
			return true
		}
	}
	return false
}

func (d sparseDocument) preview() string {
	for _, value := range []string{d.topic, d.themeLabel, d.summary, d.profileText} {
		value = strings.TrimSpace(value)
		if value != "" {
			rs := []rune(value)
			if len(rs) > 48 {
				return string(rs[:48])
			}
			return value
		}
	}
	return ""
}

var _ domain.ConstellationProfileSearchIndexer = (*ConstellationSparseSearch)(nil)
var _ domain.ConstellationProfileSparseCandidateReader = (*ConstellationSparseSearch)(nil)
