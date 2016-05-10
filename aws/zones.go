package aws

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/terminal"
)

type AZs []AZ

type AZ struct {
	Name   string
	Region string
	State  string
}

func GetAZs() (*AZs, error) {
	var wg sync.WaitGroup

	azList := new(AZs)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionAZs(region.RegionName, azList)
			if err != nil {
				terminal.ShowErrorMessage("Error gathering instance list", err.Error())
			}
		}(region)
	}

	wg.Wait()

	return azList, nil

}

func GetRegionAZs(region *string, azList *AZs) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeAvailabilityZones(nil)

	if err != nil {
		return err
	}

	azs := make(AZs, len(result.AvailabilityZones))
	for i, az := range result.AvailabilityZones {
		azs[i] = AZ{
			Name:   aws.StringValue(az.ZoneName),
			Region: aws.StringValue(az.RegionName),
			State:  aws.StringValue(az.State),
		}
	}

	*azList = append(*azList, azs[:]...)

	return nil
}

func (a *AZs) ValidateAZ(az string) bool {
	for _, vaz := range *a {
		if az == vaz.Name {
			return true
		}
	}
	return false
}
