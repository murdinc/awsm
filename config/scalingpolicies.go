package config

import (
	"strconv"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type ScalingPolicyClassConfigs map[string]ScalingPolicyClassConfig

type ScalingPolicyClassConfig struct {
	ScalingAdjustment int
	AdjustmentType    string
	Cooldown          int
	Alarms            []string
}

func DefaultScalingPolicies() ScalingPolicyClassConfigs {
	defaultScalingPolicies := make(ScalingPolicyClassConfigs)

	defaultScalingPolicies["scaleUp"] = ScalingPolicyClassConfig{
		ScalingAdjustment: 1,
		AdjustmentType:    "ChangeInCapacity",
		Cooldown:          300,
		Alarms:            []string{"cpuHigh"},
	}

	defaultScalingPolicies["scaleDown"] = ScalingPolicyClassConfig{
		ScalingAdjustment: -1,
		AdjustmentType:    "ChangeInCapacity",
		Cooldown:          300,
		Alarms:            []string{"cpuHigh"},
	}

	return defaultScalingPolicies
}

func (c *ScalingPolicyClassConfig) LoadConfig(class string) error {

	data, err := GetClassConfig("scalingpolicies", class)
	if err != nil {
		return err
	}

	c.Marshal(data.Attributes)

	return nil

}

func (c *ScalingPolicyClassConfig) Marshal(attributes []*simpledb.Attribute) {
	for _, attribute := range attributes {

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
}
