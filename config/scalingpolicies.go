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
