package postgres

import (
	"context"
	"errors"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/starmap/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ConstellationRepository struct {
	queries *sqlc.Queries
}

func NewConstellationRepository(queries *sqlc.Queries) *ConstellationRepository {
	return &ConstellationRepository{queries: queries}
}

func (r *ConstellationRepository) Create(ctx context.Context, c *domain.Constellation) error {
	uid, err := uuid.Parse(c.ID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(c.UserID)
	if err != nil {
		return err
	}

	starIDs := make([]pgtype.UUID, len(c.StarIDs))
	for i, id := range c.StarIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return err
		}
		starIDs[i] = pgtype.UUID{Bytes: [16]byte(uid), Valid: true}
	}

	return r.queries.CreateConstellation(ctx, sqlc.CreateConstellationParams{
		ID:                   pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		UserID:               pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		Topic:                c.Topic,
		TopicEmbedding:       c.TopicEmbedding,
		Name:                 c.Name,
		ConstellationInsight: c.ConstellationInsight,
		StarIds:              starIDs,
		TopicPrompts:         c.TopicPrompts,
		CreatedAt:            pgtype.Timestamptz{Time: c.CreatedAt, Valid: true},
		UpdatedAt:            pgtype.Timestamptz{Time: c.UpdatedAt, Valid: true},
	})
}

func (r *ConstellationRepository) Update(ctx context.Context, c *domain.Constellation) error {
	uid, err := uuid.Parse(c.ID)
	if err != nil {
		return err
	}

	starIDs := make([]pgtype.UUID, len(c.StarIDs))
	for i, id := range c.StarIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return err
		}
		starIDs[i] = pgtype.UUID{Bytes: [16]byte(uid), Valid: true}
	}

	return r.queries.UpdateConstellation(ctx, sqlc.UpdateConstellationParams{
		ID:                   pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		Topic:                c.Topic,
		TopicEmbedding:       c.TopicEmbedding,
		Name:                 c.Name,
		ConstellationInsight: c.ConstellationInsight,
		StarIds:              starIDs,
		TopicPrompts:         c.TopicPrompts,
		UpdatedAt:            pgtype.Timestamptz{Time: c.UpdatedAt, Valid: true},
	})
}

func (r *ConstellationRepository) FindAllByUserID(ctx context.Context, userID string) ([]domain.Constellation, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListConstellationsByUserID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		return nil, err
	}

	result := make([]domain.Constellation, len(rows))
	for i, row := range rows {
		c := toDomainConstellation(row)
		result[i] = *c
	}
	return result, nil
}

func (r *ConstellationRepository) FindByID(ctx context.Context, id string) (*domain.Constellation, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetConstellationByID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConstellationNotFound
		}
		return nil, err
	}

	return toDomainConstellation(row), nil
}

func (r *ConstellationRepository) FindByStarID(ctx context.Context, starID string) (*domain.Constellation, error) {
	uid, err := uuid.Parse(starID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetConstellationByStarID(ctx, []pgtype.UUID{{Bytes: [16]byte(uid), Valid: true}})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConstellationNotFound
		}
		return nil, err
	}

	return toDomainConstellation(row), nil
}

func toDomainConstellation(row sqlc.Constellation) *domain.Constellation {
	id, _ := uuid.FromBytes(row.ID.Bytes[:])
	userID, _ := uuid.FromBytes(row.UserID.Bytes[:])

	starIDs := make([]string, len(row.StarIds))
	for i, uid := range row.StarIds {
		sid, _ := uuid.FromBytes(uid.Bytes[:])
		starIDs[i] = sid.String()
	}

	return &domain.Constellation{
		ID:                   id.String(),
		UserID:               userID.String(),
		Topic:                row.Topic,
		TopicEmbedding:       row.TopicEmbedding,
		Name:                 row.Name,
		ConstellationInsight: row.ConstellationInsight,
		StarIDs:              starIDs,
		TopicPrompts:         row.TopicPrompts,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
	}
}
