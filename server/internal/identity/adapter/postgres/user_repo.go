package postgres

import (
	"context"
	"errors"
	"time"

	"ego-server/internal/identity/domain"
	"ego-server/internal/platform/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

func (r *UserRepository) FindByAccount(ctx context.Context, account string) (*domain.User, error) {
	row, err := r.queries.GetUserByAccount(ctx, account)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	id, err := uuid.FromBytes(row.ID.Bytes[:])
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:           id.String(),
		Account:      account,
		PasswordHash: row.PasswordHash,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	uid, err := uuid.Parse(user.ID)
	if err != nil {
		return err
	}

	now := time.Now()
	user.CreatedAt = now

	return r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		Account:      user.Account,
		PasswordHash: user.PasswordHash,
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
}
