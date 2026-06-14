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
