package models

// Instance represents an Elastic Computer Cloud (EC2) Instance
type Instance struct {
	Name                   string `json:"name" awsmTable:"Name"`
	Class                  string `json:"class" awsmTable:"Class"`
	PrivateIP              string `json:"privateIP" awsmTable:"Private IP"`
	PublicIP               string `json:"publicIP" awsmTable:"Public IP"`
	InstanceID             string `json:"instanceID" awsmTable:"Instance ID"`
	AMIID                  string `json:"amiID"`
	AMIName                string `json:"amiName" awsmTable:"AMI"`
	Root                   string `json:"root" awsmTable:"Root"`
	Size                   string `json:"size" awsmTable:"Size"`
	Virtualization         string `json:"virtualization"`
	State                  string `json:"state" awsmTable:"State"`
	KeyPair                string `json:"keyPair" awsmTable:"KeyPair"`
	AvailabilityZone       string `json:"availabilityZone" awsmTable:"Availability Zone"`
	VPC                    string `json:"vpc" awsmTable:"VPC"`
	VPCID                  string `json:"vpcID"`
	Subnet                 string `json:"subnet" awsmTable:"Subnet"`
	SubnetID               string `json:"subnetID"`
	IAMUser                string `json:"iamUser"`
	IamInstanceProfileArn  string `json:"iamInstanceProfileArn"`
	IamInstanceProfileName string `json:"iamInstanceProfileName" awsmTable:"IAM Instance Profile"`
	ShutdownBehavior       string `json:"shutdownBehavior"`
	EbsOptimized           bool   `json:"ebsOptimized"` // TODO
	Monitoring             bool   `json:"monitoring"`   // TODO
	Region                 string `json:"region"`
}
