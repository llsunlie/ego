package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

// --- mocks for stash_trace ---

type mockTraceReader struct {
	getByIDFn              func(ctx context.Context, id string) (*writingdomain.Trace, error)
	listMomentsByTraceIDFn func(ctx context.Context, traceID string) ([]writingdomain.Moment, error)
}

func (m *mockTraceReader) GetTraceByID(ctx context.Context, id string) (*writingdomain.Trace, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockTraceReader) ListMomentsByTraceID(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
	return m.listMomentsByTraceIDFn(ctx, traceID)
}

type mockTraceStasher struct {
	markStashedFn func(ctx context.Context, traceID string) error
}

func (m *mockTraceStasher) MarkStashed(ctx context.Context, traceID string) error {
	return m.markStashedFn(ctx, traceID)
}

type mockStarRepo struct {
	createFn          func(ctx context.Context, star *domain.Star) error
	findByTraceIDFn   func(ctx context.Context, traceID string) (*domain.Star, error)
	findByIDsFn       func(ctx context.Context, ids []string) ([]domain.Star, error)
	findAllByUserIDFn func(ctx context.Context, userID string) ([]domain.Star, error)
	updateTopicFn     func(ctx context.Context, starID string, topic string) error
}

func (m *mockStarRepo) Create(ctx context.Context, star *domain.Star) error {
	return m.createFn(ctx, star)
}
func (m *mockStarRepo) FindByTraceID(ctx context.Context, traceID string) (*domain.Star, error) {
	return m.findByTraceIDFn(ctx, traceID)
}
func (m *mockStarRepo) FindByIDs(ctx context.Context, ids []string) ([]domain.Star, error) {
	return m.findByIDsFn(ctx, ids)
}
func (m *mockStarRepo) FindAllByUserID(ctx context.Context, userID string) ([]domain.Star, error) {
	return m.findAllByUserIDFn(ctx, userID)
}
func (m *mockStarRepo) UpdateTopic(ctx context.Context, starID string, topic string) error {
	return m.updateTopicFn(ctx, starID, topic)
}

type mockConstellationRepo struct {
	createFn          func(ctx context.Context, c *domain.Constellation) error
	updateFn          func(ctx context.Context, c *domain.Constellation) error
	findAllByUserIDFn func(ctx context.Context, userID string) ([]domain.Constellation, error)
	findByIDFn        func(ctx context.Context, id string) (*domain.Constellation, error)
}

func (m *mockConstellationRepo) Create(ctx context.Context, c *domain.Constellation) error {
	return m.createFn(ctx, c)
}
func (m *mockConstellationRepo) Update(ctx context.Context, c *domain.Constellation) error {
	return m.updateFn(ctx, c)
}
func (m *mockConstellationRepo) FindAllByUserID(ctx context.Context, userID string) ([]domain.Constellation, error) {
	return m.findAllByUserIDFn(ctx, userID)
}
func (m *mockConstellationRepo) FindByID(ctx context.Context, id string) (*domain.Constellation, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockConstellationRepo) FindByStarID(ctx context.Context, starID string) (*domain.Constellation, error) {
	return nil, nil
}

type mockAssetGen struct {
	generateFn func(ctx context.Context, moments []writingdomain.Moment) (string, []float32, string, string, []string, error)
}

func (m *mockAssetGen) Generate(ctx context.Context, moments []writingdomain.Moment) (string, []float32, string, string, []string, error) {
	return m.generateFn(ctx, moments)
}

type mockTraceProfileGen struct {
	generateFn func(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error)
}

func (m *mockTraceProfileGen) Generate(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
	return m.generateFn(ctx, trace, moments)
}

