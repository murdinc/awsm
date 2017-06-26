package models

import (
	"time"

	"github.com/murdinc/awsm/config"
)

// LoadBalancer represents an EC2 Load Balancer
type LoadBalancer struct {
	Name                    string                         `json:"name" awsmTable:"Name"`
	Class                   string                         `json:"class" awsmTable:"Class"`
	Region                  string                         `json:"region" awsmTable:"Region"`
	AvailabilityZones       []string                       `json:"availabilityZone" awsmTable:"Availability Zones"`
	CreatedTime             time.Time                      `json:"createdTime" awsmTable:"Created"`
	SecurityGroups          []string                       `json:"securityGroups" awsmTable:"Security Groups"`
	Scheme                  string                         `json:"scheme" awsmTable:"Scheme"`
	Vpc                     string                         `json:"vpc" awsmTable:"VPC"`
	VpcID                   string                         `json:"vpcID"`
	Subnets                 []string                       `json:"subnets" awsmTable:"Subnets"`
	SubnetClasses           []string                       `json:"subnetsClasses"`
	SubnetIDs               []string                       `json:"subnetIDs"`
	DNSName                 string                         `json:"dnsName"`
	LoadBalancerListeners   []config.LoadBalancerListener  `json:"loadBalancerListeners"`
	LoadBalancerHealthCheck config.LoadBalancerHealthCheck `json:"loadBalancerHealthCheck"`
	LoadBalancerAttributes  config.LoadBalancerAttributes  `json:"loadBalancerAttributes"`
}
