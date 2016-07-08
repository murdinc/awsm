package config

import "strconv"

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

	data, err := GetClassConfig("scalingpolicy", class)
	if err != nil {
		return err
	}

	for _, attribute := range data.Attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "ScalingAdjustment":
			c.ScalingAdjustment, _ = strconv.Atoi(val)

		case "AdjustmentType":
			c.AdjustmentType = val

		case "Cooldown":
			c.Cooldown, _ = strconv.Atoi(val)

		case "Alarms":
			c.Alarms = append(c.Alarms, val)

		}
	}

	return nil

}
