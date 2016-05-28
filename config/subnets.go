package config

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type SubnetClassConfigs map[string]SubnetClassConfig

type SubnetClassConfig struct {
	CIDR string
}

func DefaultSubnetClasses() SubnetClassConfigs {
	defaultSubnets := make(SubnetClassConfigs)

	defaultSubnets["private"] = SubnetClassConfig{
		CIDR: "/24",
	}

	defaultSubnets["public"] = SubnetClassConfig{
		CIDR: "/24",
	}

	return defaultSubnets
}

func (c *SubnetClassConfig) LoadConfig(class string) error {
	data, err := GetClassConfig("subnet", class)
	if err != nil {
		return err
	}

	c.Marshall(data.Attributes)

	return nil
}

func (c *SubnetClassConfig) Marshall(attributes []*simpledb.Attribute) {
	for _, attribute := range attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "CIDR":
			c.CIDR = val

		}
	}
}

func LoadAllSubnetConfigs() (SubnetClassConfigs, error) {
	configType := "subnet"
	data, err := GetAllClassConfigs(configType)
	if err != nil {
		return SubnetClassConfigs{}, err
	}

	configs := make(SubnetClassConfigs)

	for _, item := range data.Items {
		name := strings.Replace(*item.Name, configType+"/", "", -1)
		cfg := new(SubnetClassConfig)
		cfg.Marshall(item.Attributes)
		configs[name] = *cfg
	}

	return configs, nil
}
