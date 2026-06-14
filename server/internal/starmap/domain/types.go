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
	ConstellationID  string
	UserID           string
	Topic            string
	Summary          string
	Keywords         []string
	Emotions         []string
	Scenes           []string
	CentralPattern   string
	PatternTags      []string
	ThemeCode        string
	ThemeLabel       string
	ThemeDescription string
	ThemeExamples    []string
	ProfileText      string
	TraceCount       float64
	MomentCount      float64
	Status           string
	LastError        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
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

type ConstellationProfileSparseCandidate struct {
	ConstellationID string
	Score           float64
	MatchedFields   []string
	Preview         string
}

type ConstellationProfileRefineInput struct {
	Existing             ConstellationProfile
	RuleMerged           ConstellationProfile
	IncomingTraceProfile TraceProfile
	RepresentativeMoment string
	Trigger              int
}

type ConstellationProfileRefinement struct {
	Profile          ConstellationProfile
	Model            string
	Dim              int
	ProfileEmbedding []float32
	DisplayName      string
}

const (
	ConstellationBorderlineDecisionUseExisting = "use_existing"
	ConstellationBorderlineDecisionSuggestNew  = "suggest_new"
)

type ConstellationBorderlineCandidate struct {
	ConstellationID    string
	Topic              string
	Summary            string
	Keywords           []string
	Emotions           []string
	Scenes             []string
	CentralPattern     string
	PatternTags        []string
	ThemeCode          string
	ThemeLabel         string
	ThemeDescription   string
	ThemeExamples      []string
	Score              float64
	ProfileSimilarity  float64
	CentroidSimilarity float64
	KeywordOverlap     float64
	SceneOverlap       float64
	EmotionOverlap     float64
	PatternTagsOverlap float64
	MatchedKeywords    []string
	MatchedScenes      []string
	MatchedEmotions    []string
	MatchedPatternTags []string
	Dimensions         []string
	Reason             string
}

type ConstellationBorderlineJudgeInput struct {
	TraceProfile         TraceProfile
	RepresentativeMoment string
	Candidates           []ConstellationBorderlineCandidate
}

type ConstellationBorderlineJudgement struct {
	Decision                  string
	ConstellationID           string
	ThemeCode                 string
	Confidence                float64
	SharedSituation           string
	MatchDimensions           []string
	Reason                    string
	SuggestedThemeCode        string
	SuggestedThemeLabel       string
	SuggestedThemeDescription string
	Primary                   *ConstellationBorderlineSelection
	Secondary                 []ConstellationBorderlineSelection
}

type ConstellationBorderlineSelection struct {
	ConstellationID string
	ThemeCode       string
	Confidence      float64
	SharedTheme     string
	MatchDimensions []string
	Reason          string
}
