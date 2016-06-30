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
	ImageName      string
	ImageId        string
	InstanceType   string
	KeyName        string
	SecurityGroups string
	CreationTime   time.Time
	CreatedHuman   string
	Region         string
	EbsOptimized   bool
	SnapshotIds    []string
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

	imgList := new(Images)
	GetRegionImages(region, imgList, "", false)

	lc := make(LaunchConfigs, len(result.LaunchConfigurations))
	for i, config := range result.LaunchConfigurations {
		lc[i].Marshal(config, region, secGrpList, imgList)
	}
	*lcList = append(*lcList, lc[:]...)

	return nil
}

func (l *LaunchConfig) Marshal(config *autoscaling.LaunchConfiguration, region string, secGrpList *SecurityGroups, imgList *Images) {
	secGroupNames := secGrpList.GetSecurityGroupNames(aws.StringValueSlice(config.SecurityGroups))
	secGroupNamesSorted := sort.StringSlice(secGroupNames[0:])
	secGroupNamesSorted.Sort()

	l.Name = aws.StringValue(config.LaunchConfigurationName)
	l.ImageId = aws.StringValue(config.ImageId)
	l.ImageName = imgList.GetImageName(l.ImageId)
	l.InstanceType = aws.StringValue(config.InstanceType)
	l.KeyName = aws.StringValue(config.KeyName)
	l.CreationTime = aws.TimeValue(config.CreatedTime) // robots
	l.CreatedHuman = humanize.Time(l.CreationTime)     // humans
	l.EbsOptimized = aws.BoolValue(config.EbsOptimized)
	l.SecurityGroups = strings.Join(secGroupNamesSorted, ", ")
	l.Region = region

	for _, snapshot := range config.BlockDeviceMappings {
		l.SnapshotIds = append(l.SnapshotIds, *snapshot.Ebs.SnapshotId)
	}
}

func (i *LaunchConfigs) LockedSnapshotIds() (ids map[string]bool) {
	for _, config := range *i {
		for _, snap := range config.SnapshotIds {
			ids[snap] = true
		}
	}
	return ids
}

func (i *LaunchConfigs) LockedImageIds() (ids map[string]bool) {
	for _, config := range *i {
		ids[config.ImageId] = true
	}
	return ids
}

func (i *LaunchConfigs) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.ImageName,
			val.ImageId,
			val.InstanceType,
			val.KeyName,
			val.SecurityGroups,
			val.CreatedHuman,
			fmt.Sprintf("%t", val.EbsOptimized),
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Image Name", "Image Id", "Instance Type", "Key Name", "Security Groups", "Created", "EBS Optimized", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
