package config

import (
	"strconv"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type AutoScaleGroupClassConfigs map[string]AutoScaleGroupClassConfig

type AutoScaleGroupClassConfig struct {
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

func DefaultAutoScaleGroupClasses() AutoScaleGroupClassConfigs {
	defaultASGs := make(AutoScaleGroupClassConfigs)

	defaultASGs["prod"] = AutoScaleGroupClassConfig{
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

func (c *AutoScaleGroupClassConfig) LoadConfig(class string) error {

	data, err := GetClassConfig("autoscalinggroups", class)
	if err != nil {
		return err
	}

	c.Marshal(data.Attributes)

	return nil
}

func (c *AutoScaleGroupClassConfig) Marshal(attributes []*simpledb.Attribute) {
	for _, attribute := range attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "LaunchConfigurationClass":
			c.LaunchConfigurationClass = val

		case "Propagate":
			c.Propagate, _ = strconv.ParseBool(val)

		case "Retain":
			c.Retain, _ = strconv.Atoi(val)

		case "AvailabilityZones":
			c.AvailabilityZones = append(c.AvailabilityZones, val)

		case "DesiredCapacity":
			c.DesiredCapacity, _ = strconv.Atoi(val)

		case "MinSize":
			c.MinSize, _ = strconv.Atoi(val)

		case "MaxSize":
			c.MaxSize, _ = strconv.Atoi(val)

		case "DefaultCooldown":
			c.DefaultCooldown, _ = strconv.Atoi(val)

		case "SubnetClass":
			c.SubnetClass = val

		case "HealthCheckType":
			c.HealthCheckType = val

		case "HealthCheckGracePeriod":
			c.HealthCheckGracePeriod, _ = strconv.Atoi(val)

		case "TerminationPolicies":
			c.TerminationPolicies = append(c.TerminationPolicies, val)

		case "ScalingPolicies":
			c.ScalingPolicies = append(c.ScalingPolicies, val)

		case "LoadBalancerNames":
			c.LoadBalancerNames = append(c.LoadBalancerNames, val)

		case "Alarms":
			c.Alarms = append(c.Alarms, val)

		}
	}
}
