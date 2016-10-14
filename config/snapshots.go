package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type SnapshotClasses map[string]SnapshotClass

type SnapshotClass struct {
	Retain           int      `json:"retain" awsmList:"Retain"`
	Rotate           bool     `json:"rotate" awsmList:"Rotate"`
	Propagate        bool     `json:"propagate" awsmList:"Propagate"`
	PropagateRegions []string `json:"propagateRegions" awsmList:"Propagate Regions"`
	VolumeId         string   `json:"volumeId" awsmList:"Volume ID"`
}

func DefaultSnapshotClasses() SnapshotClasses {
	defaultSnapshots := make(SnapshotClasses)

	defaultSnapshots["git"] = SnapshotClass{
		Propagate:        true,
		Retain:           5,
		Rotate:           true,
		PropagateRegions: []string{"us-west-2", "us-east-1", "eu-west-1"},
	}

	defaultSnapshots["mysql-data"] = SnapshotClass{
		Propagate:        true,
		Retain:           5,
		Rotate:           true,
		PropagateRegions: []string{"us-west-2", "us-east-1", "eu-west-1"},
	}

	return defaultSnapshots
}

func LoadSnapshotClass(name string) (SnapshotClass, error) {
	cfgs := make(SnapshotClasses)
	item, err := GetItemByName("snapshots", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllSnapshotClasses() (SnapshotClasses, error) {
	cfgs := make(SnapshotClasses)
	items, err := GetItemsByType("snapshots")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c SnapshotClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "snapshots/", "", -1)
		cfg := new(SnapshotClass)
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

			case "Rotate":
				cfg.Rotate, _ = strconv.ParseBool(val)
			}
		}
		c[name] = *cfg
	}
}
