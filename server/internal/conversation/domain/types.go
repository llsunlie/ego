package domain

import "time"

type ChatSession struct {
	ID               string
	UserID           string
	StarID           string

	CreatedAt        time.Time
}

type ChatMessage struct {
	ID                string
	UserID            string
	SessionID         string
	Role              string // "user" | "past_self"
	Content           string
	ReferencedMoments []MomentReference
	CreatedAt         time.Time
}

type MomentReference struct {
	Date    string // "X月X日"
	Snippet string // 原话片段
}
