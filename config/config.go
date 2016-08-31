package config

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/terminal"
)

func InsertClassConfigs(classType string, classInterface interface{}) error {

	var itemName string
	itemsMap := make(map[string][]*simpledb.ReplaceableAttribute)
	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	// Build Attributes
	switch classType {
	case "vpcs":
		for class, config := range classInterface.(VpcClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "subnets":
		for class, config := range classInterface.(SubnetClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "instances":
		for class, config := range classInterface.(InstanceClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "volumes":
		for class, config := range classInterface.(VolumeClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "snapshots":
		for class, config := range classInterface.(SnapshotClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "images":
		for class, config := range classInterface.(ImageClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "autoscalinggroups":
		for class, config := range classInterface.(AutoscaleGroupClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "launchconfigurations":
		for class, config := range classInterface.(LaunchConfigurationClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "scalingpolicies":
		for class, config := range classInterface.(ScalingPolicyClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "alarms":
		for class, config := range classInterface.(AlarmClassConfigs) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	/*
		case "securitygroups":
			for class, config := range classInterface.(SecurityGroupClassConfigs) {
				itemName = classType + "/" + class
				itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config)...)
			}
	*/

	default:
		return errors.New("InsertClassConfigs does not have switch for [" + classType + "]! No configurations of this type are being installed!")

	}

	items := make([]*simpledb.ReplaceableItem, len(itemsMap))

	for item, attributes := range itemsMap {

		terminal.Information("Building Configuration for [" + item + "]...")

		i := &simpledb.ReplaceableItem{
			Attributes: attributes,
			Name:       aws.String(item),
		}
		items = append(items, i)

	}

	params := &simpledb.BatchPutAttributesInput{
		DomainName: aws.String("awsm"),
		Items:      items,
	}
	terminal.Information("Installing [" + classType + "] Configurations...")
	_, err := svc.BatchPutAttributes(params)

	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil

}

func LoadAllClasses(classType string) (configs interface{}, err error) {

	switch classType {

	case "vpcs":
		return LoadAllVpcClasses()

	case "subnets":
		return LoadAllSubnetClasses()

	case "instances":
		return LoadAllInstanceClasses()

	case "volumes":
		return LoadAllVolumeClasses()

	case "snapshots":
		return LoadAllSnapshotClasses()

	case "images":
		return LoadAllImageClasses()

	case "autoscalinggroups":
		return LoadAllAutoscalingGroupClasses()

	case "launchconfigurations":
		return LoadAllLaunchConfigurationClasses()

	case "scalingpolicies":
		return LoadAllScalingPolicyClasses()

	case "alarms":
		return LoadAllAlarmClasses()

	default:
		err = errors.New("LoadAllClasses does not have switch for [" + classType + "]! No configurations of this type are being loaded!")

	}

	return configs, err
}

func LoadClassByName(classType, className string) (configs interface{}, err error) {

	switch classType {

	case "vpcs":
		return LoadVpcClass(className)

	case "subnets":
		return LoadSubnetClass(className)

	case "instances":
		return LoadInstanceClass(className)

	case "volumes":
		return LoadVolumeClass(className)

	case "snapshots":
		return LoadSnapshotClass(className)

	case "images":
		return LoadImageClass(className)

	case "autoscalinggroups":
		return LoadAutoscalingGroupClass(className)

	case "launchconfigurations":
		return LoadLaunchConfigurationClass(className)

	case "scalingpolicies":
		return LoadScalingPolicyClass(className)

	case "alarms":
		return LoadAlarmClass(className)

	default:
		err = errors.New("LoadClassByName does not have switch for [" + classType + "]! No configuration of this type is being loaded!")

	}

	return configs, err
}

func LoadAllClassNames(classType string) ([]string, error) {
	// Check for the awsm db
	if !CheckDB() {
		return nil, nil
	}

	// Get the configs
	items, err := GetItemsByType(classType)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(items))
	for i, item := range items {
		names[i] = strings.Replace(*item.Name, classType+"/", "", -1)
	}

	return names, nil
}
