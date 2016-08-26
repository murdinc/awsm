package config

import "strconv"

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

func (c *LaunchConfigurationClassConfig) LoadConfig(class string) error {

	data, err := GetClassConfig("launchconfig", class)
	if err != nil {
		return err
	}

	for _, attribute := range data.Attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "Version":
			c.Version, _ = strconv.Atoi(val)

		case "InstanceClass":
			c.InstanceClass = val

		case "Regions":
			c.Regions = append(c.Regions, val)

		case "Retain":
			c.Retain, _ = strconv.Atoi(val)

		}
	}

	return nil

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
