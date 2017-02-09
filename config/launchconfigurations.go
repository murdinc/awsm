package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// LaunchConfigurationClasses is a map of Launch Configuration Classes
type LaunchConfigurationClasses map[string]LaunchConfigurationClass

// LaunchConfigurationClass is a single Launch Configuration Class
type LaunchConfigurationClass struct {
	Version       int      `json:"version" awsmClass:"Version"`
	InstanceClass string   `json:"instanceClass" awsmClass:"Instance Class"`
	Retain        int      `json:"retain" awsmClass:"Retain"`
	Rotate        bool     `json:"rotate" awsmClass:"Rotate"`
	Regions       []string `json:"regions" awsmClass:"Regions"`
}

// DefaultLaunchConfigurationClasses returns the default Launch Configuration Classes
func DefaultLaunchConfigurationClasses() LaunchConfigurationClasses {
	defaultLCs := make(LaunchConfigurationClasses)

	defaultLCs["prod"] = LaunchConfigurationClass{
		Version:       0,
		InstanceClass: "prod",
		Retain:        5,
		Rotate:        true,
		Regions:       []string{"us-west-2", "us-east-1", "eu-west-1"},
	}

	return defaultLCs
}

// SaveLaunchConfigurationClass reads unmarshals a byte slice and inserts it into the db
func SaveLaunchConfigurationClass(className string, data []byte) (class LaunchConfigurationClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = Insert("launchconfigurations", LaunchConfigurationClasses{className: class})
	return
}

// LoadLaunchConfigurationClass returns a Launch Configuration Class by its name
func LoadLaunchConfigurationClass(name string) (LaunchConfigurationClass, error) {
	cfgs := make(LaunchConfigurationClasses)
	item, err := GetItemByName("launchconfigurations", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllLaunchConfigurationClasses returns all Launch Configuration Classes
func LoadAllLaunchConfigurationClasses() (LaunchConfigurationClasses, error) {
	cfgs := make(LaunchConfigurationClasses)
	items, err := GetItemsByType("launchconfigurations")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into a class config
func (c LaunchConfigurationClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "launchconfigurations/", "", -1)
		cfg := new(LaunchConfigurationClass)
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

			case "Rotate":
				cfg.Rotate, _ = strconv.ParseBool(val)

			}
		}
		c[name] = *cfg
	}
}

// SetVersion updates the version of a Launch Configuration
func (c *LaunchConfigurationClass) SetVersion(name string, version int) error {
	c.Version = version

	updateCfgs := make(LaunchConfigurationClasses)
	updateCfgs[name] = *c

	return Insert("launchconfigurations", updateCfgs)
}

// Increment increments the version of a Launch Configuration
func (c *LaunchConfigurationClass) Increment(name string) error {
	c.Version++
	return c.SetVersion(name, c.Version)
}

// Decrement decrements the version of a Launch Configuration
func (c *LaunchConfigurationClass) Decrement(name string) error {
	c.Version--
	return c.SetVersion(name, c.Version)
}
