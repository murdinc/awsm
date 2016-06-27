package aws

import (
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type LaunchConfigs []LaunchConfig

type LaunchConfig struct {
	Name           string
	ImageId        string
	InstanceType   string
	KeyName        string
	SecurityGroups string
	CreationTime   string
	Region         string
}

func GetLaunchConfigurations() (*LaunchConfigs, []error) {
	var wg sync.WaitGroup
	var errs []error

	lcList := new(LaunchConfigs)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionLaunchConfigurations(region.RegionName, lcList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering launch config list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return lcList, errs
}

func GetRegionLaunchConfigurations(region *string, lcList *LaunchConfigs) error {
	svc := autoscaling.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})

	if err != nil {
		return err
	}

	lc := make(LaunchConfigs, len(result.LaunchConfigurations))
	for i, config := range result.LaunchConfigurations {

		lc[i] = LaunchConfig{
			Name:         *config.LaunchConfigurationName,
			ImageId:      *config.ImageId,
			InstanceType: *config.InstanceType,
			KeyName:      *config.KeyName,
			//SecurityGroups: *config.SecurityGroups,
			CreationTime: config.CreatedTime.String(),
			Region:       fmt.Sprintf(*region),
		}
	}
	*lcList = append(*lcList, lc[:]...)

	return nil
}

func (i *LaunchConfigs) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.ImageId,
			val.InstanceType,
			val.KeyName,
			val.SecurityGroups,
			val.CreationTime,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Image Id", "Instance Type", "Key Name", "Security Groups", "Creation Time", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
