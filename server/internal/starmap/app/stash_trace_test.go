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

type mockBorderlineJudge struct {
	judgeFn func(ctx context.Context, input domain.ConstellationBorderlineJudgeInput) (*domain.ConstellationBorderlineJudgement, error)
}

func (m *mockBorderlineJudge) Judge(ctx context.Context, input domain.ConstellationBorderlineJudgeInput) (*domain.ConstellationBorderlineJudgement, error) {
	return m.judgeFn(ctx, input)
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
		profileGen, nil, profileRepo, constellationProfileRepo,
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
		profileGen, nil, profileRepo, constellationProfileRepo,
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
		profileGen, nil, profileRepo, constellationProfileRepo,
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

func TestStashTrace_BorderlineJudgementAttachesPrimaryConstellation(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{ID: "tr-2", UserID: "user-1", Stashed: false, CreatedAt: now}
	moments := []writingdomain.Moment{
		{ID: "mom-2", TraceID: "tr-2", UserID: "user-1", Content: "最后跟我说可能不能入职，打乱计划，心烦", CreatedAt: now},
	}

	done := make(chan struct{})
	judgeCalled := false
	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
		listMomentsByTraceIDFn: func(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
			return moments, nil
		},
	}
	traceStasher := &mockTraceStasher{markStashedFn: func(ctx context.Context, traceID string) error { return nil }}
	starRepo := &mockStarRepo{
		createFn:      func(ctx context.Context, star *domain.Star) error { return nil },
		updateTopicFn: func(ctx context.Context, starID string, topic string) error { return nil },
	}
	constellationRepo := &mockConstellationRepo{
		createFn: func(ctx context.Context, c *domain.Constellation) error {
			t.Fatalf("should attach existing constellation, created %#v", c)
			return nil
		},
		findByIDFn: func(ctx context.Context, id string) (*domain.Constellation, error) {
			if id != "c-job" {
				t.Fatalf("unexpected constellation id %q", id)
			}
			return &domain.Constellation{ID: "c-job", UserID: "user-1", Topic: "入职申请受阻", Name: "入职卡住", StarIDs: []string{"star-old"}}, nil
		},
		updateFn: func(ctx context.Context, c *domain.Constellation) error {
			if !containsString(c.StarIDs, "star-2") {
				t.Fatalf("expected star-2 attached, got %#v", c.StarIDs)
			}
			return nil
		},
	}
	profileGen := &mockTraceProfileGen{
		generateFn: func(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
			return &domain.TraceProfile{
					TraceID:                trace.ID,
					UserID:                 trace.UserID,
					Topic:                  "入职计划被打乱",
					Summary:                "用户表达入职可能受阻，原有计划被打乱后的心烦。",
					Keywords:               []string{"计划"},
					Emotions:               []string{"心烦"},
					Scenes:                 []string{"工作"},
					CentralPattern:         "准备中的计划被外部审核结果打断",
					PatternTags:            []string{"计划受阻"},
					RepresentativeMomentID: "mom-2",
					Status:                 domain.TraceProfileStatusReady,
				},
				&domain.TraceProfileVector{TraceID: trace.ID, UserID: trace.UserID, Model: "test", Dim: 2, Embedding: []float32{1, 0}},
				nil
		},
	}
	profileRepo := &mockTraceProfileRepo{upsertFn: func(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error {
		return nil
	}}
	constellationProfileRepo := &mockConstellationProfileRepo{
		findCandidatesFn: func(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error) {
			return []domain.ConstellationProfileCandidate{
				{
					Profile: domain.ConstellationProfile{
						ConstellationID:  "c-job",
						UserID:           "user-1",
						Topic:            "入职申请受阻",
						Summary:          "入职材料和审核反复卡住，影响后续计划。",
						Keywords:         []string{"审核"},
						Emotions:         []string{"烦"},
						Scenes:           []string{"工作"},
						CentralPattern:   "入职流程被外部审核卡住",
						PatternTags:      []string{"流程受阻"},
						ThemeCode:        "theme_job_onboarding_blocked",
						ThemeLabel:       "入职受阻",
						ThemeDescription: "围绕入职、审核、计划推进被打断的反复处境。",
						ThemeExamples:    []string{"入职申请被驳回了，审核又慢"},
						TraceCount:       1,
						MomentCount:      1,
					},
					Vector: domain.ConstellationProfileVector{
						Model:             "test",
						Dim:               2,
						ProfileEmbedding:  []float32{0.4, 0.916515},
						CentroidEmbedding: []float32{0.4, 0.916515},
						CreatedAt:         now,
					},
				},
			}, nil
		},
		upsertFn: func(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
			if profile.ConstellationID != "c-job" {
				t.Fatalf("expected existing constellation profile update, got %q", profile.ConstellationID)
			}
			if profile.ThemeCode != "theme_job_onboarding_blocked" {
				t.Fatalf("expected theme code preserved, got %q", profile.ThemeCode)
			}
			return nil
		},
		addMembershipFn: func(ctx context.Context, membership domain.ConstellationMembership) error {
			if membership.ConstellationID != "c-job" || membership.MatchType != domain.ConstellationMatchTypePrimary {
				t.Fatalf("unexpected membership: %#v", membership)
			}
			if membership.MatchScore >= constellationMiddleMatchThreshold {
				t.Fatalf("borderline test should attach below middle threshold, score=%.3f", membership.MatchScore)
			}
			if membership.MatchReason != "都在表达入职推进被外部结果打断，导致原计划无法继续。" {
				t.Fatalf("expected LLM reason, got %q", membership.MatchReason)
			}
			close(done)
			return nil
		},
	}
	borderlineJudge := &mockBorderlineJudge{
		judgeFn: func(ctx context.Context, input domain.ConstellationBorderlineJudgeInput) (*domain.ConstellationBorderlineJudgement, error) {
			judgeCalled = true
			if len(input.Candidates) != 1 || input.Candidates[0].ConstellationID != "c-job" {
				t.Fatalf("unexpected borderline candidates: %#v", input.Candidates)
			}
			return &domain.ConstellationBorderlineJudgement{
				Decision:        domain.ConstellationBorderlineDecisionUseExisting,
				ConstellationID: "c-job",
				ThemeCode:       "theme_job_onboarding_blocked",
				Confidence:      0.78,
				SharedSituation: "入职推进被外部审核或结果打断，原计划被迫悬空。",
				MatchDimensions: []string{"situation", "self_pattern"},
				Reason:          "都在表达入职推进被外部结果打断，导致原计划无法继续。",
			}, nil
		},
	}

	uc := NewStashTraceUseCaseWithTraceProfile(
		traceReader, traceStasher, starRepo, constellationRepo,
		&mockAssetGen{}, profileGen, borderlineJudge, profileRepo, constellationProfileRepo,
		&mockSeqIDGen{ids: []string{"star-2"}},
	)

	if _, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("borderline clustering did not complete within timeout")
	}
	if !judgeCalled {
		t.Fatal("expected borderline judge to be called")
	}
}

