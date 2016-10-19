package models

// SimpleDBDomain represents a SimpleDB Domain
type SimpleDBDomain struct {
	Name   string `json:"name" awsmTable:"Name"`
	Region string `json:"region" awsmTable:"Region"`
}
