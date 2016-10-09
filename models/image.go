package models

import "time"

type Image struct {
	Name         string    `json:"name" awsmTable:"Name"`
	Class        string    `json:"class" awsmTable:"Class"`
	CreationDate time.Time `json:"creationDate"`
	CreatedHuman string    `json:"createdHuman" awsmTable:"Created"`
	ImageId      string    `json:"imageId" awsmTable:"Image ID"`
	State        string    `json:"state" awsmTable:"State"`
	Root         string    `json:"root" awsmTable:"Root"`
	SnapshotId   string    `json:"snapshotId" awsmTable:"Snapshot ID"`
	VolumeSize   string    `json:"volumeSize" awsmTable:"Volume Size"`
	Region       string    `json:"region" awsmTable:"Region"`
}
