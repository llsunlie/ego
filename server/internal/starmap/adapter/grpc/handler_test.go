package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"ego-server/internal/starmap/app"
	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"
)

func userCtx(userID string) context.Context {
	return context.WithValue(context.Background(), "user_id", userID)
}

// --- fake use cases ---

type fakeStashTraceUC struct {
	star *domain.Star
	err  error
}

func (f *fakeStashTraceUC) Execute(ctx context.Context, input app.StashTraceInput) (*domain.Star, error) {
	return f.star, f.err
}

type fakeListConstellationsUC struct {
	out *app.ListConstellationsOutput
	err error
}

func (f *fakeListConstellationsUC) Execute(ctx context.Context) (*app.ListConstellationsOutput, error) {
	return f.out, f.err
}

type fakeGetConstellationUC struct {
	out *app.GetConstellationOutput
	err error
}

func (f *fakeGetConstellationUC) Execute(ctx context.Context, input app.GetConstellationInput) (*app.GetConstellationOutput, error) {
	return f.out, f.err
}

// --- tests ---

func TestHandler_StashTrace(t *testing.T) {
	now := time.Now()
	star := &domain.Star{
		ID: "star-1", UserID: "user-1", TraceID: "tr-1",
		Topic: "关于一些内容…", CreatedAt: now,
	}

	fake := &fakeStashTraceUC{star: star, err: nil}
	listUC := &fakeListConstellationsUC{}    // not used in this test
	getUC := &fakeGetConstellationUC{}       // not used in this test
	h := NewHandler(fake, listUC, getUC)

	res, err := h.StashTrace(userCtx("user-1"), &pb.StashTraceReq{TraceId: "tr-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Star.Id != "star-1" {
		t.Fatalf("expected star id 'star-1', got %q", res.Star.Id)
	}
	if res.Star.Topic != "关于一些内容…" {
		t.Fatalf("expected topic '关于一些内容…', got %q", res.Star.Topic)
	}
	if res.Star.TraceId != "tr-1" {
		t.Fatalf("expected trace_id 'tr-1', got %q", res.Star.TraceId)
	}
}

func TestHandler_StashTrace_Error(t *testing.T) {
	fake := &fakeStashTraceUC{err: domain.ErrTraceAlreadyStashed}
	listUC := &fakeListConstellationsUC{}
	getUC := &fakeGetConstellationUC{}
	h := NewHandler(fake, listUC, getUC)

	_, err := h.StashTrace(userCtx("user-1"), &pb.StashTraceReq{TraceId: "tr-1"})
	if !errors.Is(err, domain.ErrTraceAlreadyStashed) {
		t.Fatalf("expected ErrTraceAlreadyStashed, got %v", err)
	}
}

func TestHandler_ListConstellations(t *testing.T) {
	now := time.Now()
	fake := &fakeListConstellationsUC{
		out: &app.ListConstellationsOutput{
			Constellations: []domain.Constellation{
				{ID: "c1", Name: "星座A", StarIDs: []string{"s1", "s2"}, CreatedAt: now, UpdatedAt: now},
				{ID: "c2", Name: "星座B", StarIDs: []string{"s3"}, CreatedAt: now, UpdatedAt: now},
			},
			TotalStarCount: 3,
		},
	}
	stashUC := &fakeStashTraceUC{}
	getUC := &fakeGetConstellationUC{}
	h := NewHandler(stashUC, fake, getUC)

	res, err := h.ListConstellations(userCtx("user-1"), &pb.ListConstellationsReq{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TotalStarCount != 3 {
		t.Fatalf("expected 3 stars, got %d", res.TotalStarCount)
	}
	if len(res.Constellations) != 2 {
		t.Fatalf("expected 2 constellations, got %d", len(res.Constellations))
	}
	if res.Constellations[0].Name != "星座A" {
		t.Fatalf("expected '星座A', got %q", res.Constellations[0].Name)
	}
	if res.Constellations[0].StarCount != 2 {
		t.Fatalf("expected star_count 2, got %d", res.Constellations[0].StarCount)
	}
}

func TestHandler_GetConstellation(t *testing.T) {
	now := time.Now()
	fake := &fakeGetConstellationUC{
		out: &app.GetConstellationOutput{
			Constellation: domain.Constellation{
				ID: "c1", Name: "测试星座", StarIDs: []string{"s1"},
				ConstellationInsight: "一些洞察",
				TopicPrompts:         []string{"提示1"},
				CreatedAt: now, UpdatedAt: now,
			},
			Moments: []writingdomain.Moment{
				{ID: "mom-1", Content: "内容1", TraceID: "tr-1", CreatedAt: now},
			},
			Stars: []domain.Star{
				{ID: "s1", TraceID: "tr-1", Topic: "主题1", CreatedAt: now},
			},
		},
	}
	stashUC := &fakeStashTraceUC{}
	listUC := &fakeListConstellationsUC{}
	h := NewHandler(stashUC, listUC, fake)

	res, err := h.GetConstellation(userCtx("user-1"), &pb.GetConstellationReq{ConstellationId: "c1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Constellation.Id != "c1" {
		t.Fatalf("expected constellation 'c1', got %q", res.Constellation.Id)
	}
	if len(res.Moments) != 1 {
		t.Fatalf("expected 1 moment, got %d", len(res.Moments))
	}
	if len(res.Stars) != 1 {
		t.Fatalf("expected 1 star, got %d", len(res.Stars))
	}
	if res.Stars[0].Id != "s1" {
		t.Fatalf("expected star 's1', got %q", res.Stars[0].Id)
	}
}

func TestHandler_GetConstellation_Error(t *testing.T) {
	fake := &fakeGetConstellationUC{err: domain.ErrConstellationNotFound}
	stashUC := &fakeStashTraceUC{}
	listUC := &fakeListConstellationsUC{}
	h := NewHandler(stashUC, listUC, fake)

	_, err := h.GetConstellation(userCtx("user-1"), &pb.GetConstellationReq{ConstellationId: "nonexistent"})
	if !errors.Is(err, domain.ErrConstellationNotFound) {
		t.Fatalf("expected ErrConstellationNotFound, got %v", err)
	}
}

var _ StashTraceUseCase = (*fakeStashTraceUC)(nil)
var _ ListConstellationsUseCase = (*fakeListConstellationsUC)(nil)
var _ GetConstellationUseCase = (*fakeGetConstellationUC)(nil)
