package grpc

import (
	"context"
	"testing"
	"time"

	"ego-server/internal/timeline/domain"
	writingdomain "ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"

	"github.com/google/uuid"
)

func userCtx(userID string) context.Context {
	return context.WithValue(context.Background(), "user_id", userID)
}

// ============================================================================
// Fake readers
// ============================================================================

type fakeTraceReader struct {
	traceByID      map[string]*writingdomain.Trace
	tracesByUser   map[string][]writingdomain.Trace
	momentsByTrace map[string][]writingdomain.Moment
}

func (r *fakeTraceReader) GetTraceByID(ctx context.Context, id string) (*writingdomain.Trace, error) {
	tr, ok := r.traceByID[id]
	if !ok {
		return nil, writingdomain.ErrTraceNotFound
	}
	return tr, nil
}

func (r *fakeTraceReader) ListMomentsByTraceID(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
	return r.momentsByTrace[traceID], nil
}

func (r *fakeTraceReader) ListTracesByUserID(ctx context.Context, userID string, cursor string, pageSize int32) ([]writingdomain.Trace, string, bool, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	traces := r.tracesByUser[userID]

	start := 0
	if cursor != "" {
		for i, tr := range traces {
			if tr.ID == cursor {
				start = i + 1
				break
			}
		}
	}

	end := start + int(pageSize)
	hasMore := end < len(traces)
	if end > len(traces) {
		end = len(traces)
	}

	page := traces[start:end]
	nextCursor := ""
	if hasMore && len(page) > 0 {
		nextCursor = page[len(page)-1].ID
	}

	return page, nextCursor, hasMore, nil
}

type fakeMomentReader struct{}

func (fakeMomentReader) GetByID(ctx context.Context, id string) (*writingdomain.Moment, error) {
	return nil, writingdomain.ErrMomentNotFound
}
func (fakeMomentReader) ListByUserID(ctx context.Context, userID string, cursor string, pageSize int32) ([]writingdomain.Moment, string, bool, error) {
	return nil, "", false, nil
}
func (fakeMomentReader) RandomByUserID(ctx context.Context, userID string, count int32) ([]writingdomain.Moment, error) {
	if count <= 0 {
		count = 3
	}
	moments := make([]writingdomain.Moment, count)
	for i := range int(count) {
		moments[i] = writingdomain.Moment{
			ID:        uuid.NewString(),
			UserID:    userID,
			Content:   "random moment " + uuid.NewString()[:8],
			TraceID:   "tr-rand",
			CreatedAt: time.Now(),
		}
	}
	return moments, nil
}

type fakeEchoReader struct {
	byMomentID map[string]*writingdomain.Echo
}

func (r *fakeEchoReader) FindByMomentID(ctx context.Context, momentID string) (*writingdomain.Echo, error) {
	e, ok := r.byMomentID[momentID]
	if !ok {
		return nil, writingdomain.ErrEchoNotFound
	}
	return e, nil
}

type fakeInsightReader struct {
	byMomentID map[string]*writingdomain.Insight
}

func (r *fakeInsightReader) FindByMomentID(ctx context.Context, momentID string) (*writingdomain.Insight, error) {
	i, ok := r.byMomentID[momentID]
	if !ok {
		return nil, writingdomain.ErrInsightNotFound
	}
	return i, nil
}

// ============================================================================
// ListTraces
// ============================================================================

func TestHandler_ListTraces(t *testing.T) {
	userID := "user-lt"
	now := time.Now()

	reader := &fakeTraceReader{
		tracesByUser: map[string][]writingdomain.Trace{
			userID: {
				{ID: "tr-1", UserID: userID, Motivation: "direct", Stashed: false, CreatedAt: now},
				{ID: "tr-2", UserID: userID, Motivation: "constellation:c1", Stashed: true, CreatedAt: now.Add(-1 * time.Hour)},
			},
		},
	}

	h := NewHandler(fakeMomentReader{}, reader, &fakeEchoReader{}, &fakeInsightReader{})

	res, err := h.ListTraces(userCtx(userID), &pb.ListTracesReq{PageSize: 10})
	if err != nil {
		t.Fatalf("ListTraces: %v", err)
	}
	if len(res.Traces) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(res.Traces))
	}
	if res.Traces[0].Id != "tr-1" {
		t.Fatalf("expected first trace 'tr-1', got %q", res.Traces[0].Id)
	}
	if res.Traces[1].Motivation != "constellation:c1" {
		t.Fatalf("expected constellation motivation, got %q", res.Traces[1].Motivation)
	}
	if !res.Traces[1].Stashed {
		t.Fatal("expected tr-2 to be stashed")
	}
	if res.HasMore {
		t.Fatal("expected hasMore=false")
	}
}

