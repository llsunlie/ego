package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"ego-server/internal/writing/app"
	"ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"

	"github.com/google/uuid"
)

// ============================================================================
// Stubs — minimal no-op implementations for handler construction
// ============================================================================

func newSmokeHandler(
	traces domain.TraceRepository,
	moments domain.MomentRepository,
	echos domain.EchoRepository,
	embedding domain.EmbeddingGenerator,
	echoMatcher domain.EchoMatcher,
	insights domain.InsightRepository,
	insightGen domain.InsightGenerator,
) *Handler {
	createMoment := app.NewCreateMomentUseCase(traces, moments, echos, embedding, echoMatcher, uuidgen{})
	generateInsight := app.NewGenerateInsightUseCase(insights, insightGen, uuidgen{})

	return NewHandler(createMoment, generateInsight, nil)
}

type uuidgen struct{}

func (uuidgen) New() string { return uuid.NewString() }

func userCtx(userID string) context.Context {
	return context.WithValue(context.Background(), "user_id", userID)
}

// ============================================================================
// Smoke: F1 主流程 — 写字 → 回声 → 观察
// ============================================================================

func TestSmoke_F1_WriteAndObserve(t *testing.T) {
	userID := "user-f1"
	momentsStore := make(map[string]domain.Moment)
	tracesStore := make(map[string]domain.Trace)

	traces := &statefulTraceRepo{store: tracesStore}
	moments := &statefulMomentRepo{store: momentsStore}
	echos := &statefulEchoRepo{}
	embedding := &fixedEmbeddingGen{}
	echoMatcher := &fixedEchoMatcher{}
	insightRepo := &statefulInsightRepo{}
	insightGen := &fixedInsightGen{}

	// Seed one old moment so echo matching finds history
	oldMoment := domain.Moment{
		ID:        uuid.NewString(),
		TraceID:   uuid.NewString(),
		UserID:    userID,
		Content:   "上周和同事发生了争执",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}
	// Create old trace first
	traces.store[oldMoment.TraceID] = domain.Trace{
		ID: oldMoment.TraceID, UserID: userID, Motivation: "direct", Stashed: true,
	}
	moments.store[oldMoment.ID] = oldMoment

	h := newSmokeHandler(traces, moments, echos, embedding, echoMatcher, insightRepo, insightGen)

	// Step 1: Write a moment
	res1, err := h.CreateMoment(userCtx(userID), &pb.CreateMomentReq{Content: "今天和同事起了冲突"})
	if err != nil {
		t.Fatalf("CreateMoment: %v", err)
	}
	if res1.Moment == nil {
		t.Fatal("expected non-nil Moment")
	}
	if res1.Moment.Content != "今天和同事起了冲突" {
		t.Fatalf("unexpected content: %q", res1.Moment.Content)
	}
	if res1.Echo == nil {
		t.Fatal("expected non-nil Echo (history exists)")
	}
	if len(res1.Echo.MatchedMomentIds) != 1 {
		t.Fatalf("expected 1 matched moment, got %d", len(res1.Echo.MatchedMomentIds))
	}
	if res1.Echo.MatchedMomentIds[0] != oldMoment.ID {
		t.Fatalf("echo should match old moment %s, got %s", oldMoment.ID, res1.Echo.MatchedMomentIds[0])
	}

	// Step 2: Generate insight from that moment + echo
	res2, err := h.GenerateInsight(userCtx(userID), &pb.GenerateInsightReq{
		MomentId: res1.Moment.Id,
		EchoId:   res1.Echo.Id,
	})
	if err != nil {
		t.Fatalf("GenerateInsight: %v", err)
	}
	if res2.Insight == nil {
		t.Fatal("expected non-nil Insight")
	}
	if res2.Insight.MomentId != res1.Moment.Id {
		t.Fatalf("Insight.MomentId mismatch: expected %s, got %s", res1.Moment.Id, res2.Insight.MomentId)
	}
	if res2.Insight.EchoId != res1.Echo.Id {
		t.Fatalf("Insight.EchoId mismatch: expected %s, got %s", res1.Echo.Id, res2.Insight.EchoId)
	}
	if res2.Insight.Text == "" {
		t.Fatal("expected non-empty insight text")
	}
}

