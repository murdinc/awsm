package models

import "time"

// Bucket represents an S3 Bucket
type Bucket struct {
	Name         string    `json:"name" awsmTable:"Name"`
	CreationDate time.Time `json:"createDate" awsmTable:"Created"`
}
