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
	case "vpc":
		for class, config := range configInterface.(VpcClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, configType)...)
		}

	case "subnet":
		for class, config := range configInterface.(SubnetClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, configType)...)
		}

	case "ec2":
		for class, config := range configInterface.(InstanceClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, configType)...)
		}

	case "ebs":
		for class, config := range configInterface.(VolumeClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, configType)...)
		}

	case "ami":
		for class, config := range configInterface.(ImageClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, configType)...)
		}

	case "autoscale":
		for class, config := range configInterface.(AutoScaleGroupClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, configType)...)
		}

	case "scalingpolicy":
		for class, config := range configInterface.(ScalingPolicyConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, configType)...)
		}

	case "alarm":
		for class, config := range configInterface.(AlarmClassConfigs) {
			itemName = configType + "/" + class
			itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config, configType)...)
		}

	default:
		terminal.ErrorLine("InsertClassConfigs does not have switch for [" + configType + "]! No configurations of this type are being installed!")

		/*
			case "securitygroup":
				for class, config := range configInterface.(AlarmClassConfigs) {
					itemName = configType + "/" + class
					itemsMap[itemName] = append(itemsMap[itemName], BuildAttributes(config)...)
				}
		*/
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

func GetItemByName(configName string) (*simpledb.GetAttributesOutput, error) {

	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	params := &simpledb.GetAttributesInput{
		DomainName:     aws.String("awsm"),
		ItemName:       aws.String(configName),
		ConsistentRead: aws.Bool(true),
	}
	resp, err := svc.GetAttributes(params)

	if err != nil {
		return &simpledb.GetAttributesOutput{}, err
	}

	if len(resp.Attributes) < 1 {
		return &simpledb.GetAttributesOutput{}, errors.New("Unable to find the [" + configName + "] class in the database!")
	}

	return resp, nil
}

func GetItemsByType(configType string) (*simpledb.SelectOutput, error) {

	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	params := &simpledb.SelectInput{
		SelectExpression: aws.String(fmt.Sprintf("select * from awsm where ConfigType = '%s'", configType)), // Required
		ConsistentRead:   aws.Bool(true),
		//NextToken:        aws.String("String"),
	}

	//fmt.Println(params)

	resp, err := svc.Select(params)

	if err != nil {
		return &simpledb.SelectOutput{}, err
	}

	if len(resp.Items) < 1 {
		return &simpledb.SelectOutput{}, errors.New("Unable to find the [" + configType + "] class in the database!")
	}

	return resp, nil
}

func CreateAwsmDatabase() error {

	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	params := &simpledb.CreateDomainInput{
		DomainName: aws.String("awsm"),
	}
	_, err := svc.CreateDomain(params)

	if err != nil {
		return err
	}

	// Insert our default configs
	InsertClassConfigs("vpc", DefaultVpcClasses())
	InsertClassConfigs("subnet", DefaultSubnetClasses())
	InsertClassConfigs("ec2", DefaultInstanceClasses())
	InsertClassConfigs("alarm", DefaultAlarms())
	InsertClassConfigs("ami", DefaultImageClasses())
	InsertClassConfigs("scalingpolicy", DefaultScalingPolicies())
	InsertClassConfigs("ebs", DefaultVolumeClasses())
	InsertClassConfigs("autoscale", DefaultAutoScaleGroupClasses())
	//InsertClassConfigs("securitygroup", DefaultSecurityGroupClasses())

	return nil
}

func BuildAttributes(config interface{}, configType string) []*simpledb.ReplaceableAttribute {
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

	attributes = append(attributes, &simpledb.ReplaceableAttribute{
		Name:    aws.String("ConfigType"),
		Value:   aws.String(configType),
		Replace: aws.Bool(true),
	})

	//fmt.Println(attributes)

	return attributes
}
