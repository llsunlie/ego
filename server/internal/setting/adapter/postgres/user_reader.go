package postgres

import (
	"context"
	"errors"

	"ego-server/internal/platform/postgres/sqlc"
	settingdomain "ego-server/internal/setting/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserReader implements setting/domain.UserReader using sqlc.
type UserReader struct {
	queries *sqlc.Queries
}

func NewUserReader(queries *sqlc.Queries) *UserReader {
	return &UserReader{queries: queries}
}

func (r *UserReader) FindByID(ctx context.Context, id string) (*settingdomain.UserInfo, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, settingdomain.ErrUserNotFound
	}

	// Convert [16]byte to pgtype.UUID
	var arr [16]byte
	copy(arr[:], uid[:])

	row, err := r.queries.GetUserByID(ctx, pgtype.UUID{Bytes: arr, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, settingdomain.ErrUserNotFound
		}
		return nil, err
	}

	return &settingdomain.UserInfo{
		Phone:     row.Phone,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}
