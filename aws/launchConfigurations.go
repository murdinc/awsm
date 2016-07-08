package aws

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/dustin/go-humanize"
	"github.com/murdinc/awsm/config"
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

func GetLaunchConfigurations(search string) (*LaunchConfigs, []error) {
	var wg sync.WaitGroup
	var errs []error

	lcList := new(LaunchConfigs)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionLaunchConfigurations(*region.RegionName, lcList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering launch config list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return lcList, errs
}

func GetRegionLaunchConfigurations(region string, lcList *LaunchConfigs, search string) error {
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

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, c := range lc {
			rLc := reflect.ValueOf(c)

			for k := 0; k < rLc.NumField(); k++ {
				sVal := rLc.Field(k).String()

				if term.MatchString(sVal) {
					*lcList = append(*lcList, lc[i])
					continue Loop
				}
			}
		}
	} else {
		*lcList = append(*lcList, lc[:]...)
	}

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

func CreateLaunchConfigurations(class string, dryRun bool) (err error) {

	// Verify the launch config class input
	var launchConfigurationCfg config.LaunchConfigurationClassConfig
	err = launchConfigurationCfg.LoadConfig(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Launch Configuration class configuration for [" + class + "]")
	}

	// Instance Class Config
	var instanceCfg config.InstanceClassConfig
	err = instanceCfg.LoadConfig(launchConfigurationCfg.InstanceClass)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Instance class configuration for [" + launchConfigurationCfg.InstanceClass + "]")
	}

	params := &autoscaling.CreateLaunchConfigurationInput{
		LaunchConfigurationName:  aws.String(fmt.Sprintf("%s-v%d", class, launchConfigurationCfg.Version)),
		AssociatePublicIpAddress: aws.Bool(instanceCfg.PublicIpAddress),
		InstanceMonitoring: &autoscaling.InstanceMonitoring{
			Enabled: aws.Bool(instanceCfg.Monitoring),
		},
		InstanceType: aws.String(instanceCfg.InstanceType),
		UserData:     aws.String(instanceCfg.UserData),
		//KernelId:         aws.String("XmlStringMaxLen255"),
		//PlacementTenancy: aws.String("XmlStringMaxLen64"),
		//RamdiskId:        aws.String("XmlStringMaxLen255"),
		//SpotPrice: aws.String("SpotPrice"),
		//ClassicLinkVPCId:         aws.String("XmlStringMaxLen255"),
		//ClassicLinkVPCSecurityGroups: []*string{
		//aws.String("XmlStringMaxLen255"),
		//},
	}

	// IAM Profile
	if len(instanceCfg.IAMUser) > 0 {
		iam, err := GetIAMUser(instanceCfg.IAMUser)
		if err != nil {
			return err
		} else {
			terminal.Information("Found IAM User [" + iam.UserName + "]")
			params.IamInstanceProfile = aws.String(iam.Arn)
		}
	}

	// Increment the version
	terminal.Information(fmt.Sprintf("Previous version of launch configuration is [%d]", launchConfigurationCfg.Version))
	launchConfigurationCfg.Increment(class)
	terminal.Information(fmt.Sprintf("New version of launch configuration is [%d]", launchConfigurationCfg.Version))

	// Get the AZ list
	azs, _ := GetAZs()

	for _, az := range launchConfigurationCfg.AvailabilityZones {
		region := azs.GetRegion(az)

		// EBS
		ebsVolumes := make([]*autoscaling.BlockDeviceMapping, len(instanceCfg.EBSVolumes))
		for i, ebsClass := range instanceCfg.EBSVolumes {
			var volCfg config.VolumeClassConfig
			err := volCfg.LoadConfig(ebsClass)
			if err != nil {
				return err
			} else {
				terminal.Information("Found Volume Class Configuration for [" + ebsClass + "]")
			}

			latestSnapshot, err := GetLatestSnapshotByTag(region, "Class", volCfg.Snapshot)
			if err != nil {
				return err
			} else {
				terminal.Information("Found Snapshot [" + latestSnapshot.SnapshotId + "] with class [" + latestSnapshot.Class + "] created [" + latestSnapshot.CreatedHuman + "]")
			}

			ebsVolumes[i] = &autoscaling.BlockDeviceMapping{
				DeviceName: aws.String(volCfg.DeviceName),
				Ebs: &autoscaling.Ebs{
					DeleteOnTermination: aws.Bool(volCfg.DeleteOnTermination),
					SnapshotId:          aws.String(latestSnapshot.SnapshotId),
					VolumeSize:          aws.Int64(int64(volCfg.VolumeSize)),
					VolumeType:          aws.String(volCfg.VolumeType),
					//Encrypted:           aws.Bool(volCfg.Encrypted),
				},
				//NoDevice:    aws.String("String"),
				//VirtualName: aws.String("String"),
			}

			if volCfg.VolumeType == "io1" {
				ebsVolumes[i].Ebs.Iops = aws.Int64(int64(volCfg.Iops))
			}

		}

		// EBS Optimized
		if instanceCfg.EbsOptimized {
			terminal.Information("Launching as EBS Optimized")
			params.EbsOptimized = aws.Bool(instanceCfg.EbsOptimized)
		}

		params.BlockDeviceMappings = ebsVolumes

		// AMI
		ami, err := GetLatestImageByTag(region, "Class", instanceCfg.AMI)
		if err != nil {
			return err
		} else {
			terminal.Information("Found AMI [" + ami.ImageId + "] with class [" + ami.Class + "] created [" + ami.CreatedHuman + "]")
		}

		params.ImageId = aws.String(ami.ImageId)

		// KeyPair
		keyPair, err := GetKeyPairByName(region, instanceCfg.KeyName)
		if err != nil {
			return err
		} else {
			terminal.Information("Found KeyPair [" + keyPair.KeyName + "] in [" + keyPair.Region + "]")
		}

		params.KeyName = aws.String(keyPair.KeyName)

		// VPC / Subnet
		var vpc Vpc
		var subnet Subnet
		//var subnetId string
		secGroupIds := make([]*string, len(instanceCfg.SecurityGroups))
		if instanceCfg.Vpc != "" && instanceCfg.Subnet != "" {
			// VPC
			vpc, err = GetVpcByTag(region, "Class", instanceCfg.Vpc)
			if err != nil {
				return err
			} else {
				terminal.Information("Found VPC [" + vpc.VpcId + "] in Region [" + region + "]")
			}

			// Subnet
			subnet, err = vpc.GetVpcSubnetByTag("Class", instanceCfg.Subnet)
			if err != nil {
				return err
			} else {
				//subnetId = subnet.SubnetId
				terminal.Information("Found Subnet [" + subnet.SubnetId + "] in VPC [" + subnet.VpcId + "]")
			}

			// VPC Security Groups
			secGroups, err := vpc.GetVpcSecurityGroupByTagMulti("Class", instanceCfg.SecurityGroups)
			if err != nil {
				return err
			} else {
				for i, secGroup := range secGroups {
					terminal.Information("Found VPC Security Group [" + secGroup.GroupId + "] with name [" + secGroup.Name + "]")
					secGroupIds[i] = aws.String(secGroup.GroupId)
				}
			}
		} else {
			terminal.Information("No VPC and/or Subnet specified for instance Class [" + class + "]")

			// EC2-Classic security groups
			secGroups, err := GetSecurityGroupByTagMulti(region, "Class", instanceCfg.SecurityGroups)
			if err != nil {
				return err
			} else {
				for i, secGroup := range secGroups {
					terminal.Information("Found Security Group [" + secGroup.GroupId + "] with name [" + secGroup.Name + "]")
					secGroupIds[i] = aws.String(secGroup.GroupId)
				}
			}
		}

		params.SecurityGroups = secGroupIds

		svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(region)}))

		_, err = svc.CreateLaunchConfiguration(params)

		if err != nil {
			//fmt.Println(err)
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

	}

	return nil
}

func DeleteLaunchConfigurations(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	lcList := new(LaunchConfigs)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionLaunchConfigurations(region, lcList, search)
	} else {
		lcList, _ = GetLaunchConfigurations(search)
	}

	if err != nil {
		return errors.New("Error gathering Launch Configuration list")
	}

	if len(*lcList) > 0 {
		// Print the table
		lcList.PrintTable()
	} else {
		return errors.New("No Launch Configurations found!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Launch Configurations") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = deleteLaunchConfigurations(lcList, dryRun)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func deleteLaunchConfigurations(lcList *LaunchConfigs, dryRun bool) (err error) {
	for _, lc := range *lcList {
		svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(lc.Region)}))

		params := &autoscaling.DeleteLaunchConfigurationInput{
			LaunchConfigurationName: aws.String(lc.Name),
		}

		if !dryRun {
			_, err := svc.DeleteLaunchConfiguration(params)
			if err != nil {
				return err
			}

			terminal.Information("Deleted Launch Configuration [" + lc.Name + "] in [" + lc.Region + "]")
		}
	}

	return nil
}

func (i *LaunchConfigs) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Launch Configurations Found!")
		return
	}

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
