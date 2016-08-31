package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type SnapshotClasses map[string]SnapshotClass

type SnapshotClass struct {
	Retain           int      `json:"retain"`
	Propagate        bool     `json:"propagate"`
	PropagateRegions []string `json:"propagateRegions"`
	VolumeId         string   `json:"volumeId"`
}

func DefaultSnapshotClasses() SnapshotClasses {
	defaultSnapshots := make(SnapshotClasses)

	defaultSnapshots["git"] = SnapshotClass{
		Propagate:        true,
		Retain:           5,
		PropagateRegions: []string{"us-west-2", "us-east-1", "eu-west-1"},
	}

	defaultSnapshots["mysql-data"] = SnapshotClass{
		Propagate:        true,
		Retain:           5,
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
			}
		}
		c[name] = *cfg
	}
}