func TestStashTrace_BorderlineLowConfidenceCreatesNewConstellation(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{ID: "tr-3", UserID: "user-1", Stashed: false, CreatedAt: now}
	moments := []writingdomain.Moment{{ID: "mom-3", TraceID: "tr-3", UserID: "user-1", Content: "新的工作安排又变了", CreatedAt: now}}
	done := make(chan struct{})
	created := false

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) { return trace, nil },
		listMomentsByTraceIDFn: func(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
			return moments, nil
		},
	}
	traceStasher := &mockTraceStasher{markStashedFn: func(ctx context.Context, traceID string) error { return nil }}
	starRepo := &mockStarRepo{
		createFn:      func(ctx context.Context, star *domain.Star) error { return nil },
		updateTopicFn: func(ctx context.Context, starID string, topic string) error { return nil },
	}
	constellationRepo := &mockConstellationRepo{
		createFn: func(ctx context.Context, c *domain.Constellation) error {
			created = true
			if c.ID != "c-new" {
				t.Fatalf("expected new constellation id c-new, got %q", c.ID)
			}
			return nil
		},
	}
	assetGen := &mockAssetGen{
		generateFn: func(ctx context.Context, moments []writingdomain.Moment) (string, []float32, string, string, []string, error) {
			return "工作安排变化", nil, "安排变了", "工作安排发生变化。", nil, nil
		},
	}
	profileGen := &mockTraceProfileGen{
		generateFn: func(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
			return &domain.TraceProfile{TraceID: trace.ID, UserID: trace.UserID, Topic: "工作安排变化", Summary: "工作安排变化。", Scenes: []string{"工作"}, Status: domain.TraceProfileStatusReady},
				&domain.TraceProfileVector{TraceID: trace.ID, UserID: trace.UserID, Model: "test", Dim: 2, Embedding: []float32{1, 0}},
				nil
		},
	}
	profileRepo := &mockTraceProfileRepo{upsertFn: func(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error {
		return nil
	}}
	constellationProfileRepo := &mockConstellationProfileRepo{
		findCandidatesFn: func(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error) {
			return []domain.ConstellationProfileCandidate{
				{
					Profile: domain.ConstellationProfile{
						ConstellationID: "c-old",
						UserID:          "user-1",
						Topic:           "入职申请受阻",
						Scenes:          []string{"工作"},
						ThemeCode:       "theme_job_onboarding_blocked",
						TraceCount:      1,
					},
					Vector: domain.ConstellationProfileVector{Model: "test", Dim: 2, ProfileEmbedding: []float32{0.4, 0.916515}, CentroidEmbedding: []float32{0.4, 0.916515}},
				},
			}, nil
		},
		upsertFn: func(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
			if profile.ConstellationID != "c-new" {
				t.Fatalf("expected new constellation profile, got %q", profile.ConstellationID)
			}
			if profile.ThemeCode == "" || profile.ThemeLabel == "" {
				t.Fatalf("expected fallback theme codebook, got %#v", profile)
			}
			return nil
		},
		addMembershipFn: func(ctx context.Context, membership domain.ConstellationMembership) error {
			if membership.ConstellationID != "c-new" {
				t.Fatalf("expected new membership, got %#v", membership)
			}
			close(done)
			return nil
		},
	}
	borderlineJudge := &mockBorderlineJudge{
		judgeFn: func(ctx context.Context, input domain.ConstellationBorderlineJudgeInput) (*domain.ConstellationBorderlineJudgement, error) {
			return &domain.ConstellationBorderlineJudgement{
				Decision:        domain.ConstellationBorderlineDecisionUseExisting,
				ConstellationID: "c-old",
				ThemeCode:       "theme_job_onboarding_blocked",
				Confidence:      0.42,
				SharedSituation: "都和工作有关。",
				MatchDimensions: []string{"situation"},
				Reason:          "证据不足。",
			}, nil
		},
	}

	uc := NewStashTraceUseCaseWithTraceProfile(
		traceReader, traceStasher, starRepo, constellationRepo,
		assetGen, profileGen, borderlineJudge, profileRepo, constellationProfileRepo,
		&mockSeqIDGen{ids: []string{"star-3", "c-new"}},
	)

	if _, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-3"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("new constellation clustering did not complete within timeout")
	}
	if !created {
		t.Fatal("expected new constellation to be created")
	}
}

