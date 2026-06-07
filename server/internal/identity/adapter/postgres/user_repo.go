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

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	row, err := r.queries.GetUserByPhone(ctx, phone)
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
		Phone:        phone,
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
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return r.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		PasswordHash: passwordHash,
	})
}
