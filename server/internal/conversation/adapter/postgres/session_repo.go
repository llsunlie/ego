package postgres

import (
	"context"
	"errors"

	"ego-server/internal/conversation/domain"
	"ego-server/internal/platform/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionRepository struct {
	queries *sqlc.Queries
}

func NewSessionRepository(queries *sqlc.Queries) *SessionRepository {
	return &SessionRepository{queries: queries}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.ChatSession) error {
	uid, err := uuid.Parse(session.ID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(session.UserID)
	if err != nil {
		return err
	}
	starID, err := uuid.Parse(session.StarID)
	if err != nil {
		return err
	}

	return r.queries.CreateChatSession(ctx, sqlc.CreateChatSessionParams{
		ID:        pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		UserID:    pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		StarID:    pgtype.UUID{Bytes: [16]byte(starID), Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: session.CreatedAt, Valid: true},
	})
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (*domain.ChatSession, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetChatSessionByID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrChatSessionNotFound
		}
		return nil, err
	}

	return toDomainChatSession(row), nil
}

func toDomainChatSession(row sqlc.ChatSession) *domain.ChatSession {
	id, _ := uuid.FromBytes(row.ID.Bytes[:])
	userID, _ := uuid.FromBytes(row.UserID.Bytes[:])
	starID, _ := uuid.FromBytes(row.StarID.Bytes[:])

	return &domain.ChatSession{
		ID:        id.String(),
		UserID:    userID.String(),
		StarID:    starID.String(),
		CreatedAt: row.CreatedAt.Time,
	}
}
