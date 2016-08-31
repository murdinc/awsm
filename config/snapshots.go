package config

import (
	"strconv"
	"strings"

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

func LoadSnapshotClass(name string) (SnapshotClassConfig, error) {
	cfgs := make(SnapshotClassConfigs)
	item, err := GetItemByName("snapshots", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllSnapshotClasses() (SnapshotClassConfigs, error) {
	cfgs := make(SnapshotClassConfigs)
	items, err := GetItemsByType("snapshots")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c SnapshotClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "snapshots/", "", -1)
		cfg := new(SnapshotClassConfig)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "Propagate":
				cfg.Propagate, _ = strconv.ParseBool(val)

			case "Retain":
				cfg.Retain, _ = strconv.Atoi(val)

			case "PropagateRegions":
				cfg.PropagateRegions = append(cfg.PropagateRegions, val)

			case "VolumeId":
				cfg.VolumeId = val
			}
		}
		c[name] = *cfg
	}
}