// ============================================================================
// Smoke: F2 顺着再想想 — 深度探索循环
// ============================================================================

func TestSmoke_F2_ContinueTrace(t *testing.T) {
	userID := "user-f2"
	momentsStore := make(map[string]domain.Moment)
	tracesStore := make(map[string]domain.Trace)

	traces := &statefulTraceRepo{store: tracesStore}
	moments := &statefulMomentRepo{store: momentsStore}
	echos := &statefulEchoRepo{}
	embedding := &fixedEmbeddingGen{}
	echoMatcher := &fixedEchoMatcher{}
	insightRepo := &statefulInsightRepo{}
	insightGen := &fixedInsightGen{}

	h := newSmokeHandler(traces, moments, echos, embedding, echoMatcher, insightRepo, insightGen)

	// Round 1: New trace (no trace_id)
	res1, err := h.CreateMoment(userCtx(userID), &pb.CreateMomentReq{Content: "今天和同事起了冲突"})
	if err != nil {
		t.Fatalf("CreateMoment round 1: %v", err)
	}
	traceID := res1.Moment.TraceId
	if traceID == "" {
		t.Fatal("expected non-empty trace_id for new trace")
	}

	// Round 2: Continue same trace ("顺着再想想")
	res2, err := h.CreateMoment(userCtx(userID), &pb.CreateMomentReq{
		Content: "其实是我害怕被否定",
		TraceId: traceID,
	})
	if err != nil {
		t.Fatalf("CreateMoment round 2: %v", err)
	}
	if res2.Moment.TraceId != traceID {
		t.Fatalf("expected same trace_id %s, got %s", traceID, res2.Moment.TraceId)
	}
	if res2.Moment.Id == res1.Moment.Id {
		t.Fatal("expected different moment IDs across rounds")
	}

	// Round 3: Continue again
	res3, err := h.CreateMoment(userCtx(userID), &pb.CreateMomentReq{
		Content: "小时候也是这样",
		TraceId: traceID,
	})
	if err != nil {
		t.Fatalf("CreateMoment round 3: %v", err)
	}
	if res3.Moment.TraceId != traceID {
		t.Fatalf("expected same trace_id %s, got %s", traceID, res3.Moment.TraceId)
	}

	// Verify trace has 3 moments
	traceMoments, err := moments.ListByTraceID(context.Background(), traceID)
	if err != nil {
		t.Fatalf("ListByTraceID: %v", err)
	}
	if len(traceMoments) != 3 {
		t.Fatalf("expected 3 moments under trace %s, got %d", traceID, len(traceMoments))
	}

	// Round 1: cold start, no echo
	echo1, _ := echos.FindByMomentID(context.Background(), res1.Moment.Id)
	if echo1 != nil {
		t.Fatal("round 1: expected nil echo (cold start)")
	}

	// Rounds 2 & 3: echo should be persisted from prior rounds' history
	for i, m := range []*pb.Moment{res2.Moment, res3.Moment} {
		echoRes, _ := echos.FindByMomentID(context.Background(), m.Id)
		if echoRes == nil {
			t.Fatalf("round %d: expected echo to be persisted", i+2)
		}
		resInsight, err := h.GenerateInsight(userCtx(userID), &pb.GenerateInsightReq{
			MomentId: m.Id,
			EchoId:   echoRes.ID,
		})
		if err != nil {
			t.Fatalf("round %d GenerateInsight: %v", i+2, err)
		}
		if resInsight.Insight == nil {
			t.Fatalf("round %d: expected non-nil insight", i+2)
		}
	}
}

