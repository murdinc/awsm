package models

import "time"

type LoadBalancer struct {
	Name                string    `json:"name" awsmTable:"Name"`
	DNSName             string    `json:"dnsName" awsmTable:"DNS Name"`
	Region              string    `json:"region" awsmTable:"Region"`
	AvailabilityZones   string    `json:"availabilityZone" awsmTable:"Availability Zones"`
	CreatedTime         time.Time `json:"createdTime"`
	CreatedHuman        string    `json:"createdHuman" awsmTable:"Created"`
	SecurityGroups      string    `json:"securityGroups" awsmTable:"Security Groups"`
	Scheme              string    `json:"scheme" awsmTable:"Scheme"`
	VpcId               string    `json:"vpcId" awsmTable:"VPC ID"`
	Vpc                 string    `json:"vpc" awsmTable:"VPC"`
	HealthCheckTarget   string    `json:"healthCheckTarget" awsmTable:"Health Check Target"`
	HealthCheckInterval string    `json:"healthCheckInterval" awsmTable:"Health Check Interval"`
	Subnets             string    `json:"subnets" awsmTable:"Subnets"`
	SubnetIds           []string  `json:"subnetIds" awsmTable:"Subnet IDs"`
}
