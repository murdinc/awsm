package config

import "github.com/aws/aws-sdk-go/service/simpledb"

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
	data, err := GetClassConfig("subnets", class)
	if err != nil {
		return err
	}

	c.Marshal(data.Attributes)

	return nil
}

func (c *SubnetClassConfig) Marshal(attributes []*simpledb.Attribute) {
	for _, attribute := range attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "CIDR":
			c.CIDR = val

		}
	}
}