// ============================================================================
// Smoke: F9 冷启动 — 用户首次使用，无历史数据
// ============================================================================

func TestSmoke_F9_ColdStart(t *testing.T) {
	userID := "user-f9"
	momentsStore := make(map[string]domain.Moment)
	tracesStore := make(map[string]domain.Trace)

	traces := &statefulTraceRepo{store: tracesStore}
	moments := &statefulMomentRepo{store: momentsStore}
	echos := &statefulEchoRepo{}
	embedding := &fixedEmbeddingGen{}
	echoMatcher := &fixedEchoMatcher{}
	insightRepo := &statefulInsightRepo{}
	insightGen := &fixedInsightGen{}

	h := newSmokeHandler(traces, moments, echos, embedding, echoMatcher, insightRepo, insightGen)

	// First moment ever — no history
	res, err := h.CreateMoment(userCtx(userID), &pb.CreateMomentReq{Content: "第一次在这里说话"})
	if err != nil {
		t.Fatalf("CreateMoment: %v", err)
	}
	if res.Moment == nil {
		t.Fatal("expected non-nil Moment")
	}
	if res.Echo != nil {
		t.Fatal("expected nil Echo for first-time user (no history)")
	}
	if res.Moment.TraceId == "" {
		t.Fatal("expected trace_id to be created automatically")
	}

	// Verify trace was created with motivation "direct"
	tr, _ := traces.GetByID(context.Background(), res.Moment.TraceId)
	if tr == nil {
		t.Fatal("trace should exist")
	}
	if tr.Motivation != "direct" {
		t.Fatalf("expected Motivation 'direct', got %q", tr.Motivation)
	}
	if tr.Stashed {
		t.Fatal("expected Stashed to be false for new trace")
	}
}

// ============================================================================
// Handler: basic delegation & error propagation
// ============================================================================

