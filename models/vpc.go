package models

type Vpc struct {
	Name      string `json:"name" awsmTable:"Name"`
	Class     string `json:"class" awsmTable:"Class"`
	VpcID     string `json:"vpcID" awsmTable:"VPC ID"`
	State     string `json:"state" awsmTable:"State"`
	Default   bool   `json:"default"`
	CIDRBlock string `json:"cidrBlock" awsmTable:"CIDR Block"`
	DHCPOptID string `json:"dhcpOptID" awsmTable:"DHCP Opt ID"`
	Tenancy   string `json:"tenancy" awsmTable:"Tenancy"`
	Region    string `json:"region" awsmTable:"Region"`
}
