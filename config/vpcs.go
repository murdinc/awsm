package config

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

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

func LoadVpcClass(name string) (VpcClassConfig, error) {
	cfgs := make(VpcClassConfigs)
	item, err := GetItemByName("vpcs", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllVpcClasses() (VpcClassConfigs, error) {
	cfgs := make(VpcClassConfigs)
	items, err := GetItemsByType("vpcs")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c VpcClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "vpcs/", "", -1)
		cfg := new(VpcClassConfig)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "CIDR":
				cfg.CIDR = val

			case "Tenancy":
				cfg.Tenancy = val
			}
		}

		c[name] = *cfg
	}
}
