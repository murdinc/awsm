package config

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/simpledb"
)

type SecurityGroupClassConfigs map[string]SecurityGroupClassConfig

type SecurityGroupClassConfig struct {
	Description string
	SecurityGroupClassPermissions
}

type SecurityGroupClassPermissions struct {
	Ingress []*ec2.IpPermission
	Egress  []*ec2.IpPermission
}

func DefaultSecurityGroupClasses() SecurityGroupClassConfigs {
	defaultSecurityGroups := make(SecurityGroupClassConfigs)

	defaultSecurityGroups["dev"] = SecurityGroupClassConfig{}

	defaultSecurityGroups["prod"] = SecurityGroupClassConfig{}

	return defaultSecurityGroups
}

func LoadSecurityGroupClass(name string) (SecurityGroupClassConfig, error) {
	cfgs := make(SecurityGroupClassConfigs)
	item, err := GetItemByName("securitygroups", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllSecurityGroupClasses() (SecurityGroupClassConfigs, error) {
	cfgs := make(SecurityGroupClassConfigs)
	items, err := GetItemsByType("securitygroups")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c SecurityGroupClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "securitygroups/", "", -1)
		cfg := new(SecurityGroupClassConfig)
		for _, attribute := range item.Attributes {

			fmt.Println(attribute)

			/*
				val := *attribute.Value

				switch *attribute.Name {

				case "Propagate":
					cfg.Propagate, _ = strconv.ParseBool(val)

				case "Retain":
					cfg.Retain, _ = strconv.Atoi(val)

				case "PropagateRegions":
					cfg.PropagateRegions = append(cfg.PropagateRegions, val)

				case "VolumeId":
					cfg.VolumeId = val
				}
			*/
		}
		c[name] = *cfg
	}
}
