package models

// Vpc represents a Virtual Private Cloud
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

// InternetGateway represents an Internet Gateway
type InternetGateway struct {
	Name              string `json:"name" awsmTable:"Name"`
	State             string `json:"state" awsmTable:"State"`
	Attachment        string `json:"attachment" awsmTable:"Attachment"`
	InternetGatewayID string `json:"InternetGatewayID" awsmTable:"Internet Gateway ID"`
	Region            string `json:"region" awsmTable:"Region"`
}

// RouteTable represents a Route Table
type RouteTable struct {
	Name         string                  `json:"name" awsmTable:"Name"`
	RouteTableID string                  `json:"routeTableID" awsmTable:"Route Table ID"`
	Associations []RouteTableAssociation `json:"associations" awsmTable:"Associations"`
	VpcID        string                  `json:"vpcID" awsmTable:"VPC ID"`
	Region       string                  `json:"region" awsmTable:"Region"`
}

// RouteTableAssociation represents a single Route Table Association
type RouteTableAssociation struct {
	Main          bool   `json:"main" awsmTable:"Main"`
	AssociationID string `json:"associationID" awsmTable:"Association ID"`
	SubnetID      string `json:"subnetID" awsmTable:"Subnet ID"`
}
