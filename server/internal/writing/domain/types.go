package domain

import "time"

// Trace is an aggregate root representing a continuous thinking session.
// A Trace is auto-created when the user writes their first Moment,
// and subsequent Moments can be added to the same Trace ("顺着再想想").
type Trace struct {
	ID         string
	UserID     string
	Motivation string // 'direct' | 'trace:<id>' | 'constellation:<id>'
	Stashed    bool
	CreatedAt  time.Time
}

// Moment is an entity representing a single piece of writing by the user.
// Each Moment belongs to exactly one Trace.
type Moment struct {
	ID         string
	TraceID    string
	UserID     string
	Content    string
	Embeddings []EmbeddingEntry
	CreatedAt  time.Time
}

// EmbeddingEntry holds an embedding vector with its model identifier.
// Multiple entries allow coexistence of embeddings from different model versions.
type EmbeddingEntry struct {
	Model     string    `json:"model"`
	Embedding []float32 `json:"embedding"`
}

// Echo is an entity representing a single historical echo match result.
// Persisted per-Moment: one Echo per CreateMoment call.
type Echo struct {
	ID               string
	MomentID         string
	UserID           string
	MatchedMomentIDs []string
	Similarities     []float64
	CreatedAt        time.Time
}

// Insight is an entity representing an AI-generated second-person observation
// for a specific Moment, based on its Echo.
type Insight struct {
	ID               string
	UserID           string
	MomentID         string
	EchoID           string
	Text             string
	RelatedMomentIDs []string
	CreatedAt        time.Time
}

// MatchedMoment is a value object pairing a historical Moment ID with its similarity score.
type MatchedMoment struct {
	MomentID   string
	Similarity float64
}

// TraceItem groups a Moment with its Echos and Insight for trace detail views.
type TraceItem struct {
	Moment  Moment
	Echos   []Echo
	Insight *Insight
}
