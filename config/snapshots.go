package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// SnapshotClasses is a map of Snapshot Classes
type SnapshotClasses map[string]SnapshotClass

// SnapshotClass is a single Snapshot Class
type SnapshotClass struct {
	Retain           int      `json:"retain" awsmClass:"Retain"`
	Rotate           bool     `json:"rotate" awsmClass:"Rotate"`
	Propagate        bool     `json:"propagate" awsmClass:"Propagate"`
	PropagateRegions []string `json:"propagateRegions" awsmClass:"Propagate Regions"`
	VolumeID         string   `json:"volumeID" awsmClass:"Volume ID"`
}

// DefaultSnapshotClasses returns the default Snapshot Classes
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

// SaveSnapshotClass reads unmarshals a byte slice and inserts it into the db
func SaveSnapshotClass(className string, data []byte) (class SnapshotClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = InsertClasses("snapshots", SnapshotClasses{className: class})
	return
}

// LoadSnapshotClass loads a Snapshot Class by its name
func LoadSnapshotClass(name string) (SnapshotClass, error) {
	cfgs := make(SnapshotClasses)
	item, err := GetItemByName("snapshots", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllSnapshotClasses loads all Snapshot Classes
func LoadAllSnapshotClasses() (SnapshotClasses, error) {
	cfgs := make(SnapshotClasses)
	items, err := GetItemsByType("snapshots")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into a Snapshot Class
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

			case "VolumeID":
				cfg.VolumeID = val

			case "Rotate":
				cfg.Rotate, _ = strconv.ParseBool(val)
			}
		}
		c[name] = *cfg
	}
}
