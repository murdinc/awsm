package config

import "github.com/aws/aws-sdk-go/service/simpledb"

type VpcClassConfigs map[string]VpcClassConfig

type VpcClassConfig struct {
	CIDR    string
	Tenancy string
}

func DefaultVpcClasses() VpcClassConfigs {
	defaultVpcs := make(VpcClassConfigs)

	defaultVpcs["awsm"] = VpcClassConfig{
		CIDR:    "/16",
		Tenancy: "default",
	}

	return defaultVpcs
}

func (c *VpcClassConfig) LoadConfig(class string) error {
	data, err := GetClassConfig("vpcs", class)
	if err != nil {
		return err
	}

	c.Marshal(data.Attributes)

	return nil
}

func (c *VpcClassConfig) Marshal(attributes []*simpledb.Attribute) {
	for _, attribute := range attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "CIDR":
			c.CIDR = val

		case "Tenancy":
			c.Tenancy = val
		}
	}
}
