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

	ConstellationProfileStatusReady  = "ready"
	ConstellationProfileStatusFailed = "failed"

	ConstellationMatchTypePrimary   = "primary"
	ConstellationMatchTypeSecondary = "secondary"
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
	PatternTags            []string
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

type ConstellationProfile struct {
	ConstellationID string
	UserID          string
	Topic           string
	Summary         string
	Keywords        []string
	Emotions        []string
	Scenes          []string
	CentralPattern  string
	PatternTags     []string
	ProfileText     string
	TraceCount      float64
	MomentCount     float64
	Status          string
	LastError       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ConstellationProfileVector struct {
	ConstellationID   string
	UserID            string
	Model             string
	Dim               int
	ProfileEmbedding  []float32
	CentroidEmbedding []float32
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ConstellationMembership struct {
	ConstellationID string
	StarID          string
	TraceID         string
	UserID          string
	MatchScore      float64
	MatchType       string
	MatchDimensions []string
	MatchReason     string
	Weight          float64
	CreatedAt       time.Time
}

type ConstellationProfileCandidate struct {
	Profile ConstellationProfile
	Vector  ConstellationProfileVector
}
