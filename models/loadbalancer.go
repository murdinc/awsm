package models

import "time"

// LoadBalancer represents an EC2 Load Balancer
type LoadBalancer struct {
	Name                string    `json:"name" awsmTable:"Name"`
	DNSName             string    `json:"dnsName"`
	Region              string    `json:"region" awsmTable:"Region"`
	AvailabilityZones   string    `json:"availabilityZone" awsmTable:"Availability Zones"`
	CreatedTime         time.Time `json:"createdTime"`
	CreatedHuman        string    `json:"createdHuman" awsmTable:"Created"`
	SecurityGroups      string    `json:"securityGroups" awsmTable:"Security Groups"`
	Scheme              string    `json:"scheme" awsmTable:"Scheme"`
	VpcID               string    `json:"vpcID" awsmTable:"VPC ID"`
	Vpc                 string    `json:"vpc" awsmTable:"VPC"`
	HealthCheckTarget   string    `json:"healthCheckTarget" awsmTable:"Health Check Target"`
	HealthCheckInterval string    `json:"healthCheckInterval" awsmTable:"Health Check Interval"`
	Subnets             string    `json:"subnets" awsmTable:"Subnets"`
	SubnetIDs           []string  `json:"subnetIDs" awsmTable:"Subnet IDs"`
}
