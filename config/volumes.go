package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// VolumeClasses is a map of Volume Classes
type VolumeClasses map[string]VolumeClass

// VolumeClass is a single Volume Class
type VolumeClass struct {
	DeviceName          string `json:"deviceName" awsmClass:"Device Name"`
	VolumeSize          int    `json:"volumeSize" awsmClass:"Volume Size"`
	DeleteOnTermination bool   `json:"deleteOnTermination" awsmClass:"Delete On Termination"`
	MountPoint          string `json:"mountPoint" awsmClass:"Mount Point"`
	Snapshot            string `json:"snapshot" awsmClass:"Snapshot"`
	VolumeType          string `json:"volumeType" awsmClass:"Volume Type"`
	Iops                int    `json:"iops" awsmClass:"IOPS"`
	Encrypted           bool   `json:"encrypted" awsmClass:"Encrypted"`
}

// DefaultVolumeClasses returns the default Volume Classes
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

// SaveVolumeClass reads unmarshals a byte slice and inserts it into the db
func SaveVolumeClass(className string, data []byte) (class VolumeClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	if class.VolumeType != "io1" {
		class.Iops = 0
	}

	err = Insert("volumes", VolumeClasses{className: class})
	return
}

// LoadVolumeClass loads a Volume Class by its name
func LoadVolumeClass(name string) (VolumeClass, error) {
	cfgs := make(VolumeClasses)
	item, err := GetItemByName("volumes", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllVolumeClasses loads all Volume Classes
func LoadAllVolumeClasses() (VolumeClasses, error) {
	cfgs := make(VolumeClasses)
	items, err := GetItemsByType("volumes")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB int a Volume Class
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
