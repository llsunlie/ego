package elasticsearch

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	platformes "ego-server/internal/platform/elasticsearch"
	"ego-server/internal/starmap/domain"
)

const DefaultConstellationProfileIndex = "ego_constellation_profiles"

type ConstellationProfileSearch struct {
	client  *platformes.Client
	index   string
	mu      sync.Mutex
	ensured bool
}

func NewConstellationProfileSearch(client *platformes.Client, index string) *ConstellationProfileSearch {
	if index == "" {
		index = DefaultConstellationProfileIndex
	}
	return &ConstellationProfileSearch{client: client, index: index}
}

func (s *ConstellationProfileSearch) EnsureIndex(ctx context.Context) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("elasticsearch client is required")
	}
	var exists map[string]any
	err := s.client.Do(ctx, http.MethodHead, "/"+s.index, "", nil, &exists)
	if err == nil {
		return nil
	}

	body := map[string]any{
		"settings": map[string]any{
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"ego_ik_index": map[string]any{
						"type":      "custom",
						"tokenizer": "ik_max_word",
					},
					"ego_ik_search": map[string]any{
						"type":      "custom",
						"tokenizer": "ik_smart",
					},
					"ego_char_bigram": map[string]any{
						"type":      "custom",
						"tokenizer": "ego_char_bigram_tokenizer",
					},
				},
				"tokenizer": map[string]any{
					"ego_char_bigram_tokenizer": map[string]any{
						"type":        "ngram",
						"min_gram":    2,
						"max_gram":    3,
						"token_chars": []string{"letter", "digit"},
					},
				},
			},
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"constellation_id":  map[string]any{"type": "keyword"},
				"user_id":           map[string]any{"type": "keyword"},
				"theme_code":        map[string]any{"type": "keyword"},
				"trace_count":       map[string]any{"type": "float"},
				"moment_count":      map[string]any{"type": "float"},
				"updated_at":        map[string]any{"type": "date"},
				"topic":             textMapping(),
				"summary":           textMapping(),
				"keywords":          textMapping(),
				"emotions":          textMapping(),
				"scenes":            textMapping(),
				"central_pattern":   textMapping(),
				"pattern_tags":      textMapping(),
				"theme_label":       textMapping(),
				"theme_description": textMapping(),
				"theme_examples":    textMapping(),
				"profile_text":      textMapping(),
			},
		},
	}
	return s.client.DoJSON(ctx, http.MethodPut, "/"+s.index, body, nil)
}

func (s *ConstellationProfileSearch) IndexProfile(ctx context.Context, profile domain.ConstellationProfile) error {
	if s == nil || s.client == nil {
		return nil
	}
	if err := s.ensure(ctx); err != nil {
		return err
	}
	doc := toConstellationProfileDocument(profile)
	if doc.ConstellationID == "" {
		return fmt.Errorf("constellation_id is required")
	}
	path := fmt.Sprintf("/%s/_doc/%s?refresh=false", s.index, doc.ConstellationID)
	return s.client.DoJSON(ctx, http.MethodPut, path, doc, nil)
}

func (s *ConstellationProfileSearch) SearchCandidates(ctx context.Context, profile domain.TraceProfile, limit int) ([]domain.ConstellationProfileSparseCandidate, error) {
	if s == nil || s.client == nil || limit <= 0 {
		return nil, nil
	}
	if err := s.ensure(ctx); err != nil {
		return nil, err
	}
	queryText := constellationSparseQueryText(profile)
	if strings.TrimSpace(queryText) == "" {
		return nil, nil
	}
	query := map[string]any{
		"size": limit,
		"_source": []string{
			"constellation_id",
			"topic",
			"theme_code",
			"theme_label",
			"keywords",
			"scenes",
			"pattern_tags",
		},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{"term": map[string]any{"user_id": profile.UserID}},
				},
				"should": []any{
					multiMatch(queryText, "best_fields", []string{
						"topic^3",
						"theme_label^3",
						"keywords^2.5",
						"pattern_tags^2.5",
						"scenes^2",
						"summary^1.5",
						"central_pattern^1.5",
						"theme_description^1.5",
						"theme_examples",
						"profile_text",
					}),
					multiMatch(queryText, "most_fields", []string{
						"topic.ngram^0.6",
						"keywords.ngram^0.5",
						"pattern_tags.ngram^0.5",
						"scenes.ngram^0.4",
						"profile_text.ngram^0.2",
					}),
				},
				"minimum_should_match": 1,
			},
		},
	}

	var res constellationProfileSearchResponse
	if err := s.client.DoJSON(ctx, http.MethodPost, "/"+s.index+"/_search", query, &res); err != nil {
		return nil, err
	}
	result := make([]domain.ConstellationProfileSparseCandidate, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		id := hit.Source.ConstellationID
		if id == "" {
			id = hit.ID
		}
		if strings.TrimSpace(id) == "" {
			continue
		}
		result = append(result, domain.ConstellationProfileSparseCandidate{
			ConstellationID: id,
			Score:           hit.Score,
			MatchedFields:   matchedSparseFields(profile, hit.Source),
			Preview:         hit.Source.preview(),
		})
	}
	return result, nil
}

