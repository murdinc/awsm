package config

type InstanceClassConfigs map[string]InstanceClassConfig

type InstanceClassConfig struct {
	InstanceType    string
	SecurityGroups  []string
	Subnet          string
	PublicIpAddress bool
	AMI             string
	Keys            []string
}

func DefaultInstanceClasses() InstanceClassConfigs {
	defaultInstances := make(InstanceClassConfigs)

	defaultInstances["base"] = InstanceClassConfig{
		InstanceType:    "t1.micro",
		SecurityGroups:  []string{"dev"},
		Subnet:          "private",
		PublicIpAddress: false,
		AMI:             "base",
		Keys:            []string{"default"},
	}

	defaultInstances["dev"] = InstanceClassConfig{
		InstanceType:    "r3.large",
		SecurityGroups:  []string{"all", "dev"},
		Subnet:          "private",
		PublicIpAddress: false,
		AMI:             "base",
		Keys:            []string{"default"},
	}

	defaultInstances["prod"] = InstanceClassConfig{
		InstanceType:    "r3.large",
		SecurityGroups:  []string{"dev"},
		Subnet:          "private",
		PublicIpAddress: false,
		AMI:             "base",
		Keys:            []string{"default"},
	}

	return defaultInstances
}
