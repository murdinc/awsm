package config

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/awsm/terminal"
)

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

func GetAllConfigNames(configType string) ([]string, error) {
	// Check for the awsm db
	if !CheckDB() {
		return nil, nil
	}

	// Get the configs
	data, err := GetItemsByType(configType)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(data.Items))
	for i, item := range data.Items {
		names[i] = strings.Replace(*item.Name, configType+"/", "", -1)
	}

	return names, nil
}

func GetAllClassConfigs(configType string) (*simpledb.SelectOutput, error) {
	// Check for the awsm db
	if !CheckDB() {
		return nil, nil
	}

	// Get the configs
	data, err := GetItemsByType(configType)
	if err != nil {
		return nil, err
	}

	return data, nil
}
