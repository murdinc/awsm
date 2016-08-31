package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type VolumeClassConfigs map[string]VolumeClassConfig

type VolumeClassConfig struct {
	DeviceName          string
	VolumeSize          int
	DeleteOnTermination bool
	MountPoint          string
	Snapshot            string
	VolumeType          string
	Iops                int
	Encrypted           bool
}

func DefaultVolumeClasses() VolumeClassConfigs {
	defaultVolumes := make(VolumeClassConfigs)

	defaultVolumes["git-standard"] = VolumeClassConfig{
		DeviceName:          "/dev/xvdf",
		VolumeSize:          30,
		DeleteOnTermination: true,
		MountPoint:          "/mnt/git",
		Encrypted:           false,
		Snapshot:            "git",
		VolumeType:          "standard",
	}

	defaultVolumes["mysql-data-standard"] = VolumeClassConfig{
		DeviceName:          "/dev/xvdg",
		VolumeSize:          100,
		DeleteOnTermination: true,
		MountPoint:          "/media/mysql-data",
		Encrypted:           false,
		Snapshot:            "mysql-data",
		VolumeType:          "standard",
	}

	return defaultVolumes
}

func LoadVolumeClass(name string) (VolumeClassConfig, error) {
	cfgs := make(VolumeClassConfigs)
	item, err := GetItemByName("volumes", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllVolumeClasses() (VolumeClassConfigs, error) {
	cfgs := make(VolumeClassConfigs)
	items, err := GetItemsByType("volumes")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c VolumeClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "volumes/", "", -1)
		cfg := new(VolumeClassConfig)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "DeviceName":
				cfg.DeviceName = val

			case "VolumeSize":
				cfg.VolumeSize, _ = strconv.Atoi(val)

			case "DeleteOnTermination":
				cfg.DeleteOnTermination, _ = strconv.ParseBool(val)

			case "MountPoint":
				cfg.MountPoint = val

			case "Snapshot":
				cfg.Snapshot = val

			case "VolumeType":
				cfg.VolumeType = val

			case "Iops":
				cfg.Iops, _ = strconv.Atoi(val)

			case "Encrypted":
				cfg.Encrypted, _ = strconv.ParseBool(val)

			}

			c[name] = *cfg
		}
	}
}
