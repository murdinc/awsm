package models

type AutoScaleGroup struct {
	Name                   string   `json:"name" awsmTable:"Name"`
	Class                  string   `json:"class" awsmTable:"Class"`
	HealthCheckType        string   `json:"healthCheckType" awsmTable:"Health Check Type"`
	HealthCheckGracePeriod int      `json:"healthCheckGracePeriod" awsmTable:"Health Check Grace Period"`
	LaunchConfig           string   `json:"launchConfig" awsmTable:"Launch Configuration"`
	LoadBalancers          []string `json:"loadBalancers" awsmTable:"Load Balancers"`
	InstanceCount          int      `json:"instanceCount" awsmTable:"Instance Count"`
	DesiredCapacity        int      `json:"desiredCapacity" awsmTable:"Desired Capacity"`
	MinSize                int      `json:"minSize" awsmTable:"Min Size"`
	MaxSize                int      `json:"maxSize" awsmTable:"Max Size"`
	DefaultCooldown        int      `json:"defaultCooldown" awsmTable:"Default Cooldown"`
	AvailabilityZones      []string `json:"availabilityZones" awsmTable:"Availability Zones"`
	VpcName                string   `json:"vpcName" awsmTable:"VPC Name"`
	VpcID                  string   `json:"vpcID" awsmTable:"VPC ID"`
	SubnetName             string   `json:"subnetName" awsmTable:"Subnet Name"`
	SubnetID               string   `json:"subnetID" awsmTable:"Subnet ID"`
	Region                 string   `json:"region" awsmTable:"Region"`
	//Instances         string
}
