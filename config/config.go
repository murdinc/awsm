package config

import "github.com/murdinc/awsm/terminal"

type awsmConfig struct {
}

func GetClassConfig(configClass string, configType string) (map[string]string, error) {
	// Check for the awsm db
	if !CheckDB() {
		create := terminal.BoxPromptBool("No awsm database found!", "Do you want to create one now?")
		if !create {
			terminal.Information("Ok then, maybe next time.. ")
			return nil, nil
		}
		err := CreateAwsmDatabase()
		if err != nil {
			return nil, err
		}
	}

	// Check for the class requested
	_, err := SelectConfig(configType, configClass)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
