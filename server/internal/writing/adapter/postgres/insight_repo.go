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

type InsightRepository struct {
	queries *sqlc.Queries
}

func NewInsightRepository(queries *sqlc.Queries) *InsightRepository {
	return &InsightRepository{queries: queries}
}

func (r *InsightRepository) Create(ctx context.Context, insight *domain.Insight) error {
	uid, err := uuid.Parse(insight.ID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(insight.UserID)
	if err != nil {
		return err
	}
	momentID, err := uuid.Parse(insight.MomentID)
	if err != nil {
		return err
	}
	echoID, err := uuid.Parse(insight.EchoID)
	if err != nil {
		return err
	}

	relatedIDs := make([]pgtype.UUID, len(insight.RelatedMomentIDs))
	for i, id := range insight.RelatedMomentIDs {
		u, err := uuid.Parse(id)
		if err != nil {
			return err
		}
		relatedIDs[i] = pgtype.UUID{Bytes: [16]byte(u), Valid: true}
	}

	now := time.Now()
	insight.CreatedAt = now

	return r.queries.CreateInsight(ctx, sqlc.CreateInsightParams{
		ID:               pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		UserID:           pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		MomentID:         pgtype.UUID{Bytes: [16]byte(momentID), Valid: true},
		EchoID:           pgtype.UUID{Bytes: [16]byte(echoID), Valid: true},
		Text:             insight.Text,
		RelatedMomentIds: relatedIDs,
		CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
	})
}

func (r *InsightRepository) FindByMomentID(ctx context.Context, momentID string) (*domain.Insight, error) {
	mid, err := uuid.Parse(momentID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetInsightByMomentID(ctx, pgtype.UUID{Bytes: [16]byte(mid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInsightNotFound
		}
		return nil, err
	}

	return toDomainInsight(row), nil
}

func toDomainInsight(row sqlc.Insight) *domain.Insight {
	id, _ := uuid.FromBytes(row.ID.Bytes[:])
	userID, _ := uuid.FromBytes(row.UserID.Bytes[:])
	momentID, _ := uuid.FromBytes(row.MomentID.Bytes[:])
	echoID, _ := uuid.FromBytes(row.EchoID.Bytes[:])

	relatedIDs := make([]string, len(row.RelatedMomentIds))
	for i, rid := range row.RelatedMomentIds {
		u, _ := uuid.FromBytes(rid.Bytes[:])
		relatedIDs[i] = u.String()
	}

	return &domain.Insight{
		ID:               id.String(),
		UserID:           userID.String(),
		MomentID:         momentID.String(),
		EchoID:           echoID.String(),
		Text:             row.Text,
		RelatedMomentIDs: relatedIDs,
		CreatedAt:        row.CreatedAt.Time,
	}
}
