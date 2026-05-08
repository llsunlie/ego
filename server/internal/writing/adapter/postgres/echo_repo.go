package postgres

import (
	"context"
	"errors"
	"time"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type EchoRepository struct {
	queries *sqlc.Queries
}

func NewEchoRepository(queries *sqlc.Queries) *EchoRepository {
	return &EchoRepository{queries: queries}
}

func (r *EchoRepository) Create(ctx context.Context, echo *domain.Echo) error {
	uid, err := uuid.Parse(echo.ID)
	if err != nil {
		return err
	}
	momentID, err := uuid.Parse(echo.MomentID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(echo.UserID)
	if err != nil {
		return err
	}

	matchedIDs := make([]pgtype.UUID, len(echo.MatchedMomentIDs))
	for i, id := range echo.MatchedMomentIDs {
		u, err := uuid.Parse(id)
		if err != nil {
			return err
		}
		matchedIDs[i] = pgtype.UUID{Bytes: [16]byte(u), Valid: true}
	}

	now := time.Now()
	echo.CreatedAt = now

	return r.queries.CreateEcho(ctx, sqlc.CreateEchoParams{
		ID:               pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		MomentID:         pgtype.UUID{Bytes: [16]byte(momentID), Valid: true},
		UserID:           pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		MatchedMomentIds: matchedIDs,
		Similarities:     echo.Similarities,
		CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
	})
}

func (r *EchoRepository) FindByMomentID(ctx context.Context, momentID string) (*domain.Echo, error) {
	mid, err := uuid.Parse(momentID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetEchoByMomentID(ctx, pgtype.UUID{Bytes: [16]byte(mid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEchoNotFound
		}
		return nil, err
	}

	return toDomainEcho(row), nil
}

func toDomainEcho(row sqlc.Echo) *domain.Echo {
	id, _ := uuid.FromBytes(row.ID.Bytes[:])
	momentID, _ := uuid.FromBytes(row.MomentID.Bytes[:])
	userID, _ := uuid.FromBytes(row.UserID.Bytes[:])

	matchedIDs := make([]string, len(row.MatchedMomentIds))
	for i, mid := range row.MatchedMomentIds {
		u, _ := uuid.FromBytes(mid.Bytes[:])
		matchedIDs[i] = u.String()
	}

	return &domain.Echo{
		ID:               id.String(),
		MomentID:         momentID.String(),
		UserID:           userID.String(),
		MatchedMomentIDs: matchedIDs,
		Similarities:     row.Similarities,
		CreatedAt:        row.CreatedAt.Time,
	}
}
