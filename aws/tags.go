package aws

import (
	"github.com/aws/aws-sdk-go/aws"
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
