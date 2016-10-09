package models

type AutoScaleGroup struct {
	Name                   string `json:"name"`
	Class                  string `json:"class"`
	HealthCheckType        string `json:"healthCheckType"`
	HealthCheckGracePeriod int    `json:"healthCheckGracePeriod"`
	LaunchConfig           string `json:"launchConfig"`
	LoadBalancers          string `json:"loadBalancers"`
	InstanceCount          int    `json:"instanceCount"`
	DesiredCapacity        int    `json:"desiredCapacity"`
	MinSize                int    `json:"minSize"`
	MaxSize                int    `json:"maxSize"`
	DefaultCooldown        int    `json:"defaultCooldown"`
	AvailabilityZones      string `json:"availabilityZones"`
	VpcName                string `json:"vpcName"`
	VpcId                  string `json:"vpcId"`
	SubnetName             string `json:"subnetName"`
	SubnetId               string `json:"subnetId"`
	Region                 string `json:"region"`
	//Instances         string
}