type mockTraceProfileRepo struct {
	upsertFn func(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error
}

func (m *mockTraceProfileRepo) Upsert(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error {
	return m.upsertFn(ctx, profile, vector)
}

type mockConstellationProfileRepo struct {
	findCandidatesFn func(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error)
	upsertFn         func(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error
	addMembershipFn  func(ctx context.Context, membership domain.ConstellationMembership) error
}

func (m *mockConstellationProfileRepo) FindCandidates(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error) {
	return m.findCandidatesFn(ctx, userID, embedding, limit)
}
func (m *mockConstellationProfileRepo) Upsert(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
	return m.upsertFn(ctx, profile, vector)
}
func (m *mockConstellationProfileRepo) AddMembership(ctx context.Context, membership domain.ConstellationMembership) error {
	return m.addMembershipFn(ctx, membership)
}

type mockIDGen struct {
	id string
}

func (m *mockIDGen) New() string { return m.id }

type mockSeqIDGen struct {
	ids []string
}

func (m *mockSeqIDGen) New() string {
	id := m.ids[0]
	m.ids = m.ids[1:]
	return id
}

// --- tests ---

func TestStashTrace_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{
		ID: "tr-1", UserID: "user-1", Motivation: "direct",
		Stashed: false, CreatedAt: now,
	}
	moments := []writingdomain.Moment{
		{ID: "mom-1", TraceID: "tr-1", UserID: "user-1", Content: "一些内容", CreatedAt: now},
	}

	starCreated := false
	traceStashed := false
	done := make(chan struct{})

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
		listMomentsByTraceIDFn: func(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
			return moments, nil
		},
	}

	traceStasher := &mockTraceStasher{
		markStashedFn: func(ctx context.Context, traceID string) error {
			traceStashed = true
			return nil
		},
	}

	starRepo := &mockStarRepo{
		createFn: func(ctx context.Context, star *domain.Star) error {
			starCreated = true
			if star.TraceID != "tr-1" {
				t.Errorf("expected traceID 'tr-1', got %q", star.TraceID)
			}
			if star.Topic != "聚合中" {
				t.Errorf("expected topic '聚合中', got %q", star.Topic)
			}
			return nil
		},
		updateTopicFn: func(ctx context.Context, starID string, topic string) error {
			return nil
		},
	}

	constellationRepo := &mockConstellationRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Constellation, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, c *domain.Constellation) error {
			if len(c.StarIDs) != 1 {
				t.Errorf("expected 1 star in constellation, got %d", len(c.StarIDs))
			}
			return nil
		},
	}

	assetGen := &mockAssetGen{
		generateFn: func(ctx context.Context, moments []writingdomain.Moment) (string, []float32, string, string, []string, error) {
			return "关于自我探索", []float32{0.1, 0.2}, "测试星座", "一些洞察", []string{"提示1", "提示2"}, nil
		},
	}

	profileGen := &mockTraceProfileGen{
		generateFn: func(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
			return &domain.TraceProfile{
					TraceID:     trace.ID,
					UserID:      trace.UserID,
					Topic:       "关于自我探索",
					Summary:     "用户记录了一些自我探索相关内容。",
					ProfileText: "主题：关于自我探索",
					Status:      domain.TraceProfileStatusReady,
				},
				&domain.TraceProfileVector{TraceID: trace.ID, UserID: trace.UserID, Model: "test", Dim: 2, Embedding: []float32{0.1, 0.2}},
				nil
		},
	}
	profileRepo := &mockTraceProfileRepo{
		upsertFn: func(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error {
			return nil
		},
	}
	constellationProfileRepo := &mockConstellationProfileRepo{
		findCandidatesFn: func(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error) {
			return nil, nil
		},
		upsertFn: func(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
			return nil
		},
		addMembershipFn: func(ctx context.Context, membership domain.ConstellationMembership) error {
			close(done)
			return nil
		},
	}

	uc := NewStashTraceUseCaseWithTraceProfile(
		traceReader, traceStasher, starRepo, constellationRepo,
		assetGen,
		profileGen, profileRepo, constellationProfileRepo,
		&mockSeqIDGen{ids: []string{"star-1", "constellation-1"}},
	)

	star, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if star.ID != "star-1" {
		t.Fatalf("expected star id 'star-1', got %q", star.ID)
	}
	if !starCreated {
		t.Fatal("star was not created")
	}
	if !traceStashed {
		t.Fatal("trace was not marked stashed")
	}

	// Wait for async clustering to finish
	select {
	case <-done:
		// constellation created asynchronously
	case <-time.After(2 * time.Second):
		t.Fatal("async clustering did not complete within timeout")
	}
}

