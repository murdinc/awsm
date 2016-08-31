package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type AutoscaleGroupClassConfigs map[string]AutoscaleGroupClassConfig

type AutoscaleGroupClassConfig struct {
	LaunchConfigurationClass string
	Propagate                bool
	Retain                   int
	AvailabilityZones        []string
	DesiredCapacity          int
	MinSize                  int
	MaxSize                  int
	DefaultCooldown          int
	SubnetClass              string
	HealthCheckType          string
	HealthCheckGracePeriod   int
	TerminationPolicies      []string // ?
	ScalingPolicies          []string // ?
	LoadBalancerNames        []string
	Alarms                   []string
}

func DefaultAutoscaleGroupClasses() AutoscaleGroupClassConfigs {
	defaultASGs := make(AutoscaleGroupClassConfigs)

	defaultASGs["prod"] = AutoscaleGroupClassConfig{
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

func LoadAutoscalingGroupClass(name string) (AutoscaleGroupClassConfig, error) {
	cfgs := make(AutoscaleGroupClassConfigs)
	item, err := GetItemByName("autoscalinggroups", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllAutoscalingGroupClasses() (AutoscaleGroupClassConfigs, error) {
	cfgs := make(AutoscaleGroupClassConfigs)
	items, err := GetItemsByType("autoscalinggroups")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c AutoscaleGroupClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "autoscalinggroups/", "", -1)
		cfg := new(AutoscaleGroupClassConfig)
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
