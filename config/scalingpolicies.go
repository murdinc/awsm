package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type ScalingPolicyClassConfigs map[string]ScalingPolicyClassConfig

type ScalingPolicyClassConfig struct {
	ScalingAdjustment int
	AdjustmentType    string
	Cooldown          int
	Alarms            []string
}

func DefaultScalingPolicyClasses() ScalingPolicyClassConfigs {
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

func LoadScalingPolicyClass(name string) (ScalingPolicyClassConfig, error) {
	cfgs := make(ScalingPolicyClassConfigs)
	item, err := GetItemByName("scalingpolicies", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllScalingPolicyClasses() (ScalingPolicyClassConfigs, error) {
	cfgs := make(ScalingPolicyClassConfigs)
	items, err := GetItemsByType("scalingpolicies")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c ScalingPolicyClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "scalingpolicies/", "", -1)
		cfg := new(ScalingPolicyClassConfig)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "ScalingAdjustment":
				cfg.ScalingAdjustment, _ = strconv.Atoi(val)

			case "AdjustmentType":
				cfg.AdjustmentType = val

			case "Cooldown":
				cfg.Cooldown, _ = strconv.Atoi(val)

			case "Alarms":
				cfg.Alarms = append(cfg.Alarms, val)

			}
		}
		c[name] = *cfg
	}
}
