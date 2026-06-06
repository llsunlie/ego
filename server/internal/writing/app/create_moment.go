package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ego-server/internal/platform/logging"
	"ego-server/internal/writing/domain"
)

// CreateMomentUseCase orchestrates the creation of a Moment (and optionally a Trace),
// embedding generation, and echo matching.
type CreateMomentUseCase struct {
	traces           domain.TraceRepository
	moments          domain.MomentRepository
	echoCandidates   domain.EchoCandidateReader
	searchIndexer    domain.MomentSearchIndexer
	sparseCandidates domain.EchoSparseCandidateReader
	echos            domain.EchoRepository
	embedding        domain.EmbeddingGenerator
	echo             domain.EchoMatcher
	ids              IDGenerator
	echoRecallTopK   int32
	echoSparseTopK   int32
	echoHybridRRFK   int
}

func NewCreateMomentUseCase(
	traces domain.TraceRepository,
	moments domain.MomentRepository,
	echos domain.EchoRepository,
	embedding domain.EmbeddingGenerator,
	echo domain.EchoMatcher,
	ids IDGenerator,
) *CreateMomentUseCase {
	return NewCreateMomentUseCaseWithCandidates(
		traces,
		moments,
		listByUserCandidateReader{moments: moments},
		echos,
		embedding,
		echo,
		ids,
		10,
	)
}

func NewCreateMomentUseCaseWithCandidates(
	traces domain.TraceRepository,
	moments domain.MomentRepository,
	echoCandidates domain.EchoCandidateReader,
	echos domain.EchoRepository,
	embedding domain.EmbeddingGenerator,
	echo domain.EchoMatcher,
	ids IDGenerator,
	echoRecallTopK int32,
) *CreateMomentUseCase {
	return NewCreateMomentUseCaseWithHybridCandidates(
		traces,
		moments,
		echoCandidates,
		nil,
		nil,
		echos,
		embedding,
		echo,
		ids,
		echoRecallTopK,
		0,
		60,
	)
}

func NewCreateMomentUseCaseWithHybridCandidates(
	traces domain.TraceRepository,
	moments domain.MomentRepository,
	echoCandidates domain.EchoCandidateReader,
	searchIndexer domain.MomentSearchIndexer,
	sparseCandidates domain.EchoSparseCandidateReader,
	echos domain.EchoRepository,
	embedding domain.EmbeddingGenerator,
	echo domain.EchoMatcher,
	ids IDGenerator,
	echoRecallTopK int32,
	echoSparseTopK int32,
	echoHybridRRFK int,
) *CreateMomentUseCase {
	if echoRecallTopK <= 0 {
		echoRecallTopK = 10
	}
	if echoHybridRRFK <= 0 {
		echoHybridRRFK = 60
	}
	return &CreateMomentUseCase{
		traces:           traces,
		moments:          moments,
		echoCandidates:   echoCandidates,
		searchIndexer:    searchIndexer,
		sparseCandidates: sparseCandidates,
		echos:            echos,
		embedding:        embedding,
		echo:             echo,
		ids:              ids,
		echoRecallTopK:   echoRecallTopK,
		echoSparseTopK:   echoSparseTopK,
		echoHybridRRFK:   echoHybridRRFK,
	}
}

type listByUserCandidateReader struct {
	moments domain.MomentRepository
}

