package config

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type VpcClasses map[string]VpcClass

type VpcClass struct {
	CIDR    string `json:"cidr" awsmList:"CIDR"`
	Tenancy string `json:"tenancy" awsmList:"Tenancy"`
}

func DefaultVpcClasses() VpcClasses {
	defaultVpcs := make(VpcClasses)

	defaultVpcs["awsm"] = VpcClass{
		CIDR:    "/16",
		Tenancy: "default",
	}

	return defaultVpcs
}

func LoadVpcClass(name string) (VpcClass, error) {
	cfgs := make(VpcClasses)
	item, err := GetItemByName("vpcs", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllVpcClasses() (VpcClasses, error) {
	cfgs := make(VpcClasses)
	items, err := GetItemsByType("vpcs")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c VpcClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "vpcs/", "", -1)
		cfg := new(VpcClass)
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
