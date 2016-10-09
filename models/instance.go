package models

type Instance struct {
	Name             string `json:"name" awsmTable:"Name"`
	Class            string `json:"class" awsmTable:"Class""`
	PrivateIp        string `json:"privateIp" awsmTable:"Private IP""`
	PublicIp         string `json:"publicIp" awsmTable:"PublicIP""`
	InstanceId       string `json:"instanceId" awsmTable:"Instance ID""`
	AMIId            string `json:"amiId" awsmTable:"AMI ID""`
	AMIName          string `json:"amiName" awsmTable:"AMI Name""`
	Root             string `json:"root" awsmTable:"Root""`
	Size             string `json:"size" awsmTable:"Size""`
	Virtualization   string `json:"virtualization" awsmTable:"Virtualization""`
	State            string `json:"state" awsmTable:"State""`
	KeyPair          string `json:"keyPair" awsmTable:"KeyPair""`
	AvailabilityZone string `json:"availabilityZone" awsmTable:"Availability Zone""`
	VPC              string `json:"vpc" awsmTable:"VPC""`
	VPCId            string `json:"vpcId" awsmTable:"VPC ID""`
	Subnet           string `json:"subnet" awsmTable:"Subnet""`
	SubnetId         string `json:"subnetId" awsmTable:"Subnet ID""`
	IAMUser          string `json:"iamUser" awsmTable:"IAM User""`
	ShutdownBehavior string `json:"shutdownBehavior" awsmTable:"Shutdown Behavior""`
	EbsOptimized     bool   `json:"ebsOptimized"` // TODO
	Monitoring       bool   `json:"monitoring"`   // TODO
	Region           string `json:"region" awsmTable:"Region""`
}
