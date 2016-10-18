package models

import "time"

type LaunchConfig struct {
	Name           string    `json:"name" awsmTable:"Name"`
	ImageName      string    `json:"imageName" awsmTable:"Image Name"`
	ImageID        string    `json:"imageID" awsmTable:"Image ID"`
	InstanceType   string    `json:"instanceType" awsmTable:"Instance Type"`
	KeyName        string    `json:"keyName" awsmTable:"Key Name"`
	SecurityGroups string    `json:"securityGroups" awsmTable:"Security Groups"`
	CreationTime   time.Time `json:"creationTime"`
	CreatedHuman   string    `json:"createdHuman" awsmTable:"Created"`
	Region         string    `json:"region" awsmTable:"Region"`
	EbsOptimized   bool      `json:"ebsOptimized" awsmTable:"EBS Optimized"`
	SnapshotIDs    []string  `json:"snapshotID" awsmTable:"Snapshot IDs"`
}