func TestStashTrace_GeneratesTraceProfileAsync(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{
		ID: "tr-1", UserID: "user-1", Motivation: "direct",
		Stashed: false, CreatedAt: now,
	}
	moments := []writingdomain.Moment{
		{ID: "mom-1", TraceID: "tr-1", UserID: "user-1", Content: "搬进了自己的出租屋", CreatedAt: now},
	}

	profileDone := make(chan struct{})
	membershipDone := make(chan struct{})

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
		listMomentsByTraceIDFn: func(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
			return moments, nil
		},
	}
	traceStasher := &mockTraceStasher{
		markStashedFn: func(ctx context.Context, traceID string) error { return nil },
	}
	starRepo := &mockStarRepo{
		createFn:      func(ctx context.Context, star *domain.Star) error { return nil },
		updateTopicFn: func(ctx context.Context, starID string, topic string) error { return nil },
	}
	constellationRepo := &mockConstellationRepo{
		createFn: func(ctx context.Context, c *domain.Constellation) error {
			return nil
		},
	}
	assetGen := &mockAssetGen{
		generateFn: func(ctx context.Context, moments []writingdomain.Moment) (string, []float32, string, string, []string, error) {
			return "独立生活", []float32{0.1}, "独立生活", "你正在进入新的生活阶段。", []string{"还有什么想说？"}, nil
		},
	}
	profileGen := &mockTraceProfileGen{
		generateFn: func(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
			return &domain.TraceProfile{
					TraceID: trace.ID,
					UserID:  trace.UserID,
					Topic:   "独立生活",
					Status:  domain.TraceProfileStatusReady,
				},
				&domain.TraceProfileVector{TraceID: trace.ID, UserID: trace.UserID, Model: "test", Dim: 1, Embedding: []float32{0.1}},
				nil
		},
	}
	profileRepo := &mockTraceProfileRepo{
		upsertFn: func(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error {
			if profile.TraceID != "tr-1" {
				t.Errorf("expected trace profile for tr-1, got %q", profile.TraceID)
			}
			if vector == nil {
				t.Error("expected trace profile vector")
			}
			close(profileDone)
			return nil
		},
	}
	constellationProfileRepo := &mockConstellationProfileRepo{
		findCandidatesFn: func(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error) {
			return nil, nil
		},
		upsertFn: func(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
			return nil
		},
		addMembershipFn: func(ctx context.Context, membership domain.ConstellationMembership) error {
			close(membershipDone)
			return nil
		},
	}

	uc := NewStashTraceUseCaseWithTraceProfile(
		traceReader, traceStasher, starRepo, constellationRepo,
		assetGen,
		profileGen, profileRepo, constellationProfileRepo,
		&mockSeqIDGen{ids: []string{"star-1", "constellation-1"}},
	)

	if _, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for name, ch := range map[string]<-chan struct{}{"profile": profileDone, "membership": membershipDone} {
		select {
		case <-ch:
		case <-time.After(2 * time.Second):
			t.Fatalf("%s async path did not complete within timeout", name)
		}
	}
}

func TestStashTrace_ProfileGenerationRetries(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{
		ID: "tr-1", UserID: "user-1", Motivation: "direct",
		Stashed: false, CreatedAt: now,
	}
	moments := []writingdomain.Moment{
		{ID: "mom-1", TraceID: "tr-1", UserID: "user-1", Content: "今天状态反复", CreatedAt: now},
	}

	done := make(chan struct{})
	attempts := 0
	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
		listMomentsByTraceIDFn: func(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
			return moments, nil
		},
	}
	traceStasher := &mockTraceStasher{
		markStashedFn: func(ctx context.Context, traceID string) error { return nil },
	}
	starRepo := &mockStarRepo{
		createFn:      func(ctx context.Context, star *domain.Star) error { return nil },
		updateTopicFn: func(ctx context.Context, starID string, topic string) error { return nil },
	}
	constellationRepo := &mockConstellationRepo{
		createFn: func(ctx context.Context, c *domain.Constellation) error { return nil },
	}
	assetGen := &mockAssetGen{
		generateFn: func(ctx context.Context, moments []writingdomain.Moment) (string, []float32, string, string, []string, error) {
			return "状态反复", nil, "状态反复", "这些片段都围绕状态起伏。", nil, nil
		},
	}
	profileGen := &mockTraceProfileGen{
		generateFn: func(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
			attempts++
			if attempts == 1 {
				return nil, nil, errors.New("temporary profile failure")
			}
			return &domain.TraceProfile{
					TraceID:     trace.ID,
					UserID:      trace.UserID,
					Topic:       "状态反复",
					Summary:     "用户表达了状态反复的感受。",
					ProfileText: "主题：状态反复",
					Status:      domain.TraceProfileStatusReady,
				},
				&domain.TraceProfileVector{TraceID: trace.ID, UserID: trace.UserID, Model: "test", Dim: 2, Embedding: []float32{0.3, 0.7}},
				nil
		},
	}
	profileRepo := &mockTraceProfileRepo{
		upsertFn: func(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error {
			return nil
		},
	}
	constellationProfileRepo := &mockConstellationProfileRepo{
		findCandidatesFn: func(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error) {
			return nil, nil
		},
		upsertFn: func(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
			return nil
		},
		addMembershipFn: func(ctx context.Context, membership domain.ConstellationMembership) error {
			close(done)
			return nil
		},
	}

	uc := NewStashTraceUseCaseWithTraceProfile(
		traceReader, traceStasher, starRepo, constellationRepo,
		assetGen,
		profileGen, profileRepo, constellationProfileRepo,
		&mockSeqIDGen{ids: []string{"star-1", "constellation-1"}},
	)

	if _, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("profile clustering did not complete within timeout")
	}
	if attempts != 2 {
		t.Fatalf("expected profile generation to retry once, got %d attempts", attempts)
	}
}

