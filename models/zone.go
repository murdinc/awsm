package models

// AZ represents an Availability Zone
type AZ struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	State  string `json:"state"`
}
