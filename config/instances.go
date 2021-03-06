package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// InstanceClasses is a map if Instance classes
type InstanceClasses map[string]InstanceClass

// InstanceClass is a single Instance class
type InstanceClass struct {
	InstanceType       string   `json:"instanceType" awsmClass:"Instance Type"`
	SecurityGroups     []string `json:"securityGroups" awsmClass:"Security Groups"`
	EBSVolumes         []string `json:"ebsVolumes" awsmClass:"EBS Volumes"`
	Vpc                string   `json:"vpc" awsmClass:"VPC"`
	Subnet             string   `json:"subnet" awsmClass:"Subnet"`
	PublicIPAddress    bool     `json:"publicIpAddress" awsmClass:"Public IP Address"`
	AMI                string   `json:"ami" awsmClass:"AMI"`
	KeyName            string   `json:"keyName" awsmClass:"Key Name"`
	EbsOptimized       bool     `json:"ebsOptimized" awsmClass:"EBS Optimized"`
	Monitoring         bool     `json:"monitoring" awsmClass:"Monitoring"`
	ShutdownBehavior   string   `json:"shutdownBehavior" awsmClass:"Shutdown Behaviour"`
	IAMInstanceProfile string   `json:"iamInstanceProfile" awsmClass:"IAM Instance Profile"`
	UserData           string   `json:"userData"`
}

// DefaultInstanceClasses returns the default Instance classes
func DefaultInstanceClasses() InstanceClasses {
	defaultInstances := make(InstanceClasses)

	defaultInstances["hello-world"] = InstanceClass{
		InstanceType:     "t2.nano",
		SecurityGroups:   []string{"dev"},
		Vpc:              "awsm",
		Subnet:           "public",
		PublicIPAddress:  true,
		KeyName:          "awsm",
		ShutdownBehavior: "stop",
		UserData:         "#cloud-config\npackage_update: false\npackage_upgrade: false\nruncmd:\n  - curl -s http://dl.sudoba.sh/get/awsm | sh\n  - curl -s http://dl.sudoba.sh/get/crusher | sh\n  - su - ubuntu -c \"rm -rf /home/ubuntu/crusher-config\"\n  - su - ubuntu -c \"git clone https://github.com/murdinc/crusher-config.git\"\n  - su - ubuntu -c \"cd crusher-config/ && git pull && crusher lc hello_world\"",
		// No AMI Specified, will prompt user to provide one
	}

	return defaultInstances
}

// SaveInstanceClass reads unmarshals a byte slice and inserts it into the db
func SaveInstanceClass(className string, data []byte) (class InstanceClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = Insert("instances", InstanceClasses{className: class})
	return
}

// LoadInstanceClass returns an Instance class by its name
func LoadInstanceClass(name string) (InstanceClass, error) {
	cfgs := make(InstanceClasses)
	item, err := GetItemByName("instances", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllInstanceClasses returns all Instance classes
func LoadAllInstanceClasses() (InstanceClasses, error) {
	cfgs := make(InstanceClasses)
	items, err := GetItemsByType("instances")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into an Instance class
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

			case "PublicIPAddress":
				cfg.PublicIPAddress, _ = strconv.ParseBool(val)

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

			case "IAMInstanceProfile":
				cfg.IAMInstanceProfile = val

			}
		}
		c[name] = *cfg
	}
}
