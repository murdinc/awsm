package config

import (
	"errors"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/awsm/aws/regions"
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

			// Delete the existing listeners
			DeleteItemsByType(classType + "/" + class + "/listeners")

			// Load Balancer Listeners
			for _, listener := range config.Listeners {
				itemName = classType + "/" + class + "/listeners/" + uuid.NewV4().String()
				itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(listener, classType+"/"+class+"/listeners")...)
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

			// Delete the existing grants
			DeleteItemsByType(classType + "/" + class + "/grants")

			// Security Group Grants
			for _, grant := range config.SecurityGroupGrants {
				itemName = classType + "/" + class + "/grants/" + uuid.NewV4().String()
				itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(grant, classType+"/"+class+"/grants")...)
			}
		}

	case "keypairs":
		for class, config := range classInterface.(KeyPairClasses) {
			itemName = classType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
		}

	case "widgets":
		for widget, config := range classInterface.(Widgets) {
			itemName = classType + "/" + widget
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, classType)...)
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

// Export exports all configurations
func Export() (export map[string]interface{}, err error) {

	export = make(map[string]interface{})
	export["vpcs"], _ = LoadAllVpcClasses()
	export["subnets"], _ = LoadAllSubnetClasses()
	export["instances"], _ = LoadAllInstanceClasses()
	export["volumes"], _ = LoadAllVolumeClasses()
	export["snapshots"], _ = LoadAllSnapshotClasses()
	export["images"], _ = LoadAllImageClasses()
	export["autoscalegroups"], _ = LoadAllAutoscalingGroupClasses()
	export["launchconfigurations"], _ = LoadAllLaunchConfigurationClasses()
	export["loadbalancers"], _ = LoadAllLoadBalancerClasses()
	export["scalingpolicies"], _ = LoadAllScalingPolicyClasses()
	export["alarms"], _ = LoadAllAlarmClasses()
	export["securitygroups"], _ = LoadAllSecurityGroupClasses()
	export["keypairs"], _ = LoadAllKeyPairClasses()
	export["widgets"], _ = LoadAllWidgets()

	return
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

	case "keypairs":
		return LoadAllKeyPairClasses()

		/*
			case "addresses":
				return LoadAllAddresses()
		*/

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

	case "keypairs":
		return LoadKeyPairClass(className)

	default:
		err = errors.New("LoadClassByName does not have switch for [" + classType + "]! No class configuration of this type is being loaded!")

	}

	return configs, err
}

// LoadAllClassOptions loads all class options by a type
func LoadAllClassOptions(classType string) (options map[string]interface{}, err error) {

	var wg sync.WaitGroup
	var mu sync.Mutex
	var classOptionKeys []string

	options = make(map[string]interface{})

	switch classType {

	case "vpcs":

	case "subnets":

	case "instances":
		classOptionKeys = []string{"securitygroups", "volumes", "vpcs", "subnets", "images", "keypairs", "iamusers"}

	case "volumes":
		classOptionKeys = []string{"snapshots"}

	case "snapshots":
		classOptionKeys = []string{"regions"}

	case "images":
		classOptionKeys = []string{"regions"}

	case "autoscalegroups":
		classOptionKeys = []string{"launchconfigurations", "zones", "subnets", "scalingpolicies", "alarms", "loadbalancers"}

	case "launchconfigurations":
		classOptionKeys = []string{"regions", "instances"}

	case "loadbalancers":
		classOptionKeys = []string{"securitygroups", "subnets", "zones"}

	case "scalingpolicies":

	case "alarms":
		classOptionKeys = []string{"scalingpolicies"} // TODO: don't limit to only scaling policies?

	case "securitygroups":

	case "keypairs":

	default:
		err = errors.New("LoadAllClassOptions does not have switch for [" + classType + "]! No options of this type are being loaded!")
	}

	for _, key := range classOptionKeys {
		wg.Add(1)

		go func(key string) {
			defer wg.Done()
			mu.Lock()

			switch key {
			case "regions":
				options[key] = regions.GetRegionNameList()

			case "zones":
				options[key] = regions.GetAZNameList()

			default:
				options[key], _ = LoadAllClassNames(key)

			}

			mu.Unlock()
		}(key)
	}

	wg.Wait()

	return options, err
}

// LoadAllClassNames loads all class names of a type
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
