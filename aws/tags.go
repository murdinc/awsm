package aws

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// GetValue returns the value of a
func GetTagValue(key string, tags interface{}) string {
	switch v := tags.(type) {
	case []*ec2.Tag:
		for _, tag := range v {
			if aws.StringValue(tag.Key) == key {
				return aws.StringValue(tag.Value)
			}
		}
	}
	return ""
}

func SetEc2NameAndClassTags(resource *string, name, class, region string) error {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

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
