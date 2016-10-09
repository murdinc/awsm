package models

type Subnet struct {
	Name             string `json:"name" awsmTable:"Name"`
	Class            string `json:"class" awsmTable:"Class"`
	SubnetId         string `json:"subnetId" awsmTable:"Subnet ID"`
	VpcName          string `json:"vpcName" awsmTable:"VPC Name"`
	VpcId            string `json:"vpcId" awsmTable:"VPC ID"`
	State            string `json:"state" awsmTable:"State"`
	AvailabilityZone string `json:"availabilityZone" awsmTable:"Availability Zone"`
	Default          bool   `json:"default" awsmTable:"Default"`
	CIDRBlock        string `json:"cidrBlock" awsmTable:"CIDR Block"`
	AvailableIPs     int    `json:"availableIps" awsmTable:"Available IPs"`
	MapPublicIp      bool   `json:"mapPublicIp" awsmTable:"Map Public IP"`
	Region           string `json:"region" awsmTable:"Region"`
}
