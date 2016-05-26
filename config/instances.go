package config

import "strconv"

type InstanceClassConfigs map[string]InstanceClassConfig

type InstanceClassConfig struct {
	InstanceType     string
	SecurityGroups   []string
	EBSVolumes       []string
	Subnet           string
	PublicIpAddress  bool
	AMI              string
	Keys             []string
	EbsOptimized     bool
	Monitoring       bool
	ShutdownBehavior string
	IAMUser          string
}

func DefaultInstanceClasses() InstanceClassConfigs {
	defaultInstances := make(InstanceClassConfigs)

	defaultInstances["base"] = InstanceClassConfig{
		InstanceType:    "t1.micro",
		SecurityGroups:  []string{"dev"},
		EBSVolumes:      []string{},
		Subnet:          "private",
		PublicIpAddress: false,
		AMI:             "base",
		Keys:            []string{"default"},
	}

	defaultInstances["dev"] = InstanceClassConfig{
		InstanceType:    "r3.large",
		SecurityGroups:  []string{"all", "dev"},
		EBSVolumes:      []string{"git", "mysql-data"}, // TODO
		Subnet:          "private",
		PublicIpAddress: false,
		AMI:             "base",
		Keys:            []string{"default"},
	}

	defaultInstances["prod"] = InstanceClassConfig{
		InstanceType:    "r3.large",
		SecurityGroups:  []string{"dev"},
		EBSVolumes:      []string{},
		Subnet:          "private",
		PublicIpAddress: false,
		AMI:             "base",
		Keys:            []string{"default"},
	}

	return defaultInstances
}

func (c *InstanceClassConfig) LoadConfig(class string) error {

	data, err := GetClassConfig("ec2", class)
	if err != nil {
		return err
	}

	for _, attribute := range data.Attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "InstanceType":
			c.InstanceType = val

		case "SecurityGroups":
			c.SecurityGroups = append(c.SecurityGroups, val)

		case "EBSVolumes":
			c.EBSVolumes = append(c.EBSVolumes, val)

		case "Subnet":
			c.Subnet = val

		case "PublicIpAddress":
			c.PublicIpAddress, _ = strconv.ParseBool(val)

		case "AMI":
			c.AMI = val

		case "Keys":
			c.Keys = append(c.Keys, val)

		case "EbsOptimized":
			c.EbsOptimized, _ = strconv.ParseBool(val)

		case "Monitoring":
			c.Monitoring, _ = strconv.ParseBool(val)

		case "ShutdownBehavior":
			c.ShutdownBehavior = val

		case "IAMUser":
			c.IAMUser = val

		}
	}

	return nil
}
