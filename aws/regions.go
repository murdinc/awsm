package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetRegionList() []*ec2.Region {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String("us-east-1")}))

	resp, err := svc.DescribeRegions(nil)

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return resp.Regions
}

func ValidRegion(region string) bool {
	vregions := GetRegionList()
	for _, vregion := range vregions {
		if region == aws.StringValue(vregion.RegionName) {
			return true
		}
	}
	return false
}
