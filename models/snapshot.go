package models

import "time"

type Snapshot struct {
	Name         string    `json:"name" awsmTable:"Name"`
	Class        string    `json:"class" awsmTable:"Class"`
	Description  string    `json:"description" awsmTable:"Description"`
	SnapshotId   string    `json:"snapshotId" awsmTable:"Snapshot ID"`
	VolumeId     string    `json:"volumeId" awsmTable:"Volume ID"`
	State        string    `json:"state" awsmTable:"State"`
	StartTime    time.Time `json:"startTime"`
	CreatedHuman string    `json:"createdHuman" awsmTable:"Created"`
	Progress     string    `json:"progress" awsmTable:"Progress"`
	VolumeSize   string    `json:"volumeSize" awsmTable:"Volume Size"`
	Region       string    `json:"region" awsmTable:"Region"`
}