func TestHandler_ListTraces_Pagination(t *testing.T) {
	userID := "user-ltp"
	now := time.Now()

	tracesData := make([]writingdomain.Trace, 5)
	for i := range 5 {
		tracesData[i] = writingdomain.Trace{
			ID: uuid.NewString(), UserID: userID, Motivation: "direct",
			Stashed: false, CreatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
	}

	reader := &fakeTraceReader{tracesByUser: map[string][]writingdomain.Trace{userID: tracesData}}
	h := NewHandler(fakeMomentReader{}, reader, &fakeEchoReader{}, &fakeInsightReader{})

	// Page 1
	res1, err := h.ListTraces(userCtx(userID), &pb.ListTracesReq{PageSize: 2})
	if err != nil {
		t.Fatalf("ListTraces page 1: %v", err)
	}
	if len(res1.Traces) != 2 {
		t.Fatalf("expected 2 traces on page 1, got %d", len(res1.Traces))
	}
	if !res1.HasMore {
		t.Fatal("expected hasMore=true on page 1")
	}

	// Page 2
	res2, err := h.ListTraces(userCtx(userID), &pb.ListTracesReq{Cursor: res1.NextCursor, PageSize: 2})
	if err != nil {
		t.Fatalf("ListTraces page 2: %v", err)
	}
	if len(res2.Traces) != 2 {
		t.Fatalf("expected 2 traces on page 2, got %d", len(res2.Traces))
	}
	if !res2.HasMore {
		t.Fatal("expected hasMore=true on page 2")
	}

	// Page 3 (last)
	res3, err := h.ListTraces(userCtx(userID), &pb.ListTracesReq{Cursor: res2.NextCursor, PageSize: 2})
	if err != nil {
		t.Fatalf("ListTraces page 3: %v", err)
	}
	if len(res3.Traces) != 1 {
		t.Fatalf("expected 1 trace on page 3, got %d", len(res3.Traces))
	}
	if res3.HasMore {
		t.Fatal("expected hasMore=false on last page")
	}
}

// ============================================================================
// GetTraceDetail
// ============================================================================

func TestHandler_GetTraceDetail(t *testing.T) {
	now := time.Now()
	traceID := "tr-detail-1"
	userID := "user-gtd"

	trace := writingdomain.Trace{ID: traceID, UserID: userID, Motivation: "direct", Stashed: false, CreatedAt: now}
	moments := []writingdomain.Moment{
		{ID: "mom-1", TraceID: traceID, UserID: userID, Content: "first message", CreatedAt: now},
		{ID: "mom-2", TraceID: traceID, UserID: userID, Content: "second message", CreatedAt: now.Add(time.Minute)},
	}
	echo := writingdomain.Echo{
		ID: "echo-1", MomentID: "mom-1", UserID: userID,
		MatchedMomentIDs: []string{"old-1"}, Similarities: []float64{0.9},
	}
	insight := writingdomain.Insight{
		ID: "ins-1", UserID: userID, MomentID: "mom-1", EchoID: "echo-1",
		Text: "You seem to be revisiting old patterns.", RelatedMomentIDs: []string{"old-1"},
	}

	reader := &fakeTraceReader{
		traceByID:      map[string]*writingdomain.Trace{traceID: &trace},
		momentsByTrace: map[string][]writingdomain.Moment{traceID: moments},
	}
	echoReader := &fakeEchoReader{byMomentID: map[string]*writingdomain.Echo{"mom-1": &echo}}
	insightReader := &fakeInsightReader{byMomentID: map[string]*writingdomain.Insight{"mom-1": &insight}}

	h := NewHandler(fakeMomentReader{}, reader, echoReader, insightReader)

	res, err := h.GetTraceDetail(userCtx(userID), &pb.GetTraceDetailReq{TraceId: traceID})
	if err != nil {
		t.Fatalf("GetTraceDetail: %v", err)
	}
	if res.Trace.Id != traceID {
		t.Fatalf("expected trace ID %s, got %s", traceID, res.Trace.Id)
	}
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(res.Items))
	}

	// Item 1: has echo + insight
	if res.Items[0].Moment.Id != "mom-1" {
		t.Fatalf("expected first moment 'mom-1', got %q", res.Items[0].Moment.Id)
	}
	if len(res.Items[0].Echos) != 1 {
		t.Fatalf("expected 1 echo for mom-1, got %d", len(res.Items[0].Echos))
	}
	if res.Items[0].Insight == nil {
		t.Fatal("expected insight for mom-1")
	}

	// Item 2: no echo/insight
	if res.Items[1].Moment.Id != "mom-2" {
		t.Fatalf("expected second moment 'mom-2', got %q", res.Items[1].Moment.Id)
	}
	if len(res.Items[1].Echos) != 0 {
		t.Fatalf("expected 0 echos for mom-2, got %d", len(res.Items[1].Echos))
	}
	if res.Items[1].Insight != nil {
		t.Fatal("expected nil insight for mom-2")
	}
}

