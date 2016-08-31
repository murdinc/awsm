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

func LoadSubnetClass(name string) (SubnetClassConfig, error) {
	cfgs := make(SubnetClassConfigs)
	item, err := GetItemByName("subnets", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllSubnetClasses() (SubnetClassConfigs, error) {
	cfgs := make(SubnetClassConfigs)
	items, err := GetItemsByType("subnets")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c SubnetClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "subnets/", "", -1)
		cfg := new(SubnetClassConfig)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "CIDR":
				cfg.CIDR = val

			}
		}
		c[name] = *cfg
	}
}
