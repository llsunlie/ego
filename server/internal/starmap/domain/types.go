package domain

import "time"

type Star struct {
	ID        string
	UserID    string
	TraceID   string
	Topic     string
	CreatedAt time.Time
}

type Constellation struct {
	ID                   string
	UserID               string
	Topic                string
	TopicEmbedding       []float32
	Name                 string
	ConstellationInsight string
	StarIDs              []string
	TopicPrompts         []string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

const (
	TraceProfileStatusReady    = "ready"
	TraceProfileStatusFallback = "fallback"
	TraceProfileStatusFailed   = "failed"
)

type TraceProfile struct {
	TraceID                string
	UserID                 string
	Topic                  string
	Summary                string
	Keywords               []string
	Emotions               []string
	Scenes                 []string
	CentralPattern         string
	RepresentativeMomentID string
	ProfileText            string
	Status                 string
	RetryCount             int
	LastError              string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type TraceProfileVector struct {
	TraceID   string
	UserID    string
	Model     string
	Dim       int
	Embedding []float32
	CreatedAt time.Time
	UpdatedAt time.Time
}
