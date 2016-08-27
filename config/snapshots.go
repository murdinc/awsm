package config

import (
	"strconv"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type SnapshotClassConfigs map[string]SnapshotClassConfig

type SnapshotClassConfig struct {
	Retain           int
	Propagate        bool
	PropagateRegions []string
	VolumeId         string
}

func DefaultSnapshotClasses() SnapshotClassConfigs {
	defaultSnapshots := make(SnapshotClassConfigs)

	defaultSnapshots["git"] = SnapshotClassConfig{
		Propagate:        true,
		Retain:           5,
		PropagateRegions: []string{"us-west-2", "us-east-1", "eu-west-1"},
	}

	defaultSnapshots["mysql-data"] = SnapshotClassConfig{
		Propagate:        true,
		Retain:           5,
		PropagateRegions: []string{"us-west-2", "us-east-1", "eu-west-1"},
	}

	return defaultSnapshots
}

func (c *SnapshotClassConfig) LoadConfig(class string) error {

	data, err := GetClassConfig("snapshots", class)
	if err != nil {
		return err
	}

	c.Marshal(data.Attributes)

	return nil

}

func (c *SnapshotClassConfig) Marshal(attributes []*simpledb.Attribute) {
	for _, attribute := range attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "Propagate":
			c.Propagate, _ = strconv.ParseBool(val)

		case "Retain":
			c.Retain, _ = strconv.Atoi(val)

		case "PropagateRegions":
			c.PropagateRegions = append(c.PropagateRegions, val)

		case "VolumeId":
			c.VolumeId = val
		}
	}

}
