package postgres

import (
	"context"
	"encoding/json"

	"ego-server/internal/conversation/domain"
	"ego-server/internal/platform/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type MessageRepository struct {
	queries *sqlc.Queries
}

func NewMessageRepository(queries *sqlc.Queries) *MessageRepository {
	return &MessageRepository{queries: queries}
}

func (r *MessageRepository) Create(ctx context.Context, msg *domain.ChatMessage) error {
	uid, err := uuid.Parse(msg.ID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(msg.UserID)
	if err != nil {
		return err
	}
	sessionID, err := uuid.Parse(msg.SessionID)
	if err != nil {
		return err
	}

	var refsJSON []byte
	if len(msg.ReferencedMoments) > 0 {
		refs := make([]map[string]string, len(msg.ReferencedMoments))
		for i, ref := range msg.ReferencedMoments {
			refs[i] = map[string]string{
				"date":    ref.Date,
				"snippet": ref.Snippet,
			}
		}
		refsJSON, err = json.Marshal(refs)
		if err != nil {
			return err
		}
	}

	return r.queries.CreateChatMessage(ctx, sqlc.CreateChatMessageParams{
		ID:                pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		UserID:            pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		SessionID:         pgtype.UUID{Bytes: [16]byte(sessionID), Valid: true},
		Role:              msg.Role,
		Content:           msg.Content,
		ReferencedMoments: refsJSON,
		CreatedAt:         pgtype.Timestamptz{Time: msg.CreatedAt, Valid: true},
	})
}

func (r *MessageRepository) ListBySessionID(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListMessagesBySessionID(ctx, pgtype.UUID{Bytes: [16]byte(sid), Valid: true})
	if err != nil {
		return nil, err
	}

	result := make([]domain.ChatMessage, len(rows))
	for i, row := range rows {
		result[i] = *toDomainChatMessage(row)
	}
	return result, nil
}

func toDomainChatMessage(row sqlc.ChatMessage) *domain.ChatMessage {
	id, _ := uuid.FromBytes(row.ID.Bytes[:])
	userID, _ := uuid.FromBytes(row.UserID.Bytes[:])
	sessionID, _ := uuid.FromBytes(row.SessionID.Bytes[:])

	var refs []domain.MomentReference
	if len(row.ReferencedMoments) > 0 {
		var raw []struct {
			Date    string `json:"date"`
			Snippet string `json:"snippet"`
		}
		if err := json.Unmarshal(row.ReferencedMoments, &raw); err == nil {
			refs = make([]domain.MomentReference, len(raw))
			for i, r := range raw {
				refs[i] = domain.MomentReference{
					Date:    r.Date,
					Snippet: r.Snippet,
				}
			}
		}
	}

	return &domain.ChatMessage{
		ID:                id.String(),
		UserID:            userID.String(),
		SessionID:         sessionID.String(),
		Role:              row.Role,
		Content:           row.Content,
		ReferencedMoments: refs,
		CreatedAt:         row.CreatedAt.Time,
	}
}
