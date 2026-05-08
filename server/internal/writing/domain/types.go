package domain

import "time"

// Trace is an aggregate root representing a continuous thinking session.
// A Trace is auto-created when the user writes their first Moment,
// and subsequent Moments can be added to the same Trace ("顺着再想想").
type Trace struct {
	ID        string
	UserID    string
	Topic     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Moment is an entity representing a single piece of writing by the user.
// Each Moment belongs to exactly one Trace.
type Moment struct {
	ID        string
	TraceID   string
	UserID    string
	Content   string
	Embedding []float32
	Connected bool
	CreatedAt time.Time
}

// Echo is a value object representing the matched historical Moment
// along with alternative candidates for the user to explore.
type Echo struct {
	ID           string
	TargetMoment Moment
	Candidates   []Moment
	Similarity   float64
}

// Insight is an entity representing a current-session AI-generated observation.
// Writing generates this on-the-fly and returns it without persisting.
// Constellation-level insights are owned by the Starmap module.
type Insight struct {
	ID               string
	Text             string
	RelatedMomentIDs []string
}
