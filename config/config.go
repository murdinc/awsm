package config

import (
	"errors"

	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/awsm/terminal"
)

type awsmConfig struct {
}

func GetClassConfig(configType, configClass string) (*simpledb.GetAttributesOutput, error) {
	// Check for the awsm db
	if !CheckDB() {
		create := terminal.BoxPromptBool("No awsm database found!", "Do you want to create one now?")
		if !create {
			terminal.Information("Ok then, maybe next time.. ")
			return nil, errors.New("No awsm database!")

		}
		err := CreateAwsmDatabase()
		if err != nil {
			return nil, err
		}
	}

	// Check for the class requested
	data, err := GetItemByName(configType + "/" + configClass)
	if err != nil {
		return nil, err
	}

	return data, nil
}
