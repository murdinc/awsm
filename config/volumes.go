package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type VolumeClasses map[string]VolumeClass

type VolumeClass struct {
	DeviceName          string `json:"deviceName" awsmList:"Device Name"`
	VolumeSize          int    `json:"volumeSize" awsmList:"Volume Size"`
	DeleteOnTermination bool   `json:"deleteOnTermination" awsmList:"Delete On Termination"`
	MountPoint          string `json:"mountPoint" awsmList:"Mount Point"`
	Snapshot            string `json:"snapshot" awsmList:"Snapshot"`
	VolumeType          string `json:"volumeType" awsmList:"Volume Type"`
	Iops                int    `json:"iops" awsmList:"IOPS"`
	Encrypted           bool   `json:"encrypted" awsmList:"Encrypted"`
}

func DefaultVolumeClasses() VolumeClasses {
	defaultVolumes := make(VolumeClasses)

	defaultVolumes["git-standard"] = VolumeClass{
		DeviceName:          "/dev/xvdf",
		VolumeSize:          30,
		DeleteOnTermination: true,
		MountPoint:          "/mnt/git",
		Encrypted:           false,
		Snapshot:            "git",
		VolumeType:          "standard",
	}

	defaultVolumes["mysql-data-standard"] = VolumeClass{
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

func LoadVolumeClass(name string) (VolumeClass, error) {
	cfgs := make(VolumeClasses)
	item, err := GetItemByName("volumes", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllVolumeClasses() (VolumeClasses, error) {
	cfgs := make(VolumeClasses)
	items, err := GetItemsByType("volumes")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c VolumeClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "volumes/", "", -1)
		cfg := new(VolumeClass)
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
