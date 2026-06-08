package app

import (
	"context"

	"ego-server/internal/setting/domain"
)

// GetProfileUseCase retrieves the current user's profile information.
type GetProfileUseCase struct {
	userReader domain.UserReader
}

func NewGetProfileUseCase(userReader domain.UserReader) *GetProfileUseCase {
	return &GetProfileUseCase{userReader: userReader}
}

// ProfileResult holds the profile data returned to the caller.
type ProfileResult struct {
	Phone     string
	CreatedAt int64 // unix timestamp ms
}

func (uc *GetProfileUseCase) GetProfile(ctx context.Context, userID string) (*ProfileResult, error) {
	user, err := uc.userReader.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ProfileResult{
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt.UnixMilli(),
	}, nil
}
