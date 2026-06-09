package domain

import "time"

// Feedback represents a user-submitted feedback.
type Feedback struct {
	ID        string
	UserID    string
	Content   string
	CreatedAt time.Time
}
