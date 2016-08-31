package config

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type SubnetClasses map[string]SubnetClass

type SubnetClass struct {
	CIDR string `json:"cidr"`
}

func DefaultSubnetClasses() SubnetClasses {
	defaultSubnets := make(SubnetClasses)

	defaultSubnets["private"] = SubnetClass{
		CIDR: "/24",
	}

	defaultSubnets["public"] = SubnetClass{
		CIDR: "/24",
	}

	return defaultSubnets
}

func LoadSubnetClass(name string) (SubnetClass, error) {
	cfgs := make(SubnetClasses)
	item, err := GetItemByName("subnets", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllSubnetClasses() (SubnetClasses, error) {
	cfgs := make(SubnetClasses)
	items, err := GetItemsByType("subnets")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c SubnetClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "subnets/", "", -1)
		cfg := new(SubnetClass)
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
