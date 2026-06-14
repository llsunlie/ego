package elasticsearch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	platformes "ego-server/internal/platform/elasticsearch"
	"ego-server/internal/writing/domain"
)

func TestEnsureIndexCreatesMissingIndex(t *testing.T) {
	var putSeen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/ego_moments":
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPut && r.URL.Path == "/ego_moments":
			putSeen = true
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["settings"] == nil || body["mappings"] == nil {
				t.Fatalf("expected settings and mappings in create index body: %#v", body)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"acknowledged":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	search := NewMomentSearch(platformes.NewClient(platformes.Config{URL: server.URL}, nil), DefaultMomentIndex)
	if err := search.EnsureIndex(t.Context()); err != nil {
		t.Fatalf("EnsureIndex() error = %v", err)
	}
	if !putSeen {
		t.Fatal("expected missing index to be created")
	}
}

func TestSearchMomentIDsBuildsSparseQuery(t *testing.T) {
	current := domain.Moment{
		ID:        "moment-current",
		UserID:    "user-1",
		TraceID:   "trace-current",
		Content:   "每次都是我先开口，我真的有点累。",
		CreatedAt: time.Now(),
	}
	var query map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/ego_moments":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/ego_moments/_search":
			if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
				t.Fatalf("decode query: %v", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"hits":{"hits":[{"_id":"fallback-id","_source":{"moment_id":"moment-a"}},{"_id":"moment-b","_source":{}}]}}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	search := NewMomentSearch(platformes.NewClient(platformes.Config{URL: server.URL}, nil), DefaultMomentIndex)
	ids, err := search.SearchMomentIDs(t.Context(), current, 10)
	if err != nil {
		t.Fatalf("SearchMomentIDs() error = %v", err)
	}
	if got, want := strings.Join(ids, ","), "moment-a,moment-b"; got != want {
		t.Fatalf("ids = %q, want %q", got, want)
	}

	raw, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("marshal captured query: %v", err)
	}
	queryText := string(raw)
	for _, want := range []string{`"user_id":"user-1"`, `"moment_id":"moment-current"`, `"trace_id":"trace-current"`, `"content.ngram"`} {
		if !strings.Contains(queryText, want) {
			t.Fatalf("query does not contain %s: %s", want, queryText)
		}
	}
}

func TestBulkIndexMomentsReportsBulkErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodHead && r.URL.Path == "/ego_moments":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/_bulk":
			if ct := r.Header.Get("Content-Type"); ct != "application/x-ndjson" {
				t.Fatalf("content type = %q", ct)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"errors":true}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	search := NewMomentSearch(platformes.NewClient(platformes.Config{URL: server.URL}, nil), DefaultMomentIndex)
	err := search.BulkIndexMoments(t.Context(), []domain.Moment{{
		ID:        "moment-1",
		UserID:    "user-1",
		TraceID:   "trace-1",
		Content:   "hello",
		CreatedAt: time.Now(),
	}})
	if err == nil || !strings.Contains(err.Error(), "bulk index contains errors") {
		t.Fatalf("BulkIndexMoments() error = %v, want bulk error", err)
	}
}
