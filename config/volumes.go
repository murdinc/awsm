package config

import "strconv"

type VolumeClassConfigs map[string]VolumeClassConfig

type VolumeClassConfig struct {
	DeviceName          string
	VolumeSize          int
	DeleteOnTermination bool
	MountPoint          string
	//Encrypted           bool
	Snapshot   string
	VolumeType string
	Iops       int
}

func DefaultVolumeClasses() VolumeClassConfigs {
	defaultVolumes := make(VolumeClassConfigs)

	defaultVolumes["git-standard"] = VolumeClassConfig{
		DeviceName:          "/dev/xvdf",
		VolumeSize:          30,
		DeleteOnTermination: true,
		MountPoint:          "/mnt/git",
		//Encrypted:           false,
		Snapshot:   "git",
		VolumeType: "standard",
	}

	defaultVolumes["mysql-data-standard"] = VolumeClassConfig{
		DeviceName:          "/dev/xvdg",
		VolumeSize:          100,
		DeleteOnTermination: true,
		MountPoint:          "/media/mysql-data",
		//Encrypted:           false,
		Snapshot:   "mysql-data",
		VolumeType: "standard",
	}

	return defaultVolumes
}

func (c *VolumeClassConfig) LoadConfig(class string) error {

	data, err := GetClassConfig("ebs-volume", class)
	if err != nil {
		return err
	}

	for _, attribute := range data.Attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "DeviceName":
			c.DeviceName = val

		case "VolumeSize":
			c.VolumeSize, _ = strconv.Atoi(val)

		case "DeleteOnTermination":
			c.DeleteOnTermination, _ = strconv.ParseBool(val)

		case "MountPoint":
			c.MountPoint = val

		case "Snapshot":
			c.Snapshot = val

		case "VolumeType":
			c.VolumeType = val

		case "Iops":
			c.Iops, _ = strconv.Atoi(val)

			/*
				case "Encrypted":
					c.Encrypted, _ = strconv.ParseBool(val)
			*/

		}
	}

	return nil

}
