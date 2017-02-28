package config

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/terminal"
)

// CheckDB checks for an awsm database
func CheckDB() bool {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference
	svc := simpledb.New(sess)

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

// GetItemByName gets a SimpleDB item by its type and name
func GetItemByName(classType, className string) (*simpledb.Item, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference
	svc := simpledb.New(sess)

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

// GetItemsByType returns all SimpleDB items by class type
func GetItemsByType(classType string) ([]*simpledb.Item, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference
	svc := simpledb.New(sess)

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

// DeleteItemsByType batch deletes classes from SimpleDB
func DeleteItemsByType(classType string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference
	svc := simpledb.New(sess)

	existingItems, err := GetItemsByType(classType)
	if err != nil {
		return err
	}

	params := &simpledb.BatchDeleteAttributesInput{
		DomainName: aws.String("awsm"),
		//Items:      deleteList,
	}

	for _, item := range existingItems {
		itemName := aws.StringValue(item.Name)
		params.Items = append(params.Items, &simpledb.DeletableItem{
			Name: item.Name,
		})

		terminal.Delta("Deleting [" + classType + "/" + itemName + "] Configuration...")
	}

	_, err = svc.BatchDeleteAttributes(params)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// CreateAwsmDatabase creates an awsm SimpleDB Domain
func CreateAwsmDatabase(generateAwsmKeyPair bool) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference
	svc := simpledb.New(sess)

	params := &simpledb.CreateDomainInput{
		DomainName: aws.String("awsm"),
	}
	_, err := svc.CreateDomain(params)

	if err != nil {
		return err
	}

	// Insert our default configs
	Insert("securitygroups", DefaultSecurityGroupClasses())
	Insert("vpcs", DefaultVpcClasses())
	Insert("subnets", DefaultSubnetClasses())
	Insert("instances", DefaultInstanceClasses())
	Insert("alarms", DefaultAlarms())
	Insert("images", DefaultImageClasses())
	Insert("scalingpolicies", DefaultScalingPolicyClasses())
	Insert("launchconfigurations", DefaultLaunchConfigurationClasses())
	Insert("loadbalancers", DefaultLoadBalancerClasses())
	Insert("volumes", DefaultVolumeClasses())
	Insert("snapshots", DefaultSnapshotClasses())
	Insert("autoscalegroups", DefaultAutoscaleGroupClasses())
	Insert("keypairs", DefaultKeyPairClasses(generateAwsmKeyPair))
	Insert("widgets", DefaultWidgets())

	return nil
}

// BuildAttributes builds SimpleDB item attributes from class structs
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

		case float64:
			attributes = append(attributes, &simpledb.ReplaceableAttribute{
				Name:    aws.String(name),
				Value:   aws.String(fmt.Sprint(val.Field(i).Float())),
				Replace: aws.Bool(true),
			})

		case string:
			attributes = append(attributes, &simpledb.ReplaceableAttribute{
				Name:    aws.String(name),
				Value:   aws.String(val.Field(i).String()),
				Replace: aws.Bool(true),
			})

		case []string, []byte:
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

		case time.Time:
			t := val.Field(i).Interface().(time.Time).UTC().String()
			attributes = append(attributes, &simpledb.ReplaceableAttribute{
				Name:    aws.String(name),
				Value:   aws.String(t),
				Replace: aws.Bool(true),
			})

		case []SecurityGroupGrant, []LoadBalancerListener:
			// Handled in config/classes.go, for now

		default:
			println("BuildAttributes does not have a switch for type:")
			println(val.Field(i).Type().String())

		}
	}

	attributes = append(attributes, &simpledb.ReplaceableAttribute{
		Name:    aws.String("classType"),
		Value:   aws.String(classType),
		Replace: aws.Bool(true),
	})

	return attributes
}
