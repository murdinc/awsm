package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// SecurityGroupClasses is a map of Security Group Classes
type SecurityGroupClasses map[string]SecurityGroupClass

// SecurityGroupClass is a single Security Group Class
type SecurityGroupClass struct {
	Description         string               `json:"description" awsmClass:"Description"`
	SecurityGroupGrants []SecurityGroupGrant `json:"securityGroupGrants"`
}

// SecurityGroupGrant is a Security Group Grant
type SecurityGroupGrant struct {
	ID         string   `json:"id" hash:"ignore" awsm:"ignore"`
	Note       string   `json:"note" hash:"ignore"`
	Type       string   `json:"type"` // ingress / egress
	FromPort   int      `json:"fromPort"`
	ToPort     int      `json:"toPort"`
	IPProtocol string   `json:"ipProtocol"`
	CidrIP     []string `json:"cidrIP" hash:"set"`
}

// DefaultSecurityGroupClasses returns the defauly Security Group Classes
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
				IPProtocol: "tcp",
				CidrIP:     []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"},
			},
			SecurityGroupGrant{
				Note:       "http port 443",
				Type:       "ingress",
				FromPort:   443,
				ToPort:     443,
				IPProtocol: "tcp",
				CidrIP:     []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"},
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
				IPProtocol: "tcp",
				CidrIP:     []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"},
			},
			SecurityGroupGrant{
				Note:       "http port 443",
				Type:       "ingress",
				FromPort:   443,
				ToPort:     443,
				IPProtocol: "tcp",
				CidrIP:     []string{"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16"},
			},
		},
	}

	return defaultSecurityGroups
}

// SaveSecurityGroupClass reads unmarshals a byte slice and inserts it into the db
func SaveSecurityGroupClass(className string, data []byte) (class SecurityGroupClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = InsertClasses("securitygroups", SecurityGroupClasses{className: class})
	return
}

// LoadSecurityGroupClass loads a Security Group Class by its name
func LoadSecurityGroupClass(name string) (SecurityGroupClass, error) {
	cfgs := make(SecurityGroupClasses)
	item, err := GetItemByName("securitygroups", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllSecurityGroupClasses loads all Security Group Classes
func LoadAllSecurityGroupClasses() (SecurityGroupClasses, error) {
	cfgs := make(SecurityGroupClasses)
	items, err := GetItemsByType("securitygroups")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into a Security Group Class
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

			cfg.SecurityGroupGrants[i].ID = strings.Replace(*grant.Name, "securitygroups/"+name+"/grants/", "", -1)

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

				case "IPProtocol":
					cfg.SecurityGroupGrants[i].IPProtocol = val

				case "CidrIP":
					cfg.SecurityGroupGrants[i].CidrIP = append(cfg.SecurityGroupGrants[i].CidrIP, val)

				}
			}

		}

		c[name] = *cfg
	}
}
