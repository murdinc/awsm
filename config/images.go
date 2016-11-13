package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// ImageClasses is a map of Image classes
type ImageClasses map[string]ImageClass

// ImageClass is a single Image class
type ImageClass struct {
	InstanceID       string   `json:"instanceID" awsmClass:"Instance ID"`
	Rotate           bool     `json:"rotate" awsmClass:"Rotate"`
	Retain           int      `json:"retain" awsmClass:"Retain"`
	Propagate        bool     `json:"propagate" awsmClass:"Propagate"`
	PropagateRegions []string `json:"propagateRegions" awsmClass:"Propagate Regions"`
}

// DefaultImageClasses returns the default Image classes
func DefaultImageClasses() ImageClasses {
	defaultImages := make(ImageClasses)

	defaultImages["base"] = ImageClass{
		Rotate:           true,
		Retain:           5,
		Propagate:        true,
		PropagateRegions: []string{"us-west-2", "us-east-1", "eu-west-1"},
		InstanceID:       "",
	}
	defaultImages["hvm-base"] = ImageClass{
		Rotate:           true,
		Retain:           5,
		Propagate:        true,
		PropagateRegions: []string{"us-west-2", "us-east-1", "eu-west-1"},
		InstanceID:       "",
	}

	return defaultImages
}

// SaveImageClass reads unmarshals a byte slice and inserts it into the db
func SaveImageClass(className string, data []byte) (class ImageClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = Insert("images", ImageClasses{className: class})
	return
}

// LoadImageClass returns a single Image class by its name
func LoadImageClass(name string) (ImageClass, error) {
	cfgs := make(ImageClasses)
	item, err := GetItemByName("images", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllImageClasses returns all Image classes
func LoadAllImageClasses() (ImageClasses, error) {
	cfgs := make(ImageClasses)
	items, err := GetItemsByType("images")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into Image Classes
func (c ImageClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "images/", "", -1)
		cfg := new(ImageClass)

		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "Propagate":
				cfg.Propagate, _ = strconv.ParseBool(val)

			case "PropagateRegions":
				cfg.PropagateRegions = append(cfg.PropagateRegions, val)

			case "Rotate":
				cfg.Rotate, _ = strconv.ParseBool(val)

			case "Retain":
				cfg.Retain, _ = strconv.Atoi(val)

			case "InstanceID":
				cfg.InstanceID = val

			}
		}
		c[name] = *cfg
	}
}
