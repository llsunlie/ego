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

type Reader struct {
	queries *sqlc.Queries
}

func NewReader(queries *sqlc.Queries) *Reader {
	return &Reader{queries: queries}
}

var _ domain.MomentReader = (*Reader)(nil)
var _ domain.TraceReader = (*Reader)(nil)

func (r *Reader) GetByIDs(ctx context.Context, ids []string) ([]domain.Moment, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		pgIDs[i] = pgtype.UUID{Bytes: [16]byte(uid), Valid: true}
	}

	rows, err := r.queries.ListMomentsByIDs(ctx, pgIDs)
	if err != nil {
		return nil, err
	}

	moments := make([]domain.Moment, len(rows))
	for i, row := range rows {
		moments[i] = *toDomainMoment(row)
	}
	return moments, nil
}

func (r *Reader) GetByID(ctx context.Context, id string) (*domain.Moment, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetMomentByID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMomentNotFound
		}
		return nil, err
	}

	return toDomainMoment(row), nil
}

func (r *Reader) ListByUserID(ctx context.Context, userID string, cursor string, pageSize int32) ([]domain.Moment, string, bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, "", false, err
	}

	if pageSize <= 0 {
		pageSize = 20
	}

	var moments []domain.Moment

	var cursorTime pgtype.Timestamptz
	if cursor != "" {
		cursorMoment, err := r.GetByID(ctx, cursor)
		if err != nil {
			return nil, "", false, err
		}
		cursorTime = pgtype.Timestamptz{Time: cursorMoment.CreatedAt, Valid: true}
	} else {
		cursorTime = pgtype.Timestamptz{Time: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	}

	rows, err := r.queries.ListMomentsByUserIDCursor(ctx, sqlc.ListMomentsByUserIDCursorParams{
		UserID:     pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		Limit:      pageSize + 1,
		CursorTime: cursorTime,
	})
	if err != nil {
		return nil, "", false, err
	}
	moments = make([]domain.Moment, len(rows))
	for i, row := range rows {
		moments[i] = *toDomainMoment(row)
	}

	hasMore := int32(len(moments)) == pageSize+1
	if hasMore {
		moments = moments[:pageSize]
	}

	nextCursor := ""
	if len(moments) > 0 {
		nextCursor = moments[len(moments)-1].ID
	}

	return moments, nextCursor, hasMore, nil
}

func (r *Reader) RandomByUserID(ctx context.Context, userID string, count int32) ([]domain.Moment, error) {
	if count <= 0 {
		count = 3
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.RandomMomentsByUserID(ctx, sqlc.RandomMomentsByUserIDParams{
		UserID: pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		Limit:  count,
	})
	if err != nil {
		return nil, err
	}

	moments := make([]domain.Moment, len(rows))
	for i, row := range rows {
		moments[i] = *toDomainMoment(row)
	}
	return moments, nil
}

func (r *Reader) GetTraceByID(ctx context.Context, id string) (*domain.Trace, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetTraceByID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTraceNotFound
		}
		return nil, err
	}

	return toDomainTrace(row), nil
}

func (r *Reader) ListMomentsByTraceID(ctx context.Context, traceID string) ([]domain.Moment, error) {
	tid, err := uuid.Parse(traceID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListMomentsByTraceID(ctx, pgtype.UUID{Bytes: [16]byte(tid), Valid: true})
	if err != nil {
		return nil, err
	}

	moments := make([]domain.Moment, len(rows))
	for i, row := range rows {
		moments[i] = *toDomainMoment(row)
	}
	return moments, nil
}

func (r *Reader) ListTracesByUserID(ctx context.Context, userID string, cursor string, pageSize int32) ([]domain.Trace, string, bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, "", false, err
	}

	if pageSize <= 0 {
		pageSize = 20
	}

	var cursorTime pgtype.Timestamptz
	if cursor != "" {
		cursorTrace, err := r.GetTraceByID(ctx, cursor)
		if err != nil {
			return nil, "", false, err
		}
		cursorTime = pgtype.Timestamptz{Time: cursorTrace.CreatedAt, Valid: true}
	} else {
		cursorTime = pgtype.Timestamptz{Time: time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	}

	rows, err := r.queries.ListTracesByUserIDCursor(ctx, sqlc.ListTracesByUserIDCursorParams{
		UserID:     pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		Limit:      pageSize + 1,
		CursorTime: cursorTime,
	})
	if err != nil {
		return nil, "", false, err
	}

	traces := make([]domain.Trace, len(rows))
	traceIDs := make([]pgtype.UUID, len(rows))
	for i, row := range rows {
		traces[i] = *toDomainTrace(row)
		traceIDs[i] = row.ID
	}

	// Batch-fetch first moment content per trace.
	if len(traceIDs) > 0 {
		firstMoments, err := r.queries.FirstMomentsByTraceIDs(ctx, traceIDs)
		if err != nil {
			return nil, "", false, err
		}
		contentByTrace := make(map[string]string, len(firstMoments))
		for _, m := range firstMoments {
			tid, _ := uuid.FromBytes(m.TraceID.Bytes[:])
			contentByTrace[tid.String()] = m.Content
		}
		for i := range traces {
			traces[i].FirstMomentContent = contentByTrace[traces[i].ID]
		}
	}

	hasMore := int32(len(traces)) == pageSize+1
	if hasMore {
		traces = traces[:pageSize]
	}

	nextCursor := ""
	if len(traces) > 0 {
		nextCursor = traces[len(traces)-1].ID
	}

	return traces, nextCursor, hasMore, nil
}
