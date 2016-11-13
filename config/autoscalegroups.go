package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// AutoscaleGroupClasses is a map of Autoscale Group Classes
type AutoscaleGroupClasses map[string]AutoscaleGroupClass

// AutoscaleGroupClass is a single Autoscale Group Class
type AutoscaleGroupClass struct {
	LaunchConfigurationClass string   `json:"launchConfigurationClass" awsmClass:"Launch Configuration Class"`
	Propagate                bool     `json:"propagate" awsmClass:"Propagate"`
	Retain                   int      `json:"retain" awsmClass:"Retain"`
	AvailabilityZones        []string `json:"availabilityZones" awsmClass:"Availability Zone"`
	DesiredCapacity          int      `json:"desiredCapacity" awsmClass:"Desired Capacity"`
	MinSize                  int      `json:"minSize" awsmClass:"Min Size"`
	MaxSize                  int      `json:"maxSize" awsmClass:"Max Size"`
	DefaultCooldown          int      `json:"defaultCooldown" awsmClass:"Default Cooldown"`
	SubnetClass              string   `json:"subnetClass" awsmClass:"Subnet Class"`
	HealthCheckType          string   `json:"healthCheckType" awsmClass:"Health Check Type"`
	HealthCheckGracePeriod   int      `json:"healthCheckGracePeriod" awsmClass:"Health Check Grace Period"`
	TerminationPolicies      []string `json:"terminationPolicies" awsmClass:"Termination Policies"`
	ScalingPolicies          []string `json:"scalingPolicies" awsmClass:"Scaling Policies"`
	LoadBalancerNames        []string `json:"loadBalancerNames" awsmClass:"Load Balancer Names"`
	Alarms                   []string `json:"alarms" awsmClass:"Alarms"`
}

// DefaultAutoscaleGroupClasses returns the default Autoscale Group Classes
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
		Alarms:                   []string{"cpuHigh", "cpuLow"},
	}

	return defaultASGs
}

// SaveAlarmClass reads unmarshals a byte slice and inserts it into the db
func SaveAutoscalingGroupClass(className string, data []byte) (class AutoscaleGroupClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = Insert("autoscalegroups", AutoscaleGroupClasses{className: class})
	return
}

// LoadAutoscalingGroupClass loads an Autoscaling Group Class
func LoadAutoscalingGroupClass(name string) (AutoscaleGroupClass, error) {
	cfgs := make(AutoscaleGroupClasses)
	item, err := GetItemByName("autoscalegroups", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllAutoscalingGroupClasses loads all Autoscaling Group Classes
func LoadAllAutoscalingGroupClasses() (AutoscaleGroupClasses, error) {
	cfgs := make(AutoscaleGroupClasses)
	items, err := GetItemsByType("autoscalegroups")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts the items from simpledb into an AutoscaleGroupClass struct
func (c AutoscaleGroupClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "autoscalegroups/", "", -1)
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
