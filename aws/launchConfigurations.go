package aws

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/dustin/go-humanize"
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
	CreationTime   time.Time
	CreatedHuman   string
	Region         string
	EbsOptimized   bool
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
			err := GetRegionLaunchConfigurations(*region.RegionName, lcList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering launch config list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return lcList, errs
}

func GetRegionLaunchConfigurations(region string, lcList *LaunchConfigs) error {
	svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	if err != nil {
		return err
	}

	secGrpList := new(SecurityGroups)
	err = GetRegionSecurityGroups(region, secGrpList)
	if err != nil {
		return err
	}

	lc := make(LaunchConfigs, len(result.LaunchConfigurations))
	for i, config := range result.LaunchConfigurations {
		lc[i].Marshal(config, region, secGrpList)
	}
	*lcList = append(*lcList, lc[:]...)

	return nil
}

func (l *LaunchConfig) Marshal(config *autoscaling.LaunchConfiguration, region string, secGrpList *SecurityGroups) {
	l.Name = aws.StringValue(config.LaunchConfigurationName)
	l.ImageId = aws.StringValue(config.ImageId)
	l.InstanceType = aws.StringValue(config.InstanceType)
	l.KeyName = aws.StringValue(config.KeyName)
	l.CreationTime = aws.TimeValue(config.CreatedTime) // robots
	l.CreatedHuman = humanize.Time(l.CreationTime)     // humans
	config.EbsOptimized = config.EbsOptimized
	l.Region = region

	secGroupNames := secGrpList.GetSecurityGroupNames(aws.StringValueSlice(config.SecurityGroups))

	secGroupNamesSorted := sort.StringSlice(secGroupNames[0:])
	secGroupNamesSorted.Sort()

	l.SecurityGroups = strings.Join(secGroupNamesSorted, ", ")
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
			val.CreatedHuman,
			fmt.Sprintf("%t", val.EbsOptimized),
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Image Id", "Instance Type", "Key Name", "Security Groups", "Created", "EBS Optimized", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
