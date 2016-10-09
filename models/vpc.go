package models

type Vpc struct {
	Name      string `json:"name" awsmTable:"Name"`
	Class     string `json:"class" awsmTable:"Class"`
	VpcId     string `json:"vpcId" awsmTable:"VPC ID"`
	State     string `json:"state" awsmTable:"State"`
	Default   bool   `json:"default"`
	CIDRBlock string `json:"cidrBlock" awsmTable:"CIDR Block"`
	DHCPOptId string `json:"dhcpOptId" awsmTable:"DHCP Opt Id"`
	Tenancy   string `json:"tenancy" awsmTable:"Tenancy"`
	Region    string `json:"region" awsmTable:"Region"`
}
