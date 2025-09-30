package model

import "time"

// Article represents a simple article record.
type Article struct {
    ID          int64      `json:"id"`
    Title       string     `json:"title"`
    Content     string     `json:"content"`
    Author      string     `json:"author"`
    PublishedAt *time.Time `json:"published_at,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
}
