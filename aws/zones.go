package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
)

// AZs represents a slice of Availability Zones
type AZs []AZ

// AZ represents a single Availability Zone
type AZ models.AZ

// AZList returns a slice of Availability Zone names (strings)
func AZList() []string {
	azs, _ := GetAZs()
	azlist := make([]string, len(*azs))

	for i, az := range *azs {
		azlist[i] = az.Name
	}
	return azlist
}

// GetAZs returns a slice of Availability Zones
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

// GetRegionAZs returns a slice of a regions Availability Zones into the provided AZs
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

// ValidAZ returns true if the provided string is a valid Availability Zone
func (a *AZs) ValidAZ(az string) bool {
	for _, vaz := range *a {
		if az == vaz.Name {
			return true
		}
	}
	return false
}

// GetRegion returns the region of a provided Availability Zone
func (a *AZs) GetRegion(az string) string {
	for _, vaz := range *a {
		if az == vaz.Name {
			return vaz.Region
		}
	}
	return ""
}

// GetRegionMap returns a map of Regions and their Availability Zones
func (a *AZs) GetRegionMap(azList []string) map[string][]string {
	azs, _ := GetAZs()

	regionMap := make(map[string][]string)

	// Get the list of regions from the AZs
	for _, az := range azList {
		region := azs.GetRegion(az)
		regionMap[region] = append(regionMap[region], az)

	}

	return regionMap
}
