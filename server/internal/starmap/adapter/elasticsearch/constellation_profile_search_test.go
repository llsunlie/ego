package elasticsearch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	platformes "ego-server/internal/platform/elasticsearch"
	"ego-server/internal/starmap/domain"
)

func TestConstellationProfileSearchSearchCandidatesBuildsSparseQuery(t *testing.T) {
	var gotQuery map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/"+DefaultConstellationProfileIndex:
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/"+DefaultConstellationProfileIndex+"/_search":
			if err := json.NewDecoder(r.Body).Decode(&gotQuery); err != nil {
				t.Fatalf("decode query: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"hits": map[string]any{
					"hits": []map[string]any{
						{
							"_id":    "c-1",
							"_score": 3.2,
							"_source": map[string]any{
								"constellation_id": "c-1",
								"topic":            "入职等待",
								"theme_code":       "theme_onboarding",
								"keywords":         "入职 反馈",
								"scenes":           "工作",
								"pattern_tags":     "等待反馈",
							},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	search := NewConstellationProfileSearch(platformes.NewClient(platformes.Config{URL: server.URL}, nil), DefaultConstellationProfileIndex)
	got, err := search.SearchCandidates(t.Context(), domain.TraceProfile{
		UserID:      "user-1",
		Topic:       "入职等待",
		Summary:     "用户在等入职反馈。",
		Keywords:    []string{"入职", "反馈"},
		Scenes:      []string{"工作"},
		PatternTags: []string{"等待反馈"},
		ProfileText: "主题：入职等待",
	}, 10)
	if err != nil {
		t.Fatalf("SearchCandidates() error = %v", err)
	}
	if len(got) != 1 || got[0].ConstellationID != "c-1" {
		t.Fatalf("got = %#v, want c-1", got)
	}
	if gotQuery["size"] != float64(10) {
		t.Fatalf("size = %#v, want 10", gotQuery["size"])
	}
	queryJSON, _ := json.Marshal(gotQuery)
	for _, want := range []string{"user-1", "topic^3", "pattern_tags^2.5", "profile_text.ngram^0.2"} {
		if !strings.Contains(string(queryJSON), want) {
			t.Fatalf("query missing %q: %s", want, queryJSON)
		}
	}
}

func TestConstellationProfileSearchIndexProfile(t *testing.T) {
	var indexed map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/"+DefaultConstellationProfileIndex:
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPut && r.URL.Path == "/"+DefaultConstellationProfileIndex+"/_doc/c-1":
			if err := json.NewDecoder(r.Body).Decode(&indexed); err != nil {
				t.Fatalf("decode index body: %v", err)
			}
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	search := NewConstellationProfileSearch(platformes.NewClient(platformes.Config{URL: server.URL}, nil), DefaultConstellationProfileIndex)
	err := search.IndexProfile(t.Context(), domain.ConstellationProfile{
		ConstellationID: "c-1",
		UserID:          "user-1",
		Topic:           "入职等待",
		Keywords:        []string{"入职", "反馈"},
		Scenes:          []string{"工作"},
		PatternTags:     []string{"等待反馈"},
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		t.Fatalf("IndexProfile() error = %v", err)
	}
	if indexed["keywords"] != "入职 反馈" {
		t.Fatalf("keywords = %#v, want joined keywords", indexed["keywords"])
	}
	if indexed["pattern_tags"] != "等待反馈" {
		t.Fatalf("pattern_tags = %#v, want joined tags", indexed["pattern_tags"])
	}
}
