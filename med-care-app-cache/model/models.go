package model

import "time"

type FeedResponse struct {
	Articles []Article      `json:"articles"`
	Steps    []CarePlanStep `json:"steps"`
}

type Article struct {
	ID          int64
	Title       string
	Content     string
	Source      string
	Type        string
	PublishedAt time.Time
	IsRead      bool
	IsSaved     bool
	Relevance   float64
}

type CarePlanStep struct {
	ID          int64
	Type        string
	Title       string
	Description string
	Status      string
	DueDate     time.Time
	CompletedAt *time.Time
	Metadata    []byte
	OrderNumber int
}
