package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	platformes "ego-server/internal/platform/elasticsearch"
	"ego-server/internal/writing/domain"
)

const DefaultMomentIndex = "ego_moments"

type MomentSearch struct {
	client  *platformes.Client
	index   string
	mu      sync.Mutex
	ensured bool
}

func NewMomentSearch(client *platformes.Client, index string) *MomentSearch {
	if index == "" {
		index = DefaultMomentIndex
	}
	return &MomentSearch{client: client, index: index}
}

func (s *MomentSearch) EnsureIndex(ctx context.Context) error {
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
				"moment_id":  map[string]any{"type": "keyword"},
				"user_id":    map[string]any{"type": "keyword"},
				"trace_id":   map[string]any{"type": "keyword"},
				"created_at": map[string]any{"type": "date"},
				"content": map[string]any{
					"type":            "text",
					"analyzer":        "ego_ik_index",
					"search_analyzer": "ego_ik_search",
					"fields": map[string]any{
						"ngram": map[string]any{
							"type":     "text",
							"analyzer": "ego_char_bigram",
						},
					},
				},
			},
		},
	}
	return s.client.DoJSON(ctx, http.MethodPut, "/"+s.index, body, nil)
}

func (s *MomentSearch) IndexMoment(ctx context.Context, moment domain.Moment) error {
	if s == nil || s.client == nil {
		return nil
	}
	if err := s.ensure(ctx); err != nil {
		return err
	}
	doc := toMomentDocument(moment)
	if doc.MomentID == "" {
		return fmt.Errorf("moment_id is required")
	}
	path := fmt.Sprintf("/%s/_doc/%s?refresh=false", s.index, doc.MomentID)
	return s.client.DoJSON(ctx, http.MethodPut, path, doc, nil)
}

func (s *MomentSearch) BulkIndexMoments(ctx context.Context, moments []domain.Moment) error {
	if s == nil || s.client == nil || len(moments) == 0 {
		return nil
	}
	if err := s.ensure(ctx); err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, moment := range moments {
		doc := toMomentDocument(moment)
		if doc.MomentID == "" {
			continue
		}
		if err := enc.Encode(map[string]any{"index": map[string]any{"_index": s.index, "_id": doc.MomentID}}); err != nil {
			return err
		}
		if err := enc.Encode(doc); err != nil {
			return err
		}
	}
	var res bulkResponse
	if err := s.client.DoNDJSON(ctx, http.MethodPost, "/_bulk?refresh=false", buf.Bytes(), &res); err != nil {
		return err
	}
	if res.Errors {
		return fmt.Errorf("bulk index contains errors")
	}
	return nil
}

func (s *MomentSearch) SearchMomentIDs(ctx context.Context, current domain.Moment, limit int32) ([]string, error) {
	if s == nil || s.client == nil || limit <= 0 {
		return nil, nil
	}
	if err := s.ensure(ctx); err != nil {
		return nil, err
	}
	query := map[string]any{
		"size": limit,
		"_source": []string{
			"moment_id",
		},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{"term": map[string]any{"user_id": current.UserID}},
				},
				"must_not": []any{
					map[string]any{"term": map[string]any{"moment_id": current.ID}},
					map[string]any{"term": map[string]any{"trace_id": current.TraceID}},
				},
				"should": []any{
					map[string]any{"match": map[string]any{"content": map[string]any{
						"query": current.Content,
						"boost": 1.0,
					}}},
					map[string]any{"match": map[string]any{"content.ngram": map[string]any{
						"query": current.Content,
						"boost": 0.3,
					}}},
				},
				"minimum_should_match": 1,
			},
		},
	}

	var res searchResponse
	if err := s.client.DoJSON(ctx, http.MethodPost, "/"+s.index+"/_search", query, &res); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		id := hit.Source.MomentID
		if id == "" {
			id = hit.ID
		}
		if strings.TrimSpace(id) != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (s *MomentSearch) ensure(ctx context.Context) error {
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

type momentDocument struct {
	MomentID  string    `json:"moment_id"`
	UserID    string    `json:"user_id"`
	TraceID   string    `json:"trace_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func toMomentDocument(moment domain.Moment) momentDocument {
	return momentDocument{
		MomentID:  moment.ID,
		UserID:    moment.UserID,
		TraceID:   moment.TraceID,
		Content:   moment.Content,
		CreatedAt: moment.CreatedAt,
	}
}

type searchResponse struct {
	Hits struct {
		Hits []struct {
			ID     string         `json:"_id"`
			Source momentDocument `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type bulkResponse struct {
	Errors bool `json:"errors"`
}

var _ domain.MomentSearchIndexer = (*MomentSearch)(nil)
var _ domain.EchoSparseCandidateReader = (*MomentSearch)(nil)
