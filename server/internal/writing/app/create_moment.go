package app

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/platform/logging"
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
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "CreateMoment: start", "content_len", len([]rune(input.Content)), "trace_id", input.TraceID, "motivation", input.Motivation)

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
			logger.ErrorContext(ctx, "CreateMoment: create trace failed", "error", err)
			return nil, fmt.Errorf("create trace: %w", err)
		}
		traceID = trace.ID
		newTrace = true
		logger.DebugContext(ctx, "CreateMoment: new trace created", "trace_id", traceID, "motivation", motivation)
	} else {
		existing, err := uc.traces.GetByID(ctx, traceID)
		if err != nil {
			logger.ErrorContext(ctx, "CreateMoment: get trace failed", "trace_id", traceID, "error", err)
			return nil, fmt.Errorf("get trace: %w", err)
		}
		if existing.UserID != userID {
			logger.WarnContext(ctx, "CreateMoment: trace ownership mismatch", "trace_id", traceID, "trace_user", existing.UserID, "caller_user", userID)
			return nil, fmt.Errorf("trace does not belong to user")
		}
		logger.DebugContext(ctx, "CreateMoment: appending to existing trace", "trace_id", traceID)
	}

	moment, err := uc.createMoment(ctx, userID, traceID, input.Content)
	if err != nil {
		if newTrace {
			logger.WarnContext(ctx, "CreateMoment: rolling back new trace", "trace_id", traceID, "error", err)
			_ = uc.traces.Delete(ctx, traceID)
		}
		return nil, fmt.Errorf("create moment: %w", err)
	}
	logger.DebugContext(ctx, "CreateMoment: moment created", "moment_id", moment.ID, "embedding_count", len(moment.Embeddings))

	echo, err := uc.matchEcho(ctx, moment, userID)
	if err != nil {
		logger.ErrorContext(ctx, "CreateMoment: echo matching failed", "moment_id", moment.ID, "error", err)
		return nil, fmt.Errorf("match echo: %w", err)
	}

	if echo != nil {
		logger.InfoContext(ctx, "CreateMoment: echo found", "moment_id", moment.ID, "echo_id", echo.ID, "matched_count", len(echo.MatchedMomentIDs))
	} else {
		logger.DebugContext(ctx, "CreateMoment: no echo (no matching history or first moment)", "moment_id", moment.ID)
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
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "CreateMoment: generating embedding", "content_len", len([]rune(content)))
	embeddings, err := uc.embedding.Generate(ctx, content)
	if err != nil {
		logger.ErrorContext(ctx, "CreateMoment: embedding failed", "error", err)
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
	logger := logging.FromContext(ctx)
	allMoments, err := uc.moments.ListByUserID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "CreateMoment: list history failed", "user_id", userID, "error", err)
		return nil, fmt.Errorf("list history: %w", err)
	}
	logger.DebugContext(ctx, "CreateMoment: loaded history", "user_id", userID, "total_moments", len(allMoments))

	history := excludeSelf(allMoments, moment.ID)
	if len(history) == 0 {
		logger.DebugContext(ctx, "CreateMoment: no history to match (first moment)")
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
