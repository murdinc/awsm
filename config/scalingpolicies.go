package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// ScalingPolicyClasses is a map of Scaling Policy Classes
type ScalingPolicyClasses map[string]ScalingPolicyClass

// ScalingPolicyClass is a single Scaling Policy Class
type ScalingPolicyClass struct {
	ScalingAdjustment int      `json:"scalingAdjustment" awsmClass:"Scaling Adjustment"`
	AdjustmentType    string   `json:"adjustmentType" awsmClass:"Adjustment Type"`
	Cooldown          int      `json:"cooldown" awsmClass:"Cooldown"`
	Alarms            []string `json:"alarms" awsmClass:"Alarms"`
}

// DefaultScalingPolicyClasses returns the defauly Scaling Policy Classes
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

// SaveScalingPolicyClass reads unmarshals a byte slice and inserts it into the db
func SaveScalingPolicyClass(className string, data []byte) (class ScalingPolicyClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = InsertClasses("scalingpolicies", ScalingPolicyClasses{className: class})
	return
}

// LoadScalingPolicyClass loads a Scaling Policy Class by its name
func LoadScalingPolicyClass(name string) (ScalingPolicyClass, error) {
	cfgs := make(ScalingPolicyClasses)
	item, err := GetItemByName("scalingpolicies", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllScalingPolicyClasses loads all Scaling Policies Classes
func LoadAllScalingPolicyClasses() (ScalingPolicyClasses, error) {
	cfgs := make(ScalingPolicyClasses)
	items, err := GetItemsByType("scalingpolicies")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into a Scaling Policy Class
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
