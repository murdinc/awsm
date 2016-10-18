package models

import "github.com/murdinc/awsm/config"

type SecurityGroup struct {
	Name                string                      `json:"name" awsmTable:"Name"`
	Class               string                      `json:"class" awsmTable:"Class"`
	GroupID             string                      `json:"groupID" awsmTable:"Group ID"`
	Description         string                      `json:"description" awsmTable:"Description"`
	Vpc                 string                      `json:"vpc" awsmTable:"VPC"`
	VpcID               string                      `json:"vpcID" awsmTable:"VPC ID"`
	Region              string                      `json:"region" awsmTable:"Region"`
	SecurityGroupGrants []config.SecurityGroupGrant `json:"securityGroupGrants"`
}
