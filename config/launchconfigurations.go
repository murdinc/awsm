package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type LaunchConfigurationClassConfigs map[string]LaunchConfigurationClassConfig

type LaunchConfigurationClassConfig struct {
	Version       int
	InstanceClass string
	Retain        int
	Regions       []string
}

func DefaultLaunchConfigurationClasses() LaunchConfigurationClassConfigs {
	defaultLCs := make(LaunchConfigurationClassConfigs)

	defaultLCs["prod"] = LaunchConfigurationClassConfig{
		Version:       0,
		InstanceClass: "prod",
		Retain:        5,
		Regions:       []string{"us-west-2", "us-east-1", "eu-west-1"},
	}

	return defaultLCs
}

func LoadLaunchConfigurationClass(name string) (LaunchConfigurationClassConfig, error) {
	cfgs := make(LaunchConfigurationClassConfigs)
	item, err := GetItemByName("launchconfigurations", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllLaunchConfigurationClasses() (LaunchConfigurationClassConfigs, error) {
	cfgs := make(LaunchConfigurationClassConfigs)
	items, err := GetItemsByType("launchconfigurations")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c LaunchConfigurationClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "launchconfigurations/", "", -1)
		cfg := new(LaunchConfigurationClassConfig)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "Version":
				cfg.Version, _ = strconv.Atoi(val)

			case "InstanceClass":
				cfg.InstanceClass = val

			case "Regions":
				cfg.Regions = append(cfg.Regions, val)

			case "Retain":
				cfg.Retain, _ = strconv.Atoi(val)

			}
		}
		c[name] = *cfg
	}
}

func (c *LaunchConfigurationClassConfig) SetVersion(name string, version int) error {
	c.Version = version

	updateCfgs := make(LaunchConfigurationClassConfigs)
	updateCfgs[name] = *c

	return InsertClassConfigs("launchconfig", updateCfgs)
}

func (c *LaunchConfigurationClassConfig) Increment(name string) error {
	c.Version += 1
	return c.SetVersion(name, c.Version)
}

func (c *LaunchConfigurationClassConfig) Decrement(name string) error {
	c.Version -= 1
	return c.SetVersion(name, c.Version)
}
