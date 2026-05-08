package app

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/writing/domain"
)

// CreateMomentUseCase orchestrates the creation of a Moment (and optionally a Trace),
// embedding generation, and echo matching.
type CreateMomentUseCase struct {
	traces    domain.TraceRepository
	moments   domain.MomentRepository
	echos     domain.EchoRepository
	embedding domain.EmbeddingGenerator
	echo      domain.EchoMatcher
	ids       IDGenerator
}

func NewCreateMomentUseCase(
	traces domain.TraceRepository,
	moments domain.MomentRepository,
	echos domain.EchoRepository,
	embedding domain.EmbeddingGenerator,
	echo domain.EchoMatcher,
	ids IDGenerator,
) *CreateMomentUseCase {
	return &CreateMomentUseCase{
		traces:    traces,
		moments:   moments,
		echos:     echos,
		embedding: embedding,
		echo:      echo,
		ids:       ids,
	}
}

type CreateMomentInput struct {
	Content    string
	TraceID    string
	Motivation string
}

type CreateMomentOutput struct {
	Moment domain.Moment
	Echo   *domain.Echo
}

func (uc *CreateMomentUseCase) Execute(ctx context.Context, input CreateMomentInput) (*CreateMomentOutput, error) {
	if input.Content == "" {
		return nil, domain.ErrEmptyContent
	}

	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	var newTrace bool
	traceID := input.TraceID
	if traceID == "" {
		motivation := input.Motivation
		if motivation == "" {
			motivation = "direct"
		}
		trace, err := uc.createTrace(ctx, userID, motivation)
		if err != nil {
			return nil, fmt.Errorf("create trace: %w", err)
		}
		traceID = trace.ID
		newTrace = true
	} else {
		existing, err := uc.traces.GetByID(ctx, traceID)
		if err != nil {
			return nil, fmt.Errorf("get trace: %w", err)
		}
		if existing.UserID != userID {
			return nil, fmt.Errorf("trace does not belong to user")
		}
	}

	moment, err := uc.createMoment(ctx, userID, traceID, input.Content)
	if err != nil {
		if newTrace {
			_ = uc.traces.Delete(ctx, traceID)
		}
		return nil, fmt.Errorf("create moment: %w", err)
	}

	echo, err := uc.matchEcho(ctx, moment, userID)
	if err != nil {
		return nil, fmt.Errorf("match echo: %w", err)
	}

	return &CreateMomentOutput{
		Moment: *moment,
		Echo:   echo,
	}, nil
}

func (uc *CreateMomentUseCase) createTrace(ctx context.Context, userID, motivation string) (*domain.Trace, error) {
	trace := &domain.Trace{
		ID:         uc.ids.New(),
		UserID:     userID,
		Motivation: motivation,
		Stashed:    false,
		CreatedAt:  time.Now(),
	}
	if err := uc.traces.Create(ctx, trace); err != nil {
		return nil, err
	}
	return trace, nil
}

func (uc *CreateMomentUseCase) createMoment(ctx context.Context, userID, traceID, content string) (*domain.Moment, error) {
	embeddings, err := uc.embedding.Generate(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("generate embedding: %w", err)
	}

	moment := &domain.Moment{
		ID:         uc.ids.New(),
		TraceID:    traceID,
		UserID:     userID,
		Content:    content,
		Embeddings: embeddings,
		CreatedAt:  time.Now(),
	}

	if err := uc.moments.Create(ctx, moment); err != nil {
		return nil, err
	}
	return moment, nil
}

func (uc *CreateMomentUseCase) matchEcho(ctx context.Context, moment *domain.Moment, userID string) (*domain.Echo, error) {
	allMoments, err := uc.moments.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}

	history := excludeSelf(allMoments, moment.ID)
	if len(history) == 0 {
		return nil, nil
	}

	matches, err := uc.echo.Match(ctx, moment, history)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, nil
	}

	matchedIDs := make([]string, len(matches))
	similarities := make([]float64, len(matches))
	for i, m := range matches {
		matchedIDs[i] = m.MomentID
		similarities[i] = m.Similarity
	}

	echo := &domain.Echo{
		ID:               uc.ids.New(),
		MomentID:         moment.ID,
		UserID:           userID,
		MatchedMomentIDs: matchedIDs,
		Similarities:     similarities,
		CreatedAt:        time.Now(),
	}

	if err := uc.echos.Create(ctx, echo); err != nil {
		return nil, fmt.Errorf("persist echo: %w", err)
	}

	return echo, nil
}

func excludeSelf(moments []domain.Moment, selfID string) []domain.Moment {
	var result []domain.Moment
	for _, m := range moments {
		if m.ID != selfID {
			result = append(result, m)
		}
	}
	return result
}