func TestStashTrace_ProfileClusteringCreatesPrimaryConstellation(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{
		ID: "tr-1", UserID: "user-1", Motivation: "direct",
		Stashed: false, CreatedAt: now,
	}
	moments := []writingdomain.Moment{
		{ID: "mom-1", TraceID: "tr-1", UserID: "user-1", Content: "小红书的帖子挺容易让人焦虑的", CreatedAt: now},
	}

	done := make(chan struct{})
	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
		listMomentsByTraceIDFn: func(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
			return moments, nil
		},
	}
	traceStasher := &mockTraceStasher{
		markStashedFn: func(ctx context.Context, traceID string) error { return nil },
	}
	starRepo := &mockStarRepo{
		createFn: func(ctx context.Context, star *domain.Star) error {
			if star.ID != "star-1" {
				t.Errorf("expected star-1, got %q", star.ID)
			}
			return nil
		},
		updateTopicFn: func(ctx context.Context, starID string, topic string) error {
			if topic != "小红书焦虑" {
				t.Errorf("expected star topic from TraceProfile, got %q", topic)
			}
			return nil
		},
	}
	constellationRepo := &mockConstellationRepo{
		createFn: func(ctx context.Context, c *domain.Constellation) error {
			if c.ID != "constellation-1" {
				t.Errorf("expected constellation-1, got %q", c.ID)
			}
			if c.Topic != "小红书焦虑" {
				t.Errorf("expected constellation topic from TraceProfile, got %q", c.Topic)
			}
			if len(c.StarIDs) != 1 || c.StarIDs[0] != "star-1" {
				t.Errorf("expected star-1 in constellation, got %#v", c.StarIDs)
			}
			return nil
		},
	}
	assetGen := &mockAssetGen{
		generateFn: func(ctx context.Context, moments []writingdomain.Moment) (string, []float32, string, string, []string, error) {
			return "旧资产主题", nil, "小红书焦虑", "这些内容都和小红书带来的焦虑有关。", []string{"什么时候最明显？"}, nil
		},
	}
	profileGen := &mockTraceProfileGen{
		generateFn: func(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
			return &domain.TraceProfile{
					TraceID:     trace.ID,
					UserID:      trace.UserID,
					Topic:       "小红书焦虑",
					Summary:     "用户表达了小红书帖子容易引起焦虑。",
					Keywords:    []string{"小红书", "帖子", "焦虑"},
					Emotions:    []string{"焦虑"},
					Scenes:      []string{"社交媒体"},
					ProfileText: "主题：小红书焦虑",
					Status:      domain.TraceProfileStatusReady,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				&domain.TraceProfileVector{TraceID: trace.ID, UserID: trace.UserID, Model: "test", Dim: 2, Embedding: []float32{1, 0}},
				nil
		},
	}
	profileRepo := &mockTraceProfileRepo{
		upsertFn: func(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error {
			if profile.TraceID != "tr-1" || vector == nil {
				t.Fatalf("unexpected trace profile upsert: %#v %#v", profile, vector)
			}
			return nil
		},
	}
	constellationProfileRepo := &mockConstellationProfileRepo{
		findCandidatesFn: func(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error) {
			if limit != constellationCandidateLimit {
				t.Errorf("expected candidate limit %d, got %d", constellationCandidateLimit, limit)
			}
			return nil, nil
		},
		upsertFn: func(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
			if profile.ConstellationID != "constellation-1" {
				t.Errorf("expected profile for constellation-1, got %q", profile.ConstellationID)
			}
			if vector == nil || len(vector.CentroidEmbedding) != 2 {
				t.Errorf("expected constellation vector, got %#v", vector)
			}
			return nil
		},
		addMembershipFn: func(ctx context.Context, membership domain.ConstellationMembership) error {
			if membership.MatchType != domain.ConstellationMatchTypePrimary {
				t.Errorf("expected primary membership, got %q", membership.MatchType)
			}
			if membership.StarID != "star-1" || membership.ConstellationID != "constellation-1" {
				t.Errorf("unexpected membership: %#v", membership)
			}
			close(done)
			return nil
		},
	}

	uc := NewStashTraceUseCaseWithTraceProfile(
		traceReader, traceStasher, starRepo, constellationRepo,
		assetGen,
		profileGen, profileRepo, constellationProfileRepo,
		&mockSeqIDGen{ids: []string{"star-1", "constellation-1"}},
	)

	if _, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("profile clustering did not complete within timeout")
	}
}

