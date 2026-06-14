package app

import (
	"context"

	"ego-server/internal/identity/domain"
)

type RefreshTokenUseCase struct {
	tokens   TokenIssuer
	verifier RefreshTokenVerifier
}

func NewRefreshTokenUseCase(tokens TokenIssuer, verifier RefreshTokenVerifier) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{tokens: tokens, verifier: verifier}
}

func (uc *RefreshTokenUseCase) Refresh(ctx context.Context, refreshToken string) (string, error) {
	userID, err := uc.verifier.Verify(refreshToken, "refresh")
	if err != nil {
		return "", domain.ErrInvalidRefreshToken
	}

	accessToken, err := uc.tokens.Issue(userID)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
