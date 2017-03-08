package models

import (
	"time"

	"github.com/murdinc/awsm/config"
)

// LoadBalancer represents an EC2 Load Balancer
type LoadBalancer struct {
	Name                  string                        `json:"name" awsmTable:"Name"`
	Class                 string                        `json:"class" awsmTable:"Class"`
	Region                string                        `json:"region" awsmTable:"Region"`
	AvailabilityZones     string                        `json:"availabilityZone" awsmTable:"Availability Zones"`
	CreatedTime           time.Time                     `json:"createdTime" awsmTable:"Created"`
	SecurityGroups        string                        `json:"securityGroups" awsmTable:"Security Groups"`
	Scheme                string                        `json:"scheme" awsmTable:"Scheme"`
	VpcID                 string                        `json:"vpcID" awsmTable:"VPC ID"`
	Vpc                   string                        `json:"vpc" awsmTable:"VPC"`
	HealthCheckInterval   string                        `json:"healthCheckInterval" awsmTable:"Health Check Interval"`
	Subnets               string                        `json:"subnets" awsmTable:"Subnets"`
	SubnetIDs             []string                      `json:"subnetIDs"`
	DNSName               string                        `json:"dnsName"`
	LoadBalancerListeners []config.LoadBalancerListener `json:"loadBalancerListeners"`
	HealthCheckTarget     string                        `json:"healthCheckTarget"`
}
