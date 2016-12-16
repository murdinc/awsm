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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/dustin/go-humanize"
	"github.com/hashicorp/hil"
	"github.com/hashicorp/hil/ast"
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

//LaunchConfigs represents a slice of Launch Configurations
type LaunchConfigs []LaunchConfig

// LaunchConfig represents a single Launch Configuration
type LaunchConfig models.LaunchConfig

// GetLaunchConfigurationName returns the name of a Launch Configuration that matches the provided class and version, while also verifying that it exists
func GetLaunchConfigurationName(region, class string, version int) string {
	name := fmt.Sprintf("%s-v%d", class, version)
	_, err := GetLaunchConfigurationsByName(region, name)
	if err != nil {
		return ""
	}

	return name
}

// GetLaunchConfigurationsByName returns a slice of Launch Configurations for a given region and name
func GetLaunchConfigurationsByName(region, name string) (LaunchConfigs, error) {

	svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(region)}))
	params := &autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{
			aws.String(name),
		},
	}
	result, err := svc.DescribeLaunchConfigurations(params)
	if err != nil || len(result.LaunchConfigurations) == 0 {
		return LaunchConfigs{}, err
	}

	secGrpList := new(SecurityGroups)
	err = GetRegionSecurityGroups(region, secGrpList, "")

	imgList := new(Images)
	GetRegionImages(region, imgList, "", false)

	lcList := make(LaunchConfigs, len(result.LaunchConfigurations))
	for i, config := range result.LaunchConfigurations {
		lcList[i].Marshal(config, region, secGrpList, imgList)
	}

	if len(lcList) == 0 {
		return lcList, errors.New("No Launch Configurations found with name of [" + name + "] in [" + region + "].")
	}

	return lcList, err
}

