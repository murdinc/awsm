package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/awsm/terminal"
)

func CheckDB() bool {
	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	params := &simpledb.DomainMetadataInput{
		DomainName: aws.String("awsm"), // Required
	}
	_, err := svc.DomainMetadata(params)

	if err != nil {
		return false
	}

	// TODO handle the response stats?

	return true
}

func InsertClassConfigs(configType string, configInterface interface{}) error {

	var itemName string
	itemsMap := make(map[string][]*simpledb.ReplaceableAttribute)
	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	// Build Attributes
	switch configType {
	case "ec2":
		for class, config := range configInterface.(InstanceClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config)...)
		}

	case "ebs":
		for class, config := range configInterface.(VolumeClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config)...)
		}

	case "ami":
		for class, config := range configInterface.(ImageClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config)...)
		}

	case "autoscale":
		for class, config := range configInterface.(AutoScaleGroupClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config)...)
		}

	case "scalingpolicy":
		for class, config := range configInterface.(ScalingPolicyConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config)...)
		}

	case "alarm":
		for class, config := range configInterface.(AlarmClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config)...)
		}

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
	terminal.Information("Installing [" + configType + "] Configurations...")
	_, err := svc.BatchPutAttributes(params)

	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil

}

func DeleteConfig() {

}

func UpdateConfig() {

}

func SelectConfig(configType, configClass string) (*simpledb.GetAttributesOutput, error) {

	itemName := configType + "/" + configClass

	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	params := &simpledb.GetAttributesInput{
		DomainName:     aws.String("awsm"),
		ItemName:       aws.String(itemName),
		ConsistentRead: aws.Bool(true),
	}
	resp, err := svc.GetAttributes(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return &simpledb.GetAttributesOutput{}, err
	}

	if len(resp.Attributes) < 1 {
		return &simpledb.GetAttributesOutput{}, errors.New("Unable to find the [" + itemName + "] class in the database!")
	}

	return resp, nil
}

func CreateAwsmDatabase() error {

	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	params := &simpledb.CreateDomainInput{
		DomainName: aws.String("awsm"), // Required
	}
	_, err := svc.CreateDomain(params)

	if err != nil {
		return err
	}

	// Insert our default configs
	InsertClassConfigs("ec2", DefaultInstanceClasses())
	InsertClassConfigs("alarm", DefaultAlarms())
	InsertClassConfigs("ami", DefaultImageClasses())
	InsertClassConfigs("scalingpolicy", DefaultScalingPolicies())
	InsertClassConfigs("ebs", DefaultVolumeClasses())
	InsertClassConfigs("autoscale", DefaultAutoScaleGroupClasses())
	//InsertClassConfigs("securitygroup", DefaultSecurityGroupClasses())

	return nil
}

func BuildAttributes(config interface{}) []*simpledb.ReplaceableAttribute {
	typ := reflect.TypeOf(config)
	val := reflect.ValueOf(config)

	var attributes []*simpledb.ReplaceableAttribute

	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name

		switch val.Field(i).Interface().(type) {
		case int:
			attributes = append(attributes, &simpledb.ReplaceableAttribute{
				Name:    aws.String(name),
				Value:   aws.String(val.Field(i).String()),
				Replace: aws.Bool(true),
			})

		case string:
			attributes = append(attributes, &simpledb.ReplaceableAttribute{
				Name:    aws.String(name),
				Value:   aws.String(val.Field(i).String()),
				Replace: aws.Bool(true),
			})

		case []string:
			for s := 0; s < val.Field(i).Len(); s++ {
				attributes = append(attributes, &simpledb.ReplaceableAttribute{
					Name:    aws.String(name),
					Value:   aws.String(val.Field(i).Index(s).String()),
					Replace: aws.Bool(true),
				})

			}

		case bool:
			attributes = append(attributes, &simpledb.ReplaceableAttribute{
				Name:    aws.String(name),
				Value:   aws.String(fmt.Sprint(val.Field(i).Bool())),
				Replace: aws.Bool(true),
			})
		}

	}

	return attributes
}
