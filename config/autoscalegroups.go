package config

type AutoScaleGroupClassConfigs map[string]AutoScaleGroupClassConfig

type AutoScaleGroupClassConfig struct {
	Type                   string
	InstanceClass          string
	Propagate              bool
	Retain                 int
	AvailabilityZones      []string
	DesiredCapacity        int
	MinSize                int
	MaxSize                int
	DefaultCooldown        int
	Subnet                 string
	HealthCheckType        string
	HealthCheckGracePeriod int
	TerminationPolicies    []string //?
	ScalingPolicies        string
	LoadBalancerNames      []string
	Alarms                 []string
}

func DefaultAutoScaleGroupClasses() AutoScaleGroupClassConfigs {
	defaultASGs := make(AutoScaleGroupClassConfigs)

	defaultASGs["prod"] = AutoScaleGroupClassConfig{
		Type:                   "version",
		InstanceClass:          "prod",
		Propagate:              true,
		Retain:                 5,
		AvailabilityZones:      []string{"us-west-2a", "us-east-1a"},
		DesiredCapacity:        2,
		MinSize:                1,
		MaxSize:                4,
		DefaultCooldown:        60,
		Subnet:                 "private",
		HealthCheckType:        "ELB",
		HealthCheckGracePeriod: 360,
		TerminationPolicies:    []string{"NewestInstance"},
		ScalingPolicies:        "default",
		LoadBalancerNames:      []string{"prod"},
		Alarms:                 []string{"CPUAlarmHigh", "CPUAlarmLow"},
	}

	return defaultASGs
}

func (c *AutoScaleGroupClassConfig) LoadConfig(class string) error {
	/*
		data, err := GetClassConfig("ec2", class)
		if err != nil {
			return err
		}

		for _, attribute := range data.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "InstanceType":
				c.InstanceType = val

			case "SecurityGroups":
				c.SecurityGroups = append(c.SecurityGroups, val)

			case "Subnet":
				c.Subnet = val

			case "PublicIpAddress":
				c.PublicIpAddress, _ = strconv.ParseBool(val)

			case "AMI":
				c.AMI = val

			case "Keys":
				c.Keys = append(c.SecurityGroups, val)

			}
		}
	*/
	return nil

}
