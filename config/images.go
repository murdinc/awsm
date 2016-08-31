package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type ImageClasses map[string]ImageClass

type ImageClass struct {
	Propagate        bool     `json:"propagate"`
	PropagateRegions []string `json:"propagateRegions"`
	Retain           int      `json:"retain"`
	InstanceId       string   `json:"instanceId"`
}

func DefaultImageClasses() ImageClasses {
	defaultImages := make(ImageClasses)

	defaultImages["base"] = ImageClass{
		Propagate:        true,
		Retain:           5,
		PropagateRegions: []string{"us-west-2", "us-east-1", "eu-west-1"},
	}

	return defaultImages
}

func LoadImageClass(name string) (ImageClass, error) {
	cfgs := make(ImageClasses)
	item, err := GetItemByName("images", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllImageClasses() (ImageClasses, error) {
	cfgs := make(ImageClasses)
	items, err := GetItemsByType("images")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c *ImageClasses) Marshal(items []*simpledb.Item) {

	cfgs := make(ImageClasses)

	for _, item := range items {
		name := strings.Replace(*item.Name, "images/", "", -1)
		cfg := new(ImageClass)

		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "Propagate":
				cfg.Propagate, _ = strconv.ParseBool(val)

			case "Retain":
				cfg.Retain, _ = strconv.Atoi(val)

			case "PropagateRegions":
				cfg.PropagateRegions = append(cfg.PropagateRegions, val)

			case "InstanceId":
				cfg.InstanceId = val

			}
		}

		cfgs[name] = *cfg
	}

	c = &cfgs
}
