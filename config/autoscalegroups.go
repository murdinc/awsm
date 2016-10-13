package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type AutoscaleGroupClasses map[string]AutoscaleGroupClass

type AutoscaleGroupClass struct {
	LaunchConfigurationClass string   `json:"launchConfigurationClass" awsmList:"Launch Configuration Class"`
	Propagate                bool     `json:"propagate" awsmList:"Propagate"`
	Retain                   int      `json:"retain" awsmList:"Retain"`
	AvailabilityZones        []string `json:"availabilityZones" awsmList:"Availability Zone"`
	DesiredCapacity          int      `json:"desiredCapacity" awsmList:"Desired Capacity"`
	MinSize                  int      `json:"minSize" awsmList:"Min Size"`
	MaxSize                  int      `json:"maxSize" awsmList:"Max Size"`
	DefaultCooldown          int      `json:"defaultCooldown" awsmList:"Default Cooldown"`
	SubnetClass              string   `json:"subnetClass" awsmList:"Subnet Class"`
	HealthCheckType          string   `json:"healthCheckType" awsmList:"Health Check Type"`
	HealthCheckGracePeriod   int      `json:"healthCheckGracePeriod" awsmList:"Health Check Grace Period"`
	TerminationPolicies      []string `json:"terminationPolicies" awsmList:"Termination Policies"`
	ScalingPolicies          []string `json:"scalingPolicies" awsmList:"Scaling Policies"`
	LoadBalancerNames        []string `json:"loadBalancerNames" awsmList:"Load Balancer Names"`
	Alarms                   []string `json:"alarms" awsmList:"Alarms"`
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
