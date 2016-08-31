package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type InstanceClassConfigs map[string]InstanceClassConfig

type InstanceClassConfig struct {
	InstanceType     string
	SecurityGroups   []string
	EBSVolumes       []string
	Vpc              string
	Subnet           string
	PublicIpAddress  bool
	AMI              string
	KeyName          string
	EbsOptimized     bool
	Monitoring       bool
	ShutdownBehavior string
	IAMUser          string
	UserData         string
}

func DefaultInstanceClasses() InstanceClassConfigs {
	defaultInstances := make(InstanceClassConfigs)

	defaultInstances["base"] = InstanceClassConfig{
		InstanceType:     "t1.micro",
		SecurityGroups:   []string{"dev"},
		EBSVolumes:       []string{},
		Vpc:              "awsm",
		Subnet:           "private",
		PublicIpAddress:  false,
		AMI:              "base",
		KeyName:          "awsm",
		ShutdownBehavior: "terminate",
	}

	defaultInstances["dev"] = InstanceClassConfig{
		InstanceType:     "r3.large",
		SecurityGroups:   []string{"all", "dev"},
		EBSVolumes:       []string{"git-standard", "mysql-data-standard"}, // TODO
		Vpc:              "awsm",
		Subnet:           "private",
		PublicIpAddress:  false,
		AMI:              "hvm-base",
		KeyName:          "awsm",
		ShutdownBehavior: "terminate",
		UserData:         "#!/bin/bash \n echo wemadeit > ~/didwemakeit",
	}

	defaultInstances["prod"] = InstanceClassConfig{
		InstanceType:     "r3.large",
		SecurityGroups:   []string{"dev"},
		EBSVolumes:       []string{},
		Vpc:              "awsm",
		Subnet:           "private",
		PublicIpAddress:  false,
		AMI:              "hvm-base",
		KeyName:          "awsm",
		ShutdownBehavior: "terminate",
	}

	return defaultInstances
}

func LoadInstanceClass(name string) (InstanceClassConfig, error) {
	cfgs := make(InstanceClassConfigs)
	item, err := GetItemByName("instances", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllInstanceClasses() (InstanceClassConfigs, error) {
	cfgs := make(InstanceClassConfigs)
	items, err := GetItemsByType("instances")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c InstanceClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "instances/", "", -1)
		cfg := new(InstanceClassConfig)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "InstanceType":
				cfg.InstanceType = val

			case "SecurityGroups":
				cfg.SecurityGroups = append(cfg.SecurityGroups, val)

			case "EBSVolumes":
				cfg.EBSVolumes = append(cfg.EBSVolumes, val)

			case "Subnet":
				cfg.Subnet = val

			case "Vpc":
				cfg.Vpc = val

			case "PublicIpAddress":
				cfg.PublicIpAddress, _ = strconv.ParseBool(val)

			case "AMI":
				cfg.AMI = val

			case "KeyName":
				cfg.KeyName = val

			case "EbsOptimized":
				cfg.EbsOptimized, _ = strconv.ParseBool(val)

			case "Monitoring":
				cfg.Monitoring, _ = strconv.ParseBool(val)

			case "ShutdownBehavior":
				cfg.ShutdownBehavior = val

			case "UserData":
				cfg.UserData = val

			case "IAMUser":
				cfg.IAMUser = val

			}
		}
		c[name] = *cfg
	}
}
