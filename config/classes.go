package config

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/terminal"
	"github.com/satori/go.uuid"
)

// DeleteClass deletes a class from SimpleDB
func DeleteClass(classType, className string) error {
	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	itemName := classType + "/" + className

	params := &simpledb.DeleteAttributesInput{
		DomainName: aws.String("awsm"),
		ItemName:   aws.String(itemName),
	}

	terminal.Information("Deleting [" + itemName + "] Configuration...")
	_, err := svc.DeleteAttributes(params)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// InsertClasses inserts Classes into SimpleDB
func InsertClasses(classType string, classInterface interface{}) error {

	var itemName string
	itemsMap := make(map[string][]*simpledb.ReplaceableAttribute)
	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	// Build Attributes
	switch classType {
	case "vpcs":
		for class, config := range classInterface.(VpcClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "subnets":
		for class, config := range classInterface.(SubnetClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "instances":
		for class, config := range classInterface.(InstanceClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "volumes":
		for class, config := range classInterface.(VolumeClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "snapshots":
		for class, config := range classInterface.(SnapshotClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "images":
		for class, config := range classInterface.(ImageClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "autoscalegroups":
		for class, config := range classInterface.(AutoscaleGroupClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "launchconfigurations":
		for class, config := range classInterface.(LaunchConfigurationClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "loadbalancers":
		for class, config := range classInterface.(LoadBalancerClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)

			// Load Balancer Listeners
			for _, rule := range config.Listeners {
				itemName = classType + "/" + class + "/listeners/" + uuid.NewV4().String()
				itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(rule, classType+"/"+class+"/listeners")...)
			}
		}

	case "scalingpolicies":
		for class, config := range classInterface.(ScalingPolicyClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "alarms":
		for class, config := range classInterface.(AlarmClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "securitygroups":
		for class, config := range classInterface.(SecurityGroupClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)

			// Security Group Grants
			for _, rule := range config.SecurityGroupGrants {
				itemName = classType + "/" + class + "/grants/" + uuid.NewV4().String()
				itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(rule, classType+"/"+class+"/grants")...)
			}
		}

	default:
		return errors.New("InsertClasses does not have switch for [" + classType + "]! No configurations of this type are being installed!")

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

// LoadAllClasses loads all classes of a type
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

	case "autoscalegroups":
		return LoadAllAutoscalingGroupClasses()

	case "launchconfigurations":
		return LoadAllLaunchConfigurationClasses()

	case "loadbalancers":
		return LoadAllLoadBalancerClasses()

	case "scalingpolicies":
		return LoadAllScalingPolicyClasses()

	case "alarms":
		return LoadAllAlarmClasses()

	case "securitygroups":
		return LoadAllSecurityGroupClasses()

	default:
		err = errors.New("LoadAllClasses does not have switch for [" + classType + "]! No configurations of this type are being loaded!")

	}

	return configs, err
}

// LoadClassByName loads a class by its type and name
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

	case "autoscalegroups":
		return LoadAutoscalingGroupClass(className)

	case "launchconfigurations":
		return LoadLaunchConfigurationClass(className)

	case "loadbalancers":
		return LoadLoadBalancerClass(className)

	case "scalingpolicies":
		return LoadScalingPolicyClass(className)

	case "alarms":
		return LoadAlarmClass(className)

	case "securitygroups":
		return LoadSecurityGroupClass(className)

	default:
		err = errors.New("LoadClassByName does not have switch for [" + classType + "]! No configuration of this type is being loaded!")

	}

	return configs, err
}

// LoadAllClassNames loads all class named by a type
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
