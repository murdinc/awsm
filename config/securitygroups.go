package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type SecurityGroupClasses map[string]SecurityGroupClass

type SecurityGroupClass struct {
	Description         string               `json:"description"`
	SecurityGroupGrants []SecurityGroupGrant `json:"securityGroupGrants"`
}

type SecurityGroupGrant struct {
	Note       string   `json:"note"`
	Type       string   `json:"type"` // ingress / egress
	FromPort   int      `json:"fromPort"`
	ToPort     int      `json:"toPort"`
	IpProtocol string   `json:"ipProtocol"`
	CidrIp     []string `json:"cidrIp"`
}

func DefaultSecurityGroupClasses() SecurityGroupClasses {
	defaultSecurityGroups := make(SecurityGroupClasses)

	defaultSecurityGroups["dev"] = SecurityGroupClass{
		Description: "dev server security group",
		SecurityGroupGrants: []SecurityGroupGrant{
			SecurityGroupGrant{
				Note:       "http port 80",
				Type:       "ingress",
				FromPort:   80,
				ToPort:     80,
				IpProtocol: "tcp",
				CidrIp:     []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"},
			},
			SecurityGroupGrant{
				Note:       "http port 443",
				Type:       "ingress",
				FromPort:   443,
				ToPort:     443,
				IpProtocol: "tcp",
				CidrIp:     []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"},
			},
		},
	}

	defaultSecurityGroups["prod"] = SecurityGroupClass{
		Description: "prod server security group",
		SecurityGroupGrants: []SecurityGroupGrant{
			SecurityGroupGrant{
				Note:       "http port 80",
				Type:       "ingress",
				FromPort:   80,
				ToPort:     80,
				IpProtocol: "tcp",
				CidrIp:     []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"},
			},
			SecurityGroupGrant{
				Note:       "http port 443",
				Type:       "ingress",
				FromPort:   443,
				ToPort:     443,
				IpProtocol: "tcp",
				CidrIp:     []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"},
			},
		},
	}

	return defaultSecurityGroups
}

func LoadSecurityGroupClass(name string) (SecurityGroupClass, error) {
	cfgs := make(SecurityGroupClasses)
	item, err := GetItemByName("securitygroups", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllSecurityGroupClasses() (SecurityGroupClasses, error) {
	cfgs := make(SecurityGroupClasses)
	items, err := GetItemsByType("securitygroups")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c SecurityGroupClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "securitygroups/", "", -1)
		cfg := new(SecurityGroupClass)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "Description":
				cfg.Description = val
			}

		}

		// Get the grants
		grants, _ := GetItemsByType("securitygroups/" + name + "/grants")
		cfg.SecurityGroupGrants = make([]SecurityGroupGrant, len(grants))
		for i, grant := range grants {

			for _, attribute := range grant.Attributes {

				val := *attribute.Value

				switch *attribute.Name {

				case "Note":
					cfg.SecurityGroupGrants[i].Note = val

				case "Type":
					cfg.SecurityGroupGrants[i].Type = val

				case "FromPort":
					cfg.SecurityGroupGrants[i].FromPort, _ = strconv.Atoi(val)

				case "ToPort":
					cfg.SecurityGroupGrants[i].ToPort, _ = strconv.Atoi(val)

				case "IpProtocol":
					cfg.SecurityGroupGrants[i].IpProtocol = val

				case "CidrIp":
					cfg.SecurityGroupGrants[i].CidrIp = append(cfg.SecurityGroupGrants[i].CidrIp, val)

				}
			}

		}

		c[name] = *cfg
	}
}
