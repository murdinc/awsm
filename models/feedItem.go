package models

import "time"

type FeedItem struct {
	Date     time.Time `json:"date" awsmTable:"Posted"`
	Title    string    `json:"title" awsmTable:"Title"`
	Summary  string    `json:"summary"`
	Category string    `json:"category"`
	ID       string    `json:"guid"`
	Link     string    `json:"link" awsmLink:"Title"`
}
