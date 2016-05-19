package config

type ScalingPolicyConfigs map[string]ScalingPolicyConfig

type ScalingPolicyConfig struct {
	ScalingAdjustment int
	AdjustmentType    string
	Cooldown          int
	Alarms            []string
}

func DefaultScalingPolicies() ScalingPolicyConfigs {
	defaultScalingPolicies := make(ScalingPolicyConfigs)

	defaultScalingPolicies["scaleUp"] = ScalingPolicyConfig{
		ScalingAdjustment: 1,
		AdjustmentType:    "ChangeInCapacity",
		Cooldown:          300,
		Alarms:            []string{"cpuHigh"},
	}

	defaultScalingPolicies["scaleDown"] = ScalingPolicyConfig{
		ScalingAdjustment: -1,
		AdjustmentType:    "ChangeInCapacity",
		Cooldown:          300,
		Alarms:            []string{"cpuHigh"},
	}

	return defaultScalingPolicies
}

func (c *ScalingPolicyConfig) LoadConfig(class string) error {
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
