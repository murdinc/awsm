package config

import "strconv"

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

func (c *ImageClassConfig) LoadConfig(class string) error {

	data, err := GetClassConfig("ami", class)
	if err != nil {
		return err
	}

	for _, attribute := range data.Attributes {

		val := *attribute.Value

		switch *attribute.Name {

		case "Propagate":
			c.Propagate, _ = strconv.ParseBool(val)

		case "Retain":
			c.Retain, _ = strconv.Atoi(val)

		case "PropagateRegions":
			c.PropagateRegions = append(c.PropagateRegions, val)

		case "InstanceId":
			c.InstanceId = val

		}
	}

	return nil

}
