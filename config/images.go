package config

type ImageClassConfigs map[string]ImageClassConfig

type ImageClassConfig struct {
	Propagate bool
	Retain    int
}

func DefaultImageClasses() ImageClassConfigs {
	defaultImages := make(ImageClassConfigs)

	defaultImages["base"] = ImageClassConfig{
		Propagate: true,
		Retain:    5,
	}

	return defaultImages
}

func (c *ImageClassConfig) LoadConfig(class string) error {
	/*
		data, err := GetClassConfig("ec2", class)
		if err != nil {
			return err
		}

		for _, attribute := range data.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "InstanceType":
				c.InstanceType = val

			case "SecurityGroups":
				c.SecurityGroups = append(c.SecurityGroups, val)

			case "Subnet":
				c.Subnet = val

			case "PublicIpAddress":
				c.PublicIpAddress, _ = strconv.ParseBool(val)

			case "AMI":
				c.AMI = val

			case "Keys":
				c.Keys = append(c.SecurityGroups, val)

			}
		}
	*/
	return nil

}
