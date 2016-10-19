package models

// Subnet represents a VPC Subnet
type Subnet struct {
	Name             string `json:"name" awsmTable:"Name"`
	Class            string `json:"class" awsmTable:"Class"`
	SubnetID         string `json:"subnetID" awsmTable:"Subnet ID"`
	VpcName          string `json:"vpcName" awsmTable:"VPC Name"`
	VpcID            string `json:"vpcID" awsmTable:"VPC ID"`
	State            string `json:"state" awsmTable:"State"`
	AvailabilityZone string `json:"availabilityZone" awsmTable:"Availability Zone"`
	Default          bool   `json:"default" awsmTable:"Default"`
	CIDRBlock        string `json:"cidrBlock" awsmTable:"CIDR Block"`
	AvailableIPs     int    `json:"availableIPs" awsmTable:"Available IPs"`
	MapPublicIP      bool   `json:"mapPublicIP" awsmTable:"Map Public IP"`
	Region           string `json:"region" awsmTable:"Region"`
}
