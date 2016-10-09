package models

type SimpleDBDomain struct {
	Name   string `json:"name" awsmTable:"Name"`
	Region string `json:"region" awsmTable:"Region"`
}
