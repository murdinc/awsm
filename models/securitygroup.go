package models

import "github.com/murdinc/awsm/config"

type SecurityGroup struct {
	Name                string                      `json:"name" awsmTable:"Name"`
	Class               string                      `json:"class" awsmTable:"Class"`
	GroupId             string                      `json:"groupId" awsmTable:"Group ID"`
	Description         string                      `json:"description" awsmTable:"Description"`
	Vpc                 string                      `json:"vpc" awsmTable:"VPC"`
	VpcId               string                      `json:"vpcId" awsmTable:"VPC ID"`
	Region              string                      `json:"region" awsmTable:"Region"`
	SecurityGroupGrants []config.SecurityGroupGrant `json:"securityGroupGrants"`
}
