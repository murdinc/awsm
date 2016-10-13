package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type InstanceClasses map[string]InstanceClass

type InstanceClass struct {
	InstanceType     string   `json:"instanceType" awsmList:"Instance Type"`
	SecurityGroups   []string `json:"securityGroups" awsmList:"Security Groups"`
	EBSVolumes       []string `json:"ebsVolumes" awsmList:"EBS Volumes"`
	Vpc              string   `json:"vpc" awsmList:"VPC"`
	Subnet           string   `json:"subnet" awsmList:"Subnet"`
	PublicIpAddress  bool     `json:"publicIpAddress" awsmList:"Public IP Address"`
	AMI              string   `json:"ami" awsmList:"AMI"`
	KeyName          string   `json:"keyName" awsmList:"Key Name"`
	EbsOptimized     bool     `json:"ebsOptimized" awsmList:"EBS Optimized"`
	Monitoring       bool     `json:"monitoring" awsmList:"Monitoring"`
	ShutdownBehavior string   `json:"shutdownBehavior" awsmList:"Shutdown Behaviour"`
	IAMUser          string   `json:"iamUser" awsmList:"IAM User"`
	UserData         string   `json:"userData"`
}

func DefaultInstanceClasses() InstanceClasses {
	defaultInstances := make(InstanceClasses)

	defaultInstances["base"] = InstanceClass{
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

	defaultInstances["dev"] = InstanceClass{
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

	defaultInstances["prod"] = InstanceClass{
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

func LoadInstanceClass(name string) (InstanceClass, error) {
	cfgs := make(InstanceClasses)
	item, err := GetItemByName("instances", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllInstanceClasses() (InstanceClasses, error) {
	cfgs := make(InstanceClasses)
	items, err := GetItemsByType("instances")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c InstanceClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "instances/", "", -1)
		cfg := new(InstanceClass)
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