func TestRankConstellationCandidates_PatternTagsAndSingleTraceAllowMiddleMatch(t *testing.T) {
	traceProfile := &domain.TraceProfile{
		TraceID:     "tr-new",
		UserID:      "user-1",
		Topic:       "入职前紧张",
		Keywords:    []string{"入职", "紧张", "担心"},
		Emotions:    []string{"紧张", "担心"},
		Scenes:      []string{"工作", "入职"},
		PatternTags: []string{"新开始", "工作变化", "不确定性", "担心"},
	}
	traceVector := &domain.TraceProfileVector{
		TraceID:   "tr-new",
		UserID:    "user-1",
		Embedding: []float32{1, 0},
	}
	candidates := []domain.ConstellationProfileCandidate{
		{
			Profile: domain.ConstellationProfile{
				ConstellationID: "c-job",
				UserID:          "user-1",
				Topic:           "入职前的期待与担心",
				Keywords:        []string{"入职", "担心", "期待"},
				Emotions:        []string{"期待", "担心"},
				Scenes:          []string{"工作", "入职"},
				PatternTags:     []string{"新开始", "工作变化", "期待", "担心"},
				TraceCount:      1,
				MomentCount:     1,
			},
			Vector: domain.ConstellationProfileVector{
				ProfileEmbedding:  []float32{0.64, 0.7683749},
				CentroidEmbedding: []float32{0.64, 0.7683749},
			},
		},
	}

	ranked := rankConstellationCandidates(traceProfile, traceVector, candidates)
	if len(ranked) != 1 {
		t.Fatalf("expected one ranked candidate, got %d", len(ranked))
	}
	top := ranked[0]
	if top.score < constellationMiddleMatchThreshold {
		t.Fatalf("score = %.3f, want >= middle threshold %.3f", top.score, constellationMiddleMatchThreshold)
	}
	if !isPrimaryAttachCandidate(top) {
		t.Fatalf("expected candidate to be accepted as primary attach candidate: %#v", top)
	}
	if top.patternTagsOverlap <= 0 {
		t.Fatalf("expected pattern_tags overlap, got %.3f", top.patternTagsOverlap)
	}
	if top.explainableMiddle {
		t.Fatalf("score already reaches middle; explainable_middle should not be needed")
	}
	components := constellationScoreComponents(top)
	if components["centroid_similarity"] != 0 {
		t.Fatalf("single trace candidate should not double count centroid, components=%#v", components)
	}
	if len(top.matchedPatternTags) != 3 {
		t.Fatalf("matched pattern tags = %#v, want 3 matches", top.matchedPatternTags)
	}
}

func TestStashTrace_TraceNotFound(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return nil, writingdomain.ErrTraceNotFound
		},
	}

	uc := NewStashTraceUseCase(
		traceReader, nil, nil, nil, nil, nil,
	)

	_, err := uc.Execute(ctx, StashTraceInput{TraceID: "nonexistent"})
	if !errors.Is(err, writingdomain.ErrTraceNotFound) {
		t.Fatalf("expected ErrTraceNotFound, got %v", err)
	}
}

func TestStashTrace_AlreadyStashed(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{
		ID: "tr-1", UserID: "user-1", Stashed: true, CreatedAt: now,
	}

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
	}

	uc := NewStashTraceUseCase(
		traceReader, nil, nil, nil, nil, nil,
	)

	_, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-1"})
	if !errors.Is(err, domain.ErrTraceAlreadyStashed) {
		t.Fatalf("expected ErrTraceAlreadyStashed, got %v", err)
	}
}

func TestStashTrace_WrongUser(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{
		ID: "tr-1", UserID: "user-2", Stashed: false, CreatedAt: now,
	}

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
	}

	uc := NewStashTraceUseCase(
		traceReader, nil, nil, nil, nil, nil,
	)

	_, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-1"})
	if !errors.Is(err, domain.ErrTraceNotFound) {
		t.Fatalf("expected ErrTraceNotFound, got %v", err)
	}
}