func (s *ConstellationProfileSearch) ensure(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ensured {
		return nil
	}
	if err := s.EnsureIndex(ctx); err != nil {
		return err
	}
	s.ensured = true
	return nil
}

func textMapping() map[string]any {
	return map[string]any{
		"type":            "text",
		"analyzer":        "ego_ik_index",
		"search_analyzer": "ego_ik_search",
		"fields": map[string]any{
			"ngram": map[string]any{
				"type":     "text",
				"analyzer": "ego_char_bigram",
			},
		},
	}
}

func multiMatch(query string, matchType string, fields []string) map[string]any {
	return map[string]any{"multi_match": map[string]any{
		"query":  query,
		"type":   matchType,
		"fields": fields,
	}}
}

type constellationProfileDocument struct {
	ConstellationID  string    `json:"constellation_id"`
	UserID           string    `json:"user_id"`
	Topic            string    `json:"topic"`
	Summary          string    `json:"summary"`
	Keywords         string    `json:"keywords"`
	Emotions         string    `json:"emotions"`
	Scenes           string    `json:"scenes"`
	CentralPattern   string    `json:"central_pattern"`
	PatternTags      string    `json:"pattern_tags"`
	ThemeCode        string    `json:"theme_code"`
	ThemeLabel       string    `json:"theme_label"`
	ThemeDescription string    `json:"theme_description"`
	ThemeExamples    string    `json:"theme_examples"`
	ProfileText      string    `json:"profile_text"`
	TraceCount       float64   `json:"trace_count"`
	MomentCount      float64   `json:"moment_count"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func toConstellationProfileDocument(profile domain.ConstellationProfile) constellationProfileDocument {
	return constellationProfileDocument{
		ConstellationID:  profile.ConstellationID,
		UserID:           profile.UserID,
		Topic:            profile.Topic,
		Summary:          profile.Summary,
		Keywords:         strings.Join(profile.Keywords, " "),
		Emotions:         strings.Join(profile.Emotions, " "),
		Scenes:           strings.Join(profile.Scenes, " "),
		CentralPattern:   profile.CentralPattern,
		PatternTags:      strings.Join(profile.PatternTags, " "),
		ThemeCode:        profile.ThemeCode,
		ThemeLabel:       profile.ThemeLabel,
		ThemeDescription: profile.ThemeDescription,
		ThemeExamples:    strings.Join(profile.ThemeExamples, " "),
		ProfileText:      profile.ProfileText,
		TraceCount:       profile.TraceCount,
		MomentCount:      profile.MomentCount,
		UpdatedAt:        profile.UpdatedAt,
	}
}

func constellationSparseQueryText(profile domain.TraceProfile) string {
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

func matchedSparseFields(profile domain.TraceProfile, doc constellationProfileDocument) []string {
	var fields []string
	if containsAny(profile.Keywords, doc.Keywords) {
		fields = append(fields, "keywords")
	}
	if containsAny(profile.Scenes, doc.Scenes) {
		fields = append(fields, "scenes")
	}
	if containsAny(profile.PatternTags, doc.PatternTags) {
		fields = append(fields, "pattern_tags")
	}
	if strings.TrimSpace(profile.Topic) != "" && strings.Contains(doc.Topic, strings.TrimSpace(profile.Topic)) {
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

func (d constellationProfileDocument) preview() string {
	for _, value := range []string{d.Topic, d.ThemeLabel, d.Summary, d.ProfileText} {
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

type constellationProfileSearchResponse struct {
	Hits struct {
		Hits []struct {
			ID     string                       `json:"_id"`
			Score  float64                      `json:"_score"`
			Source constellationProfileDocument `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

var _ domain.ConstellationProfileSearchIndexer = (*ConstellationProfileSearch)(nil)
var _ domain.ConstellationProfileSparseCandidateReader = (*ConstellationProfileSearch)(nil)
