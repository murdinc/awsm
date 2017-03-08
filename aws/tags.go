package aws

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
)

// GetTagValue returns the tag with the given key if available.
func GetTagValue(key string, tags interface{}) string {
	switch v := tags.(type) {
	case []*ec2.Tag:
		for _, tag := range v {
			if aws.StringValue(tag.Key) == key {
				return aws.StringValue(tag.Value)
			}
		}
	case []*elb.Tag:
		for _, tag := range v {
			if aws.StringValue(tag.Key) == key {
				return aws.StringValue(tag.Value)
			}
		}

	}

	return ""
}

// SetEc2NameAndClassTags sets the Name and Class tags of an EC2 asset
func SetEc2NameAndClassTags(resource *string, name, class, region string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.CreateTagsInput{
		Resources: []*string{
			resource,
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
			{
				Key:   aws.String("Class"),
				Value: aws.String(class),
			},
		},
	}
	_, err := svc.CreateTags(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}
	return nil
}