// GetLaunchConfigurations returns a slice of Launch Configurations that match the provided search term
func GetLaunchConfigurations(search string) (*LaunchConfigs, []error) {
	var wg sync.WaitGroup
	var errs []error

	lcList := new(LaunchConfigs)
	regions := regions.GetRegionList()

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

// GetRegionLaunchConfigurations returns a slice of Launch Configurations into the provided LaunchConfigs slice that match the region and search term
func GetRegionLaunchConfigurations(region string, lcList *LaunchConfigs, search string) error {
	svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	if err != nil || len(result.LaunchConfigurations) == 0 {
		return err
	}

	secGrpList := new(SecurityGroups)
	err = GetRegionSecurityGroups(region, secGrpList, "")

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

// Marshal parses the response from the aws sdk into an awsm LaunchConfig
func (l *LaunchConfig) Marshal(config *autoscaling.LaunchConfiguration, region string, secGrpList *SecurityGroups, imgList *Images) {
	secGroupNames := secGrpList.GetSecurityGroupNames(aws.StringValueSlice(config.SecurityGroups))
	secGroupNamesSorted := sort.StringSlice(secGroupNames[0:])
	secGroupNamesSorted.Sort()

	l.Name = aws.StringValue(config.LaunchConfigurationName)
	l.ImageID = aws.StringValue(config.ImageId)
	l.ImageName = imgList.GetImageName(l.ImageID)
	l.InstanceType = aws.StringValue(config.InstanceType)
	l.KeyName = aws.StringValue(config.KeyName)
	l.CreationTime = aws.TimeValue(config.CreatedTime) // robots
	l.CreatedHuman = humanize.Time(l.CreationTime)     // humans
	l.EbsOptimized = aws.BoolValue(config.EbsOptimized)
	l.SecurityGroups = strings.Join(secGroupNamesSorted, ", ")
	l.Region = region

	for _, snapshot := range config.BlockDeviceMappings {
		l.SnapshotIDs = append(l.SnapshotIDs, *snapshot.Ebs.SnapshotId)
	}
}

// LockedSnapshotIds returns a map of locked EBS Snapshots (that are currently being used in Launch Configurations)
func (l *LaunchConfigs) LockedSnapshotIds() map[string]bool {
	ids := make(map[string]bool)
	for _, config := range *l {
		for _, snap := range config.SnapshotIDs {
			ids[snap] = true
		}
	}
	return ids
}

// LockedImageIds returns a list of locked AMI's (that are currently being used in Launch Configurations)
func (l *LaunchConfigs) LockedImageIds() (ids map[string]bool) {
	for _, config := range *l {
		ids[config.ImageID] = true
	}
	return ids
}

// CreateLaunchConfigurations creates a new Launch Configuration of a given class
func CreateLaunchConfigurations(class string, dryRun bool) (err error) {

	// Verify the launch config class input
	cfg, err := config.LoadLaunchConfigurationClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found Launch Configuration class configuration for [" + class + "]")

	// Instance Class Config
	instanceCfg, err := config.LoadInstanceClass(cfg.InstanceClass)
	if err != nil {
		return err
	}

	terminal.Information("Found Instance class configuration for [" + cfg.InstanceClass + "]")

	// Increment the version
	terminal.Information(fmt.Sprintf("Previous version of launch configuration is [%d]", cfg.Version))
	cfg.Increment(class)
	terminal.Delta(fmt.Sprintf("New version of launch configuration is [%d]", cfg.Version))

	params := &autoscaling.CreateLaunchConfigurationInput{
		LaunchConfigurationName:  aws.String(fmt.Sprintf("%s-v%d", class, cfg.Version)),
		AssociatePublicIpAddress: aws.Bool(instanceCfg.PublicIPAddress),
		InstanceMonitoring: &autoscaling.InstanceMonitoring{
			Enabled: aws.Bool(instanceCfg.Monitoring),
		},
		InstanceType: aws.String(instanceCfg.InstanceType),
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
		}

		terminal.Information("Found IAM User [" + iam.UserName + "]")
		params.IamInstanceProfile = aws.String(iam.Arn)

	}

	for _, region := range cfg.Regions {

		if !regions.ValidRegion(region) {
			return errors.New("Region [" + region + "] is not valid!")
		}

		// EBS
		ebsVolumes := make([]*autoscaling.BlockDeviceMapping, len(instanceCfg.EBSVolumes))
		for i, ebsClass := range instanceCfg.EBSVolumes {
			volCfg, err := config.LoadVolumeClass(ebsClass)
			if err != nil {
				return err
			}

			terminal.Information("Found Volume Class Configuration for [" + ebsClass + "]")

			latestSnapshot, err := GetLatestSnapshotByTag(region, "Class", volCfg.Snapshot)
			if err != nil {
				return err
			}

			terminal.Information("Found Snapshot [" + latestSnapshot.SnapshotID + "] with class [" + latestSnapshot.Class + "] created [" + latestSnapshot.CreatedHuman + "]")

			ebsVolumes[i] = &autoscaling.BlockDeviceMapping{
				DeviceName: aws.String(volCfg.DeviceName),
				Ebs: &autoscaling.Ebs{
					DeleteOnTermination: aws.Bool(volCfg.DeleteOnTermination),
					SnapshotId:          aws.String(latestSnapshot.SnapshotID),
					VolumeSize:          aws.Int64(int64(volCfg.VolumeSize)),
					VolumeType:          aws.String(volCfg.VolumeType),
					Encrypted:           aws.Bool(volCfg.Encrypted),
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
		}

		terminal.Information("Found AMI [" + ami.ImageID + "] with class [" + ami.Class + "] created [" + ami.CreatedHuman + "]")
		params.ImageId = aws.String(ami.ImageID)

		// KeyPair
		keyPair, err := GetKeyPairByName(region, instanceCfg.KeyName)
		if err != nil {
			return err
		}

		terminal.Information("Found KeyPair [" + keyPair.KeyName + "] in [" + keyPair.Region + "]")
		params.KeyName = aws.String(keyPair.KeyName)

		// VPC / Subnet
		var vpc Vpc
		var subnet Subnet
		secGroupIds := make([]*string, len(instanceCfg.SecurityGroups))
		if instanceCfg.Vpc != "" && instanceCfg.Subnet != "" {
			// VPC
			vpc, err = GetVpcByTag(region, "Class", instanceCfg.Vpc)
			if err != nil {
				return err
			}

			terminal.Information("Found VPC [" + vpc.VpcID + "] in Region [" + region + "]")

			// Subnet
			subnet, err = vpc.GetVpcSubnetByTag("Class", instanceCfg.Subnet)
			if err != nil {
				return err
			}

			terminal.Information("Found Subnet [" + subnet.SubnetID + "] in VPC [" + subnet.VpcID + "]")

			// VPC Security Groups
			secGroups, err := vpc.GetVpcSecurityGroupByTagMulti("Class", instanceCfg.SecurityGroups)
			if err != nil {
				return err
			}

			for i, secGroup := range secGroups {
				terminal.Information("Found VPC Security Group [" + secGroup.GroupID + "] with name [" + secGroup.Name + "]")
				secGroupIds[i] = aws.String(secGroup.GroupID)
			}

		} else {
			terminal.Notice("No VPC and/or Subnet specified for instance Class [" + class + "]")

			// EC2-Classic security groups
			secGroups, err := GetSecurityGroupByTagMulti(region, "Class", instanceCfg.SecurityGroups)
			if err != nil {
				return err
			}

			for i, secGroup := range secGroups {
				terminal.Information("Found Security Group [" + secGroup.GroupID + "] with name [" + secGroup.Name + "]")
				secGroupIds[i] = aws.String(secGroup.GroupID)
			}

		}

		// Parse Userdata
		tree, err := hil.Parse(instanceCfg.UserData)
		if err != nil {
			return err
		}

		config := &hil.EvalConfig{
			GlobalScope: &ast.BasicScope{
				VarMap: map[string]ast.Variable{
					"var.class": ast.Variable{
						Type:  ast.TypeString,
						Value: class,
					},
					"var.sequence": ast.Variable{
						Type:  ast.TypeString,
						Value: cfg.Version,
					},
					"var.locale": ast.Variable{
						Type:  ast.TypeString,
						Value: region,
					},
				},
			},
		}

		result, err := hil.Eval(tree, config)
		if err != nil {
			return err
		}

		parsedUserData := result.Value.(string)

		if dryRun {
			terminal.Information("User Data:")
			terminal.Information(parsedUserData)
		}

		params.UserData = aws.String(parsedUserData)
		params.SecurityGroups = secGroupIds

		svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(region)}))

		_, err = svc.CreateLaunchConfiguration(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

	}

	// Rotate out older launch configurations
	if cfg.Retain > 1 {
		err := RotateLaunchConfigurations(class, cfg, dryRun)
		if err != nil {
			terminal.ShowErrorMessage(fmt.Sprintf("Error rotating [%s] launch configurations!", class), err.Error())
			return err
		}
	}

	return nil
}

// RotateLaunchConfigurations rotates out older Launch Configurations
func RotateLaunchConfigurations(class string, cfg config.LaunchConfigurationClass, dryRun bool) error {
	var wg sync.WaitGroup
	var errs []error

	autoScaleGroups, err := GetAutoScaleGroups("")
	if err != nil {
		return errors.New("Error while retrieving the list of launch configurations to exclude from rotation!")
	}
	excludedConfigs := autoScaleGroups.LockedLaunchConfigurations()

	regions := regions.GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()

			// Get all the launch configs of this class in this region
			launchConfigs, err := GetLaunchConfigurationsByName(*region.RegionName, class)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering launch configuration list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}

			// Exclude the launch configs being used in Autoscale Groups
			for i, lc := range launchConfigs {
				if excludedConfigs[lc.Name] {
					terminal.Notice("Launch Configuration [" + lc.Name + ") ] is being used in an autoscale group, skipping!")
					launchConfigs = append(launchConfigs[:i], launchConfigs[i+1:]...)
				}
			}

			// Delete the oldest ones if we have more than the retention number
			if len(launchConfigs) > cfg.Retain {
				sort.Sort(launchConfigs) // important!
				ds := launchConfigs[cfg.Retain:]
				deleteLaunchConfigurations(&ds, dryRun)
			}

		}(region)
	}
	wg.Wait()

	if errs != nil {
		return errors.New("Error rotating snapshots for [" + class + "]!")
	}

	return nil
}

// Len returns the current number of Launch Configurations in the slice
func (l LaunchConfigs) Len() int {
	return len(l)
}

// Swap swaps the position of two Launch Configurations in the slice
func (l LaunchConfigs) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// Less returns true of the Launch Configuration at indes i was created after the Launch Configuration at index j
func (l LaunchConfigs) Less(i, j int) bool {
	return l[i].CreationTime.After(l[j].CreationTime)
}

// DeleteLaunchConfigurations deletes one or more Launch Configurations that match the provided search term and optional region
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
	if !terminal.PromptBool("Are you sure you want to delete these Launch Configurations?") {
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

			terminal.Delta("Deleted Launch Configuration [" + lc.Name + "] in [" + lc.Region + "]")
		}
	}

	return nil
}

// PrintTable Prints an ascii table of the list of Launch Configurations
func (l *LaunchConfigs) PrintTable() {
	if len(*l) == 0 {
		terminal.ShowErrorMessage("Warning", "No Launch Configurations Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*l))

	for index, lc := range *l {
		models.ExtractAwsmTable(index, lc, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
