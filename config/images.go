package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type ImageClassConfigs map[string]ImageClassConfig

type ImageClassConfig struct {
	Propagate        bool
	PropagateRegions []string
	Retain           int
	InstanceId       string
}

func DefaultImageClasses() ImageClassConfigs {
	defaultImages := make(ImageClassConfigs)

	defaultImages["base"] = ImageClassConfig{
		Propagate:        true,
		Retain:           5,
		PropagateRegions: []string{"us-west-2", "us-east-1", "eu-west-1"},
	}

	return defaultImages
}

func LoadImageClass(name string) (ImageClassConfig, error) {
	cfgs := make(ImageClassConfigs)
	item, err := GetItemByName("images", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllImageClasses() (ImageClassConfigs, error) {
	cfgs := make(ImageClassConfigs)
	items, err := GetItemsByType("images")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c *ImageClassConfigs) Marshal(items []*simpledb.Item) {

	cfgs := make(ImageClassConfigs)

	for _, item := range items {
		name := strings.Replace(*item.Name, "images/", "", -1)
		cfg := new(ImageClassConfig)

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
