package app

import (
	"context"

	"ego-server/internal/timeline/domain"
	writingdomain "ego-server/internal/writing/domain"
)

// GetRandomMomentsUseCase returns a random set of historical Moments.
type GetRandomMomentsUseCase struct {
	moments domain.MomentReader
}

func NewGetRandomMomentsUseCase(moments domain.MomentReader) *GetRandomMomentsUseCase {
	return &GetRandomMomentsUseCase{moments: moments}
}

type GetRandomMomentsInput struct {
	UserID string
	Count  int32
}

type GetRandomMomentsOutput struct {
	Moments []writingdomain.Moment
}

func (uc *GetRandomMomentsUseCase) Execute(ctx context.Context, input GetRandomMomentsInput) (*GetRandomMomentsOutput, error) {
	count := input.Count
	if count <= 0 {
		count = 3
	}

	moments, err := uc.moments.RandomByUserID(ctx, input.UserID, count)
	if err != nil {
		return nil, err
	}

	return &GetRandomMomentsOutput{
		Moments: moments,
	}, nil
}
