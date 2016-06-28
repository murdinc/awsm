package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/terminal"
)

type AZs []AZ

type AZ struct {
	Name   string
	Region string
	State  string
}

func AZList() []string {
	azs, _ := GetAZs()
	azlist := make([]string, len(*azs))

	for i, az := range *azs {
		azlist[i] = az.Name
	}
	return azlist
}

func GetAZs() (*AZs, []error) {
	var wg sync.WaitGroup
	var errs []error

	azList := new(AZs)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionAZs(*region.RegionName, azList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering az list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}

	wg.Wait()

	return azList, errs
}

func GetRegionAZs(region string, azList *AZs) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
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

func (a *AZs) ValidAZ(az string) bool {
	for _, vaz := range *a {
		if az == vaz.Name {
			return true
		}
	}
	return false
}

func (a *AZs) GetRegion(az string) string {
	for _, vaz := range *a {
		if az == vaz.Name {
			return vaz.Region
		}
	}
	return ""
}