func (r listByUserCandidateReader) FindNearestMoments(ctx context.Context, userID string, currentMomentID string, _ string, _ []float32, _ int32) ([]domain.Moment, error) {
	allMoments, err := r.moments.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var result []domain.Moment
	for _, m := range allMoments {
		if m.ID != currentMomentID {
			result = append(result, m)
		}
	}
	return result, nil
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
	logger.DebugContext(ctx, "writing create moment started", "content_len", len([]rune(input.Content)), "trace_id", input.TraceID, "motivation", input.Motivation)

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
			logger.ErrorContext(ctx, "writing create moment trace create failed", "error", err)
			return nil, fmt.Errorf("create trace: %w", err)
		}
		traceID = trace.ID
		newTrace = true
		logger.DebugContext(ctx, "writing create moment trace created", "trace_id", traceID, "motivation", motivation)
	} else {
		existing, err := uc.traces.GetByID(ctx, traceID)
		if err != nil {
			logger.ErrorContext(ctx, "writing create moment trace load failed", "trace_id", traceID, "error", err)
			return nil, fmt.Errorf("get trace: %w", err)
		}
		if existing.UserID != userID {
			logger.WarnContext(ctx, "writing create moment trace ownership mismatch", "trace_id", traceID, "trace_user", existing.UserID, "caller_user", userID)
			return nil, fmt.Errorf("trace does not belong to user")
		}
		logger.DebugContext(ctx, "writing create moment appending to trace", "trace_id", traceID)
	}

	moment, err := uc.createMoment(ctx, userID, traceID, input.Content)
	if err != nil {
		if newTrace {
			logger.WarnContext(ctx, "writing create moment rolling back trace", "trace_id", traceID, "error", err)
			_ = uc.traces.Delete(ctx, traceID)
		}
		return nil, fmt.Errorf("create moment: %w", err)
	}
	logger.DebugContext(ctx, "writing create moment created", "moment_id", moment.ID, "embedding_count", len(moment.Embeddings))

	echo, err := uc.matchEcho(ctx, moment, userID)
	if err != nil {
		logger.ErrorContext(ctx, "writing create moment echo matching failed", "moment_id", moment.ID, "error", err)
		return nil, fmt.Errorf("match echo: %w", err)
	}

	if echo != nil {
		logger.InfoContext(ctx, "writing create moment echo found", "moment_id", moment.ID, "echo_id", echo.ID, "matched_count", len(echo.MatchedMomentIDs))
	} else {
		logger.DebugContext(ctx, "writing create moment no echo", "moment_id", moment.ID)
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
	logger.DebugContext(ctx, "writing create moment embedding started", "content_len", len([]rune(content)))
	embeddings, err := uc.embedding.Generate(ctx, content)
	if err != nil {
		logger.ErrorContext(ctx, "writing create moment embedding failed", "error", err)
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
	if uc.searchIndexer != nil {
		if err := uc.searchIndexer.IndexMoment(ctx, *moment); err != nil {
			logger.WarnContext(ctx, "writing create moment search index failed", "moment_id", moment.ID, "error", err)
		}
	}
	return moment, nil
}

func (uc *CreateMomentUseCase) matchEcho(ctx context.Context, moment *domain.Moment, userID string) (*domain.Echo, error) {
	logger := logging.FromContext(ctx)
	if len(moment.Embeddings) == 0 {
		logger.DebugContext(ctx, "writing echo matching skipped no embedding", "moment_id", moment.ID)
		return nil, nil
	}

	currentEmbedding := moment.Embeddings[0]
	denseCandidates := make(chan echoCandidateResult, 1)
	sparseCandidates := make(chan []domain.Moment, 1)

	go func() {
		denseHistory, err := uc.echoCandidates.FindNearestMoments(
			ctx,
			userID,
			moment.ID,
			currentEmbedding.Model,
			currentEmbedding.Embedding,
			uc.echoRecallTopK,
		)
		denseCandidates <- echoCandidateResult{moments: denseHistory, err: err}
	}()
	go func() {
		sparseCandidates <- uc.loadSparseEchoCandidates(ctx, *moment)
	}()

	denseResult := <-denseCandidates
	sparseHistory := <-sparseCandidates
	if denseResult.err != nil {
		logger.ErrorContext(ctx, "writing echo dense candidates failed", "user_id", userID, "moment_id", moment.ID, "error", denseResult.err)
		return nil, fmt.Errorf("nearest echo candidates: %w", denseResult.err)
	}

	denseHistory := denseResult.moments
	history := denseHistory
	if len(sparseHistory) > 0 {
		history = mergeEchoCandidatesRRF(denseHistory, sparseHistory, uc.echoHybridRRFK, maxInt(len(denseHistory), int(uc.echoSparseTopK)))
	}
	logger.DebugContext(ctx, "writing echo recall candidates",
		"user_id", userID,
		"moment_id", moment.ID,
		"current_preview", echoLogPreview(moment.Content),
		"dense_candidate_count", len(denseHistory),
		"sparse_candidate_count", len(sparseHistory),
		"fused_candidate_count", len(history),
		"dense_top_k", uc.echoRecallTopK,
		"sparse_top_k", uc.echoSparseTopK,
		"rrf_k", uc.echoHybridRRFK,
		"dense_candidates", echoLogCandidates(denseHistory),
		"es_candidates", echoLogCandidates(sparseHistory),
		"fused_candidates", echoLogCandidates(history),
	)

	if len(history) == 0 {
		logger.DebugContext(ctx, "writing echo matching skipped no history")
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
	logger.DebugContext(ctx, "writing echo final matches",
		"moment_id", moment.ID,
		"match_count", len(matches),
		"matches", echoLogMatches(matches, history),
	)

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

type echoCandidateResult struct {
	moments []domain.Moment
	err     error
}

type momentByIDsReader interface {
	GetByIDs(ctx context.Context, ids []string) ([]domain.Moment, error)
}

func (uc *CreateMomentUseCase) loadSparseEchoCandidates(ctx context.Context, moment domain.Moment) []domain.Moment {
	logger := logging.FromContext(ctx)
	if uc.sparseCandidates == nil || uc.echoSparseTopK <= 0 {
		return nil
	}
	ids, err := uc.sparseCandidates.SearchMomentIDs(ctx, moment, uc.echoSparseTopK)
	if err != nil {
		logger.WarnContext(ctx, "writing echo sparse candidates failed", "moment_id", moment.ID, "error", err)
		return nil
	}
	if len(ids) == 0 {
		logger.DebugContext(ctx, "writing echo sparse candidates empty",
			"moment_id", moment.ID,
			"sparse_top_k", uc.echoSparseTopK,
		)
		return nil
	}
	reader, ok := uc.moments.(momentByIDsReader)
	if !ok {
		logger.WarnContext(ctx, "writing echo sparse candidates skipped no id reader", "moment_id", moment.ID)
		return nil
	}
	moments, err := reader.GetByIDs(ctx, ids)
	if err != nil {
		logger.WarnContext(ctx, "writing echo sparse candidates load failed", "moment_id", moment.ID, "error", err)
		return nil
	}
	missing := maxInt(len(ids)-len(moments), 0)
	logger.DebugContext(ctx, "writing echo sparse candidates loaded",
		"moment_id", moment.ID,
		"raw_id_count", len(ids),
		"loaded_count", len(moments),
		"missing_count", missing,
		"raw_ids", logStringList(ids, echoLogCandidateLimit),
	)
	return orderMomentsByIDs(ids, moments)
}

const echoLogCandidateLimit = 5
const echoLogPreviewRunes = 48

func echoLogCandidates(moments []domain.Moment) []map[string]any {
	limit := echoLogCandidateLimit
	if len(moments) < limit {
		limit = len(moments)
	}
	items := make([]map[string]any, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, echoLogCandidate(moments[i], i+1))
	}
	return items
}

func echoLogMatches(matches []domain.MatchedMoment, history []domain.Moment) []map[string]any {
	byID := make(map[string]domain.Moment, len(history))
	for _, moment := range history {
		byID[moment.ID] = moment
	}
	items := make([]map[string]any, 0, len(matches))
	for i, match := range matches {
		item := map[string]any{
			"rank":       i + 1,
			"moment_id":  match.MomentID,
			"similarity": match.Similarity,
		}
		if moment, ok := byID[match.MomentID]; ok {
			item["trace_id"] = moment.TraceID
			item["content_preview"] = echoLogPreview(moment.Content)
			if !moment.CreatedAt.IsZero() {
				item["created_at"] = moment.CreatedAt.Format(time.RFC3339)
			}
		}
		items = append(items, item)
	}
	return items
}

func echoLogCandidate(moment domain.Moment, rank int) map[string]any {
	item := map[string]any{
		"rank":            rank,
		"moment_id":       moment.ID,
		"trace_id":        moment.TraceID,
		"content_preview": echoLogPreview(moment.Content),
	}
	if !moment.CreatedAt.IsZero() {
		item["created_at"] = moment.CreatedAt.Format(time.RFC3339)
	}
	return item
}

func echoLogPreview(content string) string {
	content = strings.Join(strings.Fields(content), " ")
	runes := []rune(content)
	if len(runes) <= echoLogPreviewRunes {
		return content
	}
	return string(runes[:echoLogPreviewRunes]) + "..."
}

func logStringList(values []string, limit int) []string {
	if limit > 0 && len(values) > limit {
		values = values[:limit]
	}
	return append([]string(nil), values...)
}
