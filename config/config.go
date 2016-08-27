package config

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/terminal"
)

func LoadAllConfigs(configType string) (configs map[string]interface{}, err error) {
	data, err := GetAllClassConfigs(configType)
	if err != nil {
		return nil, err
	}

	configs = make(map[string]interface{})

	// Build Configs
	switch configType {

	case "vpcs":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(VpcClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}
	case "subnets":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(SubnetClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}
	case "instances":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(InstanceClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}
	case "volumes":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(VolumeClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}
	case "snapshots":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(SnapshotClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}
	case "images":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(ImageClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}
	case "autoscalinggroups":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(AutoScaleGroupClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}
	case "launchconfigurations":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(LaunchConfigurationClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}
	case "scalingpolicies":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(ScalingPolicyClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}
	case "alarms":
		for _, item := range data.Items {
			name := strings.Replace(*item.Name, configType+"/", "", -1)
			cfg := new(AlarmClassConfig)
			cfg.Marshal(item.Attributes)
			configs[name] = *cfg
		}

	default:
		err = errors.New("LoadAllConfigs does not have switch for [" + configType + "]! No configurations of this type are being loaded!")

	}

	return configs, err
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
		return nil, errors.New("No database found!")
	}

	// Get the configs
	data, err := GetItemsByType(configType)
	if err != nil {
		return nil, err
	}

	return data, nil
}
