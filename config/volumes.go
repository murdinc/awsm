package config

type VolumeClassConfigs map[string]VolumeClassConfig

type VolumeClassConfig struct {
	Retain              int
	DeviceName          string
	Propagate           bool
	VolumeSize          int
	DeleteOnTermination bool
	MountPoint          string
}

func DefaultVolumeClasses() VolumeClassConfigs {
	defaultVolumes := make(VolumeClassConfigs)

	defaultVolumes["git"] = VolumeClassConfig{
		Retain:              5,
		DeviceName:          "/dev/xvdf",
		Propagate:           true,
		VolumeSize:          30,
		DeleteOnTermination: true,
		MountPoint:          "/mnt/git",
	}

	return defaultVolumes
}

func (c *VolumeClassConfig) LoadConfig(class string) error {
	/*
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

			case "Subnet":
				c.Subnet = val

			case "PublicIpAddress":
				c.PublicIpAddress, _ = strconv.ParseBool(val)

			case "AMI":
				c.AMI = val

			case "Keys":
				c.Keys = append(c.SecurityGroups, val)

			}
		}
	*/
	return nil

}