func TestHandler_CreateMoment_Delegation(t *testing.T) {
	h := NewHandler(nil, nil, nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.createMoment != nil || h.generateInsight != nil {
		t.Fatal("expected nil use cases from constructor")
	}
}

func TestHandler_CreateMoment_ErrorPropagation(t *testing.T) {
	uc := app.NewCreateMomentUseCase(nil, nil, nil, nil, nil, &stubIDGen{})
	h := NewHandler(uc, nil, nil)

	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	_, err := h.CreateMoment(ctx, &pb.CreateMomentReq{Content: ""})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestHandler_GenerateInsight_ErrorPropagation(t *testing.T) {
	uc := app.NewGenerateInsightUseCase(nil, nil, &stubIDGen{})
	h := NewHandler(nil, uc, nil)

	_, err := h.GenerateInsight(context.Background(), &pb.GenerateInsightReq{MomentId: ""})
	if err == nil {
		t.Fatal("expected error for empty MomentId")
	}
}

func TestHandler_GenerateInsight_Success(t *testing.T) {
	uc := app.NewGenerateInsightUseCase(
		&stubHandlerInsightRepo{},
		&stubHandlerInsightGenerator{},
		&stubIDGen{},
	)
	h := NewHandler(nil, uc, nil)

	res, err := h.GenerateInsight(context.Background(), &pb.GenerateInsightReq{
		MomentId: "mom-1",
		EchoId:   "echo-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Insight == nil {
		t.Fatal("expected non-nil insight")
	}
	if res.Insight.Text != "You are making progress." {
		t.Fatalf("unexpected text: %q", res.Insight.Text)
	}
}

func TestHandler_GetMoments_Success(t *testing.T) {
	reader := &stubMomentReader{moments: map[string]domain.Moment{
		"mom-1": {ID: "mom-1", Content: "hello", TraceID: "tr-1", UserID: "u-1"},
		"mom-2": {ID: "mom-2", Content: "world", TraceID: "tr-1", UserID: "u-1"},
	}}
	h := NewHandler(nil, nil, reader)

	res, err := h.GetMoments(context.Background(), &pb.GetMomentsReq{Ids: []string{"mom-1", "mom-2"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Moments) != 2 {
		t.Fatalf("expected 2 moments, got %d", len(res.Moments))
	}
	if res.Moments[0].Content != "hello" {
		t.Fatalf("expected 'hello', got %q", res.Moments[0].Content)
	}
	if res.Moments[1].Content != "world" {
		t.Fatalf("expected 'world', got %q", res.Moments[1].Content)
	}
}

func TestHandler_GetMoments_Empty(t *testing.T) {
	reader := &stubMomentReader{moments: map[string]domain.Moment{}}
	h := NewHandler(nil, nil, reader)

	res, err := h.GetMoments(context.Background(), &pb.GetMomentsReq{Ids: []string{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Moments) != 0 {
		t.Fatalf("expected 0 moments, got %d", len(res.Moments))
	}
}

func TestHandler_GetMoments_PartialMatch(t *testing.T) {
	reader := &stubMomentReader{moments: map[string]domain.Moment{
		"mom-1": {ID: "mom-1", Content: "exists", TraceID: "tr-1", UserID: "u-1"},
	}}
	h := NewHandler(nil, nil, reader)

	res, err := h.GetMoments(context.Background(), &pb.GetMomentsReq{Ids: []string{"mom-1", "nonexistent"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Moments) != 1 {
		t.Fatalf("expected 1 moment, got %d", len(res.Moments))
	}
	if res.Moments[0].Content != "exists" {
		t.Fatalf("expected 'exists', got %q", res.Moments[0].Content)
	}
}

// ============================================================================
// Stateful mocks — behave like real repositories in memory
// ============================================================================

type stubIDGen struct{}

func (s *stubIDGen) New() string { return uuid.NewString() }

type stubHandlerInsightRepo struct{}

func (s *stubHandlerInsightRepo) Create(ctx context.Context, insight *domain.Insight) error {
	return nil
}
func (s *stubHandlerInsightRepo) FindByMomentID(ctx context.Context, momentID string) (*domain.Insight, error) {
	return nil, nil
}

type stubHandlerInsightGenerator struct{}

func (s *stubHandlerInsightGenerator) Generate(ctx context.Context, momentID string, echoID string) (*domain.Insight, error) {
	if momentID == "" {
		return nil, errors.New("empty momentID")
	}
	return &domain.Insight{
		ID:               "ins-1",
		MomentID:         momentID,
		EchoID:           echoID,
		Text:             "You are making progress.",
		RelatedMomentIDs: []string{"echo-1"},
	}, nil
}

// --- MomentReader stub for handler tests ---

type stubMomentReader struct {
	moments map[string]domain.Moment
}

func (r *stubMomentReader) GetByID(ctx context.Context, id string) (*domain.Moment, error) {
	m, ok := r.moments[id]
	if !ok {
		return nil, domain.ErrMomentNotFound
	}
	return &m, nil
}

func (r *stubMomentReader) GetByIDs(ctx context.Context, ids []string) ([]domain.Moment, error) {
	var result []domain.Moment
	for _, id := range ids {
		if m, ok := r.moments[id]; ok {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *stubMomentReader) ListByUserID(ctx context.Context, userID string, cursor string, pageSize int32) ([]domain.Moment, string, bool, error) {
	return nil, "", false, nil
}

func (r *stubMomentReader) RandomByUserID(ctx context.Context, userID string, count int32) ([]domain.Moment, error) {
	return nil, nil
}

// --- Stateful repos for smoke tests ---

type statefulTraceRepo struct {
	store map[string]domain.Trace
}

func (r *statefulTraceRepo) Create(ctx context.Context, trace *domain.Trace) error {
	r.store[trace.ID] = *trace
	return nil
}
func (r *statefulTraceRepo) GetByID(ctx context.Context, id string) (*domain.Trace, error) {
	tr, ok := r.store[id]
	if !ok {
		return nil, domain.ErrTraceNotFound
	}
	return &tr, nil
}
func (r *statefulTraceRepo) Update(ctx context.Context, trace *domain.Trace) error {
	r.store[trace.ID] = *trace
	return nil
}
func (r *statefulTraceRepo) Delete(ctx context.Context, id string) error {
	delete(r.store, id)
	return nil
}

type statefulMomentRepo struct {
	store map[string]domain.Moment
}

func (r *statefulMomentRepo) Create(ctx context.Context, moment *domain.Moment) error {
	r.store[moment.ID] = *moment
	return nil
}
func (r *statefulMomentRepo) GetByID(ctx context.Context, id string) (*domain.Moment, error) {
	m, ok := r.store[id]
	if !ok {
		return nil, domain.ErrMomentNotFound
	}
	return &m, nil
}
func (r *statefulMomentRepo) ListByTraceID(ctx context.Context, traceID string) ([]domain.Moment, error) {
	var result []domain.Moment
	for _, m := range r.store {
		if m.TraceID == traceID {
			result = append(result, m)
		}
	}
	return result, nil
}
func (r *statefulMomentRepo) ListByUserID(ctx context.Context, userID string) ([]domain.Moment, error) {
	var result []domain.Moment
	for _, m := range r.store {
		if m.UserID == userID {
			result = append(result, m)
		}
	}
	return result, nil
}

type statefulEchoRepo struct {
	byMomentID map[string]*domain.Echo
}

func (r *statefulEchoRepo) ensure() {
	if r.byMomentID == nil {
		r.byMomentID = make(map[string]*domain.Echo)
	}
}

func (r *statefulEchoRepo) Create(ctx context.Context, echo *domain.Echo) error {
	r.ensure()
	r.byMomentID[echo.MomentID] = echo
	return nil
}
func (r *statefulEchoRepo) FindByMomentID(ctx context.Context, momentID string) (*domain.Echo, error) {
	r.ensure()
	e, ok := r.byMomentID[momentID]
	if !ok {
		return nil, domain.ErrEchoNotFound
	}
	return e, nil
}

type statefulInsightRepo struct {
	byMomentID map[string]*domain.Insight
}

func (r *statefulInsightRepo) ensure() {
	if r.byMomentID == nil {
		r.byMomentID = make(map[string]*domain.Insight)
	}
}

func (r *statefulInsightRepo) Create(ctx context.Context, insight *domain.Insight) error {
	r.ensure()
	r.byMomentID[insight.MomentID] = insight
	return nil
}
func (r *statefulInsightRepo) FindByMomentID(ctx context.Context, momentID string) (*domain.Insight, error) {
	r.ensure()
	i, ok := r.byMomentID[momentID]
	if !ok {
		return nil, domain.ErrInsightNotFound
	}
	return i, nil
}

type fixedEmbeddingGen struct{}

func (fixedEmbeddingGen) Generate(ctx context.Context, content string) ([]domain.EmbeddingEntry, error) {
	return []domain.EmbeddingEntry{{Model: "test", Embedding: []float32{0.1}}}, nil
}

type fixedEchoMatcher struct{}

func (fixedEchoMatcher) Match(ctx context.Context, current *domain.Moment, history []domain.Moment) ([]domain.MatchedMoment, error) {
	if len(history) == 0 {
		return nil, nil
	}
	var matches []domain.MatchedMoment
	for _, h := range history {
		matches = append(matches, domain.MatchedMoment{MomentID: h.ID, Similarity: 0.8})
	}
	return matches, nil
}

type fixedInsightGen struct{}

func (fixedInsightGen) Generate(ctx context.Context, momentID, echoID string) (*domain.Insight, error) {
	return &domain.Insight{
		MomentID:         momentID,
		EchoID:           echoID,
		Text:             "你似乎在重复类似的模式。",
		RelatedMomentIDs: []string{echoID},
	}, nil
}
