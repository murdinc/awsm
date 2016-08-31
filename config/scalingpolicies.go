package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type ScalingPolicyClasses map[string]ScalingPolicyClass

type ScalingPolicyClass struct {
	ScalingAdjustment int      `json:"scalingAdjustment"`
	AdjustmentType    string   `json:"adjustmentType"`
	Cooldown          int      `json:"cooldown"`
	Alarms            []string `json:"alarms"`
}

func DefaultScalingPolicyClasses() ScalingPolicyClasses {
	defaultScalingPolicies := make(ScalingPolicyClasses)

	defaultScalingPolicies["scaleUp"] = ScalingPolicyClass{
		ScalingAdjustment: 1,
		AdjustmentType:    "ChangeInCapacity",
		Cooldown:          300,
		Alarms:            []string{"cpuHigh"},
	}

	defaultScalingPolicies["scaleDown"] = ScalingPolicyClass{
		ScalingAdjustment: -1,
		AdjustmentType:    "ChangeInCapacity",
		Cooldown:          300,
		Alarms:            []string{"cpuHigh"},
	}

	return defaultScalingPolicies
}

func LoadScalingPolicyClass(name string) (ScalingPolicyClass, error) {
	cfgs := make(ScalingPolicyClasses)
	item, err := GetItemByName("scalingpolicies", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllScalingPolicyClasses() (ScalingPolicyClasses, error) {
	cfgs := make(ScalingPolicyClasses)
	items, err := GetItemsByType("scalingpolicies")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c ScalingPolicyClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "scalingpolicies/", "", -1)
		cfg := new(ScalingPolicyClass)
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