func TestHandler_GetTraceDetail_NotFound(t *testing.T) {
	reader := &fakeTraceReader{}
	h := NewHandler(fakeMomentReader{}, reader, &fakeEchoReader{}, &fakeInsightReader{})

	_, err := h.GetTraceDetail(userCtx("user-1"), &pb.GetTraceDetailReq{TraceId: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent trace")
	}
}

// ============================================================================
// GetRandomMoments
// ============================================================================

func TestHandler_GetRandomMoments(t *testing.T) {
	userID := "user-rand"

	h := NewHandler(fakeMomentReader{}, &fakeTraceReader{}, &fakeEchoReader{}, &fakeInsightReader{})

	res, err := h.GetRandomMoments(userCtx(userID), &pb.GetRandomMomentsReq{Count: 3})
	if err != nil {
		t.Fatalf("GetRandomMoments: %v", err)
	}
	if len(res.Moments) != 3 {
		t.Fatalf("expected 3 moments, got %d", len(res.Moments))
	}
	for _, m := range res.Moments {
		if m.Id == "" {
			t.Fatal("expected non-empty moment ID")
		}
		if m.TraceId == "" {
			t.Fatal("expected non-empty trace ID")
		}
	}
}

func TestHandler_GetRandomMoments_DefaultCount(t *testing.T) {
	userID := "user-rand-default"

	// Override RandomByUserID to return 3 (default) when count <= 0
	reader := &countingMomentReader{}

	h := NewHandler(reader, &fakeTraceReader{}, &fakeEchoReader{}, &fakeInsightReader{})

	res, err := h.GetRandomMoments(userCtx(userID), &pb.GetRandomMomentsReq{})
	if err != nil {
		t.Fatalf("GetRandomMoments: %v", err)
	}
	if len(res.Moments) != 3 {
		t.Fatalf("expected 3 moments (default), got %d", len(res.Moments))
	}
}

type countingMomentReader struct {
	count int32
}

func (r *countingMomentReader) GetByID(ctx context.Context, id string) (*writingdomain.Moment, error) {
	return nil, nil
}
func (r *countingMomentReader) ListByUserID(ctx context.Context, userID string, cursor string, pageSize int32) ([]writingdomain.Moment, string, bool, error) {
	return nil, "", false, nil
}
func (r *countingMomentReader) RandomByUserID(ctx context.Context, userID string, count int32) ([]writingdomain.Moment, error) {
	r.count = count
	moments := make([]writingdomain.Moment, count)
	for i := range int(count) {
		moments[i] = writingdomain.Moment{
			ID:      uuid.NewString(),
			TraceID: "tr-1",
		}
	}
	return moments, nil
}

var _ domain.MomentReader = (*fakeMomentReader)(nil)
var _ domain.MomentReader = (*countingMomentReader)(nil)
var _ domain.TraceReader = (*fakeTraceReader)(nil)
var _ domain.EchoReader = (*fakeEchoReader)(nil)
var _ domain.InsightReader = (*fakeInsightReader)(nil)
