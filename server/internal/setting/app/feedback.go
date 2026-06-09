package app

import (
	"context"
	"strings"
	"time"

	"ego-server/internal/setting/domain"
)

// IDGenerator creates unique identifiers.
type IDGenerator interface {
	New() string
}

// SubmitFeedbackUseCase handles user feedback submission.
type SubmitFeedbackUseCase struct {
	feedbackWriter domain.FeedbackWriter
	idGenerator    IDGenerator
}

func NewSubmitFeedbackUseCase(
	feedbackWriter domain.FeedbackWriter,
	idGenerator IDGenerator,
) *SubmitFeedbackUseCase {
	return &SubmitFeedbackUseCase{
		feedbackWriter: feedbackWriter,
		idGenerator:    idGenerator,
	}
}

// FeedbackResult holds the result of a successful feedback submission.
type FeedbackResult struct {
	ID        string
	CreatedAt int64 // unix timestamp ms
}

// Submit validates and persists user feedback.
func (uc *SubmitFeedbackUseCase) Submit(ctx context.Context, userID, content string) (*FeedbackResult, error) {
	if strings.TrimSpace(content) == "" {
		return nil, domain.ErrFeedbackEmpty
	}

	fb := &domain.Feedback{
		ID:        uc.idGenerator.New(),
		UserID:    userID,
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now(),
	}

	if err := uc.feedbackWriter.Save(ctx, fb); err != nil {
		return nil, err
	}

	return &FeedbackResult{
		ID:        fb.ID,
		CreatedAt: fb.CreatedAt.UnixMilli(),
	}, nil
}
