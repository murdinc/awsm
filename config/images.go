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
