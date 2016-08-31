package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type AutoscaleGroupClasses map[string]AutoscaleGroupClass

type AutoscaleGroupClass struct {
	LaunchConfigurationClass string   `json:"launchConfigurationClass"`
	Propagate                bool     `json:"propagate"`
	Retain                   int      `json:"retain"`
	AvailabilityZones        []string `json:"availabilityZones"`
	DesiredCapacity          int      `json:"desiredCapacity"`
	MinSize                  int      `json:"minSize"`
	MaxSize                  int      `json:"maxSize"`
	DefaultCooldown          int      `json:"defaultCooldown"`
	SubnetClass              string   `json:"subnetClass"`
	HealthCheckType          string   `json:"healthCheckType"`
	HealthCheckGracePeriod   int      `json:"healthCheckGracePeriod"`
	TerminationPolicies      []string `json:"terminationPolicies"`
	ScalingPolicies          []string `json:"scalingPolicies"`
	LoadBalancerNames        []string `json:"loadBalancerNames"`
	Alarms                   []string `json:"alarms"`
}

func DefaultAutoscaleGroupClasses() AutoscaleGroupClasses {
	defaultASGs := make(AutoscaleGroupClasses)

	defaultASGs["prod"] = AutoscaleGroupClass{
		LaunchConfigurationClass: "prod",
		Propagate:                true,
		Retain:                   5,
		AvailabilityZones:        []string{"us-west-2a", "us-east-1a"},
		DesiredCapacity:          2,
		MinSize:                  1,
		MaxSize:                  4,
		DefaultCooldown:          60,
		SubnetClass:              "private",
		HealthCheckType:          "ELB",
		HealthCheckGracePeriod:   360,
		TerminationPolicies:      []string{"OldestInstance"},
		ScalingPolicies:          []string{"scaleUp", "scaleDown"},
		LoadBalancerNames:        []string{"prod"},
		Alarms:                   []string{"CPUAlarmHigh", "CPUAlarmLow"},
	}

	return defaultASGs
}

func LoadAutoscalingGroupClass(name string) (AutoscaleGroupClass, error) {
	cfgs := make(AutoscaleGroupClasses)
	item, err := GetItemByName("autoscalinggroups", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllAutoscalingGroupClasses() (AutoscaleGroupClasses, error) {
	cfgs := make(AutoscaleGroupClasses)
	items, err := GetItemsByType("autoscalinggroups")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c AutoscaleGroupClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "autoscalinggroups/", "", -1)
		cfg := new(AutoscaleGroupClass)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "LaunchConfigurationClass":
				cfg.LaunchConfigurationClass = val

			case "Propagate":
				cfg.Propagate, _ = strconv.ParseBool(val)

			case "Retain":
				cfg.Retain, _ = strconv.Atoi(val)

			case "AvailabilityZones":
				cfg.AvailabilityZones = append(cfg.AvailabilityZones, val)

			case "DesiredCapacity":
				cfg.DesiredCapacity, _ = strconv.Atoi(val)

			case "MinSize":
				cfg.MinSize, _ = strconv.Atoi(val)

			case "MaxSize":
				cfg.MaxSize, _ = strconv.Atoi(val)

			case "DefaultCooldown":
				cfg.DefaultCooldown, _ = strconv.Atoi(val)

			case "SubnetClass":
				cfg.SubnetClass = val

			case "HealthCheckType":
				cfg.HealthCheckType = val

			case "HealthCheckGracePeriod":
				cfg.HealthCheckGracePeriod, _ = strconv.Atoi(val)

			case "TerminationPolicies":
				cfg.TerminationPolicies = append(cfg.TerminationPolicies, val)

			case "ScalingPolicies":
				cfg.ScalingPolicies = append(cfg.ScalingPolicies, val)

			case "LoadBalancerNames":
				cfg.LoadBalancerNames = append(cfg.LoadBalancerNames, val)

			case "Alarms":
				cfg.Alarms = append(cfg.Alarms, val)

			}
		}
		c[name] = *cfg
	}
}