func TestStashTrace_BorderlineJudgementAttachesPrimaryAndSecondary(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{ID: "tr-4", UserID: "user-1", Stashed: false, CreatedAt: now}
	moments := []writingdomain.Moment{{ID: "mom-4", TraceID: "tr-4", UserID: "user-1", Content: "入职审核让我继续等反馈，感觉一直卡着", CreatedAt: now}}
	done := make(chan struct{})
	memberships := make([]domain.ConstellationMembership, 0, 2)

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) { return trace, nil },
		listMomentsByTraceIDFn: func(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
			return moments, nil
		},
	}
	traceStasher := &mockTraceStasher{markStashedFn: func(ctx context.Context, traceID string) error { return nil }}
	starRepo := &mockStarRepo{
		createFn:      func(ctx context.Context, star *domain.Star) error { return nil },
		updateTopicFn: func(ctx context.Context, starID string, topic string) error { return nil },
	}
	constellationRepo := &mockConstellationRepo{
		createFn: func(ctx context.Context, c *domain.Constellation) error {
			t.Fatalf("should not create constellation, created %#v", c)
			return nil
		},
		findByIDFn: func(ctx context.Context, id string) (*domain.Constellation, error) {
			switch id {
			case "c-onboarding":
				return &domain.Constellation{ID: id, UserID: "user-1", Topic: "入职流程卡住", Name: "入职卡住", StarIDs: []string{"star-old"}}, nil
			case "c-identity":
				return &domain.Constellation{ID: id, UserID: "user-1", Topic: "被审核感", Name: "总被确认", StarIDs: []string{"star-other"}}, nil
			default:
				t.Fatalf("unexpected constellation id %q", id)
				return nil, nil
			}
		},
		updateFn: func(ctx context.Context, c *domain.Constellation) error {
			if !containsString(c.StarIDs, "star-4") {
				t.Fatalf("expected star-4 attached, got %#v", c.StarIDs)
			}
			return nil
		},
	}
	profileGen := &mockTraceProfileGen{
		generateFn: func(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
			return &domain.TraceProfile{
					TraceID:                trace.ID,
					UserID:                 trace.UserID,
					Topic:                  "入职审核等待",
					Summary:                "用户在入职审核中等待反馈，感觉流程一直卡住。",
					Keywords:               []string{"入职", "审核", "反馈", "等待"},
					Emotions:               []string{"无奈"},
					Scenes:                 []string{"入职", "工作"},
					CentralPattern:         "入职推进被审核反馈拖住，只能等待。",
					PatternTags:            []string{"等待反馈", "流程受阻"},
					RepresentativeMomentID: "mom-4",
					Status:                 domain.TraceProfileStatusReady,
				},
				&domain.TraceProfileVector{TraceID: trace.ID, UserID: trace.UserID, Model: "test", Dim: 2, Embedding: []float32{1, 0}},
				nil
		},
	}
	profileRepo := &mockTraceProfileRepo{upsertFn: func(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error {
		return nil
	}}
	constellationProfileRepo := &mockConstellationProfileRepo{
		findCandidatesFn: func(ctx context.Context, userID string, embedding []float32, limit int) ([]domain.ConstellationProfileCandidate, error) {
			return []domain.ConstellationProfileCandidate{
				{
					Profile: domain.ConstellationProfile{
						ConstellationID:  "c-onboarding",
						UserID:           "user-1",
						Topic:            "入职资料问题",
						Summary:          "入职过程中资料、审核和反馈反复卡住。",
						Keywords:         []string{"入职资料", "审核"},
						Scenes:           []string{"入职", "工作"},
						PatternTags:      []string{"流程受阻"},
						ThemeCode:        "theme_onboarding_blocked",
						ThemeLabel:       "入职流程卡住",
						ThemeDescription: "入职、资料、审核、反馈、等待等流程反复卡住。",
						TraceCount:       1,
						MomentCount:      1,
					},
					Vector: domain.ConstellationProfileVector{Model: "test", Dim: 2, ProfileEmbedding: []float32{0.5, 0.8660254}, CentroidEmbedding: []float32{0.5, 0.8660254}, CreatedAt: now},
				},
				{
					Profile: domain.ConstellationProfile{
						ConstellationID:  "c-identity",
						UserID:           "user-1",
						Topic:            "被审核感",
						Summary:          "反复处在被确认、被审核的位置。",
						Keywords:         []string{"审核", "确认"},
						Scenes:           []string{"工作"},
						PatternTags:      []string{"被动等待"},
						ThemeCode:        "theme_being_reviewed",
						ThemeLabel:       "被审核感",
						ThemeDescription: "反复处在等待别人确认或审核的位置。",
						TraceCount:       1,
						MomentCount:      1,
					},
					Vector: domain.ConstellationProfileVector{Model: "test", Dim: 2, ProfileEmbedding: []float32{0.42, 0.907524}, CentroidEmbedding: []float32{0.42, 0.907524}, CreatedAt: now},
				},
			}, nil
		},
		upsertFn: func(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
			if profile.ConstellationID != "c-onboarding" && profile.ConstellationID != "c-identity" {
				t.Fatalf("unexpected profile upsert %q", profile.ConstellationID)
			}
			return nil
		},
		addMembershipFn: func(ctx context.Context, membership domain.ConstellationMembership) error {
			memberships = append(memberships, membership)
			if len(memberships) == 2 {
				close(done)
			}
			return nil
		},
	}
	borderlineJudge := &mockBorderlineJudge{
		judgeFn: func(ctx context.Context, input domain.ConstellationBorderlineJudgeInput) (*domain.ConstellationBorderlineJudgement, error) {
			if len(input.Candidates) != 2 {
				t.Fatalf("expected two borderline candidates, got %#v", input.Candidates)
			}
			return &domain.ConstellationBorderlineJudgement{
				Decision: domain.ConstellationBorderlineDecisionUseExisting,
				Primary: &domain.ConstellationBorderlineSelection{
					ConstellationID: "c-onboarding",
					ThemeCode:       "theme_onboarding_blocked",
					Confidence:      0.82,
					SharedTheme:     "入职资料、审核和反馈反复卡住。",
					MatchDimensions: []string{"situation"},
					Reason:          "这是入职流程卡住主题的自然延伸。",
				},
				Secondary: []domain.ConstellationBorderlineSelection{
					{
						ConstellationID: "c-identity",
						ThemeCode:       "theme_being_reviewed",
						Confidence:      0.72,
						SharedTheme:     "也可以从反复处在被审核位置的视角理解。",
						MatchDimensions: []string{"identity"},
						Reason:          "副视角是被确认和等待审核的位置感。",
					},
				},
			}, nil
		},
	}

	uc := NewStashTraceUseCaseWithTraceProfile(
		traceReader, traceStasher, starRepo, constellationRepo,
		&mockAssetGen{}, profileGen, borderlineJudge, profileRepo, constellationProfileRepo,
		&mockSeqIDGen{ids: []string{"star-4"}},
	)

	if _, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-4"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("primary and secondary clustering did not complete within timeout")
	}
	if len(memberships) != 2 {
		t.Fatalf("expected two memberships, got %#v", memberships)
	}
	if memberships[0].ConstellationID != "c-onboarding" || memberships[0].MatchType != domain.ConstellationMatchTypePrimary {
		t.Fatalf("unexpected primary membership: %#v", memberships[0])
	}
	if memberships[1].ConstellationID != "c-identity" || memberships[1].MatchType != domain.ConstellationMatchTypeSecondary {
		t.Fatalf("unexpected secondary membership: %#v", memberships[1])
	}
	if memberships[1].Weight != 0.5 {
		t.Fatalf("secondary weight = %.1f, want 0.5", memberships[1].Weight)
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
