package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/simpledb"
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

func GetItemByName(classType, className string) (*simpledb.Item, error) {

	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	params := &simpledb.GetAttributesInput{
		DomainName:     aws.String("awsm"),
		ItemName:       aws.String(classType + "/" + className),
		ConsistentRead: aws.Bool(true),
	}
	resp, err := svc.GetAttributes(params)

	if err != nil {
		return &simpledb.Item{}, err
	}

	if len(resp.Attributes) < 1 {
		return &simpledb.Item{}, errors.New("Unable to find the [" + className + "] class in the database!")
	}

	item := &simpledb.Item{
		Name:       aws.String(classType + "/" + className),
		Attributes: resp.Attributes,
	}

	return item, nil
}

func GetItemsByType(classType string) ([]*simpledb.Item, error) {

	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference

	params := &simpledb.SelectInput{
		SelectExpression: aws.String(fmt.Sprintf("select * from awsm where classType = '%s'", classType)),
		ConsistentRead:   aws.Bool(true),
		//NextToken:        aws.String("String"),
	}

	resp, err := svc.Select(params)

	if err != nil {
		return []*simpledb.Item{}, err
	}

	if len(resp.Items) < 1 {
		return []*simpledb.Item{}, errors.New("Unable to find the [" + classType + "] class in the database!")
	}

	return resp.Items, nil
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
	InsertClasses("securitygroups", DefaultSecurityGroupClasses())
	InsertClasses("vpcs", DefaultVpcClasses())
	InsertClasses("subnets", DefaultSubnetClasses())
	InsertClasses("instances", DefaultInstanceClasses())
	InsertClasses("alarms", DefaultAlarms())
	InsertClasses("images", DefaultImageClasses())
	InsertClasses("scalingpolicies", DefaultScalingPolicyClasses())
	InsertClasses("launchconfigurations", DefaultLaunchConfigurationClasses())
	InsertClasses("volumes", DefaultVolumeClasses())
	InsertClasses("snapshots", DefaultSnapshotClasses())
	InsertClasses("autoscalinggroups", DefaultAutoscaleGroupClasses())

	return nil
}

func BuildAttributes(class interface{}, classType string) []*simpledb.ReplaceableAttribute {

	typ := reflect.TypeOf(class)
	val := reflect.ValueOf(class)

	var attributes []*simpledb.ReplaceableAttribute

	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name

		// Ignore if tagged ignore
		if typ.Field(i).Tag.Get("awsm") == "ignore" {
			continue
		}

		switch val.Field(i).Interface().(type) {
		case int:
			attributes = append(attributes, &simpledb.ReplaceableAttribute{
				Name:    aws.String(name),
				Value:   aws.String(fmt.Sprint(val.Field(i).Int())),
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
		Name:    aws.String("classType"),
		Value:   aws.String(classType),
		Replace: aws.Bool(true),
	})

	fmt.Println(attributes)

	return attributes
}
