package aws

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type AutoScaleGroups []AutoScaleGroup

type AutoScaleGroup models.AutoScaleGroup

func GetAutoScaleGroups(search string) (*AutoScaleGroups, []error) {
	var wg sync.WaitGroup
	var errs []error

	asgList := new(AutoScaleGroups)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionAutoScaleGroups(*region.RegionName, asgList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering autoscale group list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return asgList, errs
}

func GetRegionAutoScaleGroups(region string, asgList *AutoScaleGroups, search string) error {
	svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})

	if err != nil {
		return err
	}

	subList := new(Subnets)
	GetRegionSubnets(region, subList, "")

	asg := make(AutoScaleGroups, len(result.AutoScalingGroups))
	for i, autoscalegroup := range result.AutoScalingGroups {
		asg[i].Marshal(autoscalegroup, region, subList)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, g := range asg {
			rAsg := reflect.ValueOf(g)

			for k := 0; k < rAsg.NumField(); k++ {
				sVal := rAsg.Field(k).String()

				if term.MatchString(sVal) {
					*asgList = append(*asgList, asg[i])
					continue Loop
				}
			}
		}
	} else {
		*asgList = append(*asgList, asg[:]...)
	}

	return nil
}

func (a *AutoScaleGroup) Marshal(autoscalegroup *autoscaling.Group, region string, subList *Subnets) {
	a.Name = aws.StringValue(autoscalegroup.AutoScalingGroupName)
	a.Class = GetTagValue("Class", autoscalegroup.Tags)
	a.HealthCheckType = aws.StringValue(autoscalegroup.HealthCheckType)
	a.HealthCheckGracePeriod = int(aws.Int64Value(autoscalegroup.HealthCheckGracePeriod))
	a.LaunchConfig = aws.StringValue(autoscalegroup.LaunchConfigurationName)
	a.LoadBalancers = strings.Join(aws.StringValueSlice(autoscalegroup.LoadBalancerNames), ", ")
	a.InstanceCount = len(autoscalegroup.Instances)
	a.DesiredCapacity = int(aws.Int64Value(autoscalegroup.DesiredCapacity))
	a.MinSize = int(aws.Int64Value(autoscalegroup.MinSize))
	a.MaxSize = int(aws.Int64Value(autoscalegroup.MaxSize))
	a.DefaultCooldown = int(aws.Int64Value(autoscalegroup.DefaultCooldown))
	a.AvailabilityZones = strings.Join(aws.StringValueSlice(autoscalegroup.AvailabilityZones), ", ")
	a.SubnetID = aws.StringValue(autoscalegroup.VPCZoneIdentifier)
	a.SubnetName = subList.GetSubnetName(a.SubnetID)
	a.VpcID = subList.GetVpcIDBySubnetID(a.SubnetID)
	a.VpcName = subList.GetVpcNameBySubnetID(a.SubnetID)
	a.Region = region
}

func (a *AutoScaleGroups) LockedLaunchConfigurations() map[string]bool {

	ids := make(map[string]bool, len(*a))
	for _, asg := range *a {
		ids[asg.LaunchConfig] = true
	}
	return ids
}

func CreateAutoScaleGroups(class string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Verify the asg config class input
	cfg, err := config.LoadAutoscalingGroupClass(class)
	if err != nil {
		return err
	}
	terminal.Information("Found Autoscaling group class configuration for [" + class + "]")

	// Verify the launchconfig class input
	launchConfigurationCfg, err := config.LoadLaunchConfigurationClass(cfg.LaunchConfigurationClass)
	if err != nil {
		return err
	}
	terminal.Information("Found Launch Configuration class configuration for [" + cfg.LaunchConfigurationClass + "]")

	// Get the AZs
	azs, errs := GetAZs()
	if errs != nil {
		return errors.New("Error gathering region list")
	}

	for region, regionAZs := range azs.GetRegionMap(cfg.AvailabilityZones) {

		// Verify that the latest Launch Configuration is available in this region
		lcName := GetLaunchConfigurationName(region, cfg.LaunchConfigurationClass, launchConfigurationCfg.Version)
		if lcName == "" {
			return fmt.Errorf("Launch Configuration [%s] version [%d] is not available in [%s]!", cfg.LaunchConfigurationClass, launchConfigurationCfg.Version, region)
		}
		terminal.Information(fmt.Sprintf("Found latest Launch Configuration [%s] version [%d] in [%s]", cfg.LaunchConfigurationClass, launchConfigurationCfg.Version, region))

		svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(region)}))

		params := &autoscaling.CreateAutoScalingGroupInput{
			AutoScalingGroupName:    aws.String(class),
			MaxSize:                 aws.Int64(int64(cfg.MaxSize)),
			MinSize:                 aws.Int64(int64(cfg.MinSize)),
			DefaultCooldown:         aws.Int64(int64(cfg.DefaultCooldown)),
			DesiredCapacity:         aws.Int64(int64(cfg.DesiredCapacity)),
			HealthCheckGracePeriod:  aws.Int64(int64(cfg.HealthCheckGracePeriod)),
			HealthCheckType:         aws.String(cfg.HealthCheckType),
			LaunchConfigurationName: aws.String(lcName),
			// InstanceId:                       aws.String("XmlStringMaxLen19"),  // TODO ?
			// NewInstancesProtectedFromScaleIn: aws.Bool(true),                   // TODO ?
			// PlacementGroup:                   aws.String("XmlStringMaxLen255"), // TODO ?
			Tags: []*autoscaling.Tag{
				{
					// Name
					Key:               aws.String("Name"),
					PropagateAtLaunch: aws.Bool(true),
					ResourceId:        aws.String(class),
					ResourceType:      aws.String("auto-scaling-group"),
					Value:             aws.String(lcName),
				},
				{
					// Class
					Key:               aws.String("Class"),
					PropagateAtLaunch: aws.Bool(true),
					ResourceId:        aws.String(class),
					ResourceType:      aws.String("auto-scaling-group"),
					Value:             aws.String(class),
				},
			},
		}

		subList := new(Subnets)
		var vpcZones []string

		if cfg.SubnetClass != "" {
			err := GetRegionSubnets(region, subList, "")
			if err != nil {
				return err
			}
		}

		// Set the AZs
		for _, az := range regionAZs {
			if !azs.ValidAZ(az) {
				return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
			}
			terminal.Information("Found Availability Zone [" + az + "]!")

			params.AvailabilityZones = append(params.AvailabilityZones, aws.String(az))

			for _, sub := range *subList {
				if sub.Class == cfg.SubnetClass && sub.AvailabilityZone == az {
					vpcZones = append(vpcZones, sub.SubnetID)
				}
			}

		}

		// Set the VPCZoneIdentifier (SubnetIds seperated by comma)
		params.VPCZoneIdentifier = aws.String(strings.Join(vpcZones, ", "))

		// Set the Load Balancers
		for _, elb := range cfg.LoadBalancerNames {
			params.LoadBalancerNames = append(params.LoadBalancerNames, aws.String(elb))
		}

		// Set the Termination Policies
		for _, terminationPolicy := range cfg.LoadBalancerNames {
			params.TerminationPolicies = append(params.TerminationPolicies, aws.String(terminationPolicy)) // ??
		}

		// Create it!
		if !dryRun {
			_, err := svc.CreateAutoScalingGroup(params)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					return errors.New(awsErr.Message())
				}
				return err
			}

			terminal.Information("Created AutoScaling Group [" + aws.StringValue(params.AutoScalingGroupName) + "] in [" + region + "]!")

			terminal.Information("Done!")
		} else {
			fmt.Println(params)
		}

	}

	return nil

}

// UpdateAutoScaleGroups - Public function with confirmation terminal prompt
func UpdateAutoScaleGroups(name, version string, double, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	asgList, _ := GetAutoScaleGroups(name)

	if len(*asgList) > 0 {
		// Print the table
		asgList.PrintTable()
	} else {
		return errors.New("No AutoScaling Groups found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to update these AutoScaling Groups?") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = updateAutoScaleGroups(asgList, version, double, dryRun)
	if err == nil {
		terminal.Information("Done!")
	}

	return
}

// Private function without the confirmation terminal prompts
func updateAutoScaleGroups(asgList *AutoScaleGroups, version string, double, dryRun bool) (err error) {

	for _, asg := range *asgList {

		// Get the ASG class config
		cfg, err := config.LoadAutoscalingGroupClass(asg.Class)
		if err != nil {
			return err
		}

		terminal.Information("Found Autoscaling group class configuration for [" + asg.Class + "]")

		// Get the Launch Configuration class config
		launchConfigurationCfg, err := config.LoadLaunchConfigurationClass(cfg.LaunchConfigurationClass)
		if err != nil {
			return err
		}

		terminal.Information("Found Launch Configuration class configuration for [" + cfg.LaunchConfigurationClass + "]")

		// Get the AZs
		azs, errs := GetAZs()
		if errs != nil {
			return errors.New("Error gathering region list")
		}

		for region, regionAZs := range azs.GetRegionMap(cfg.AvailabilityZones) {

			// TODO check if exists yet ?

			// Verify that the latest Launch Configuration is available in this region
			lcName := GetLaunchConfigurationName(region, cfg.LaunchConfigurationClass, launchConfigurationCfg.Version)
			if lcName == "" {
				return fmt.Errorf("Launch Configuration [%s] version [%d] is not available in [%s]!", cfg.LaunchConfigurationClass, launchConfigurationCfg.Version, region)
			}
			terminal.Information(fmt.Sprintf("Found latest Launch Configuration [%s] version [%d] in [%s]", cfg.LaunchConfigurationClass, launchConfigurationCfg.Version, asg.Region))

			svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(region)}))

			params := &autoscaling.UpdateAutoScalingGroupInput{
				AutoScalingGroupName: aws.String(asg.Name),
				AvailabilityZones: []*string{
					aws.String("XmlStringMaxLen255"), // Required
					// More values...
				},
				DefaultCooldown:         aws.Int64(int64(cfg.DefaultCooldown)),
				DesiredCapacity:         aws.Int64(int64(cfg.DesiredCapacity)),
				HealthCheckGracePeriod:  aws.Int64(int64(cfg.HealthCheckGracePeriod)),
				HealthCheckType:         aws.String(cfg.HealthCheckType),
				LaunchConfigurationName: aws.String(lcName),
				MaxSize:                 aws.Int64(int64(asg.MaxSize)),
				MinSize:                 aws.Int64(int64(asg.MinSize)),
				//NewInstancesProtectedFromScaleIn: aws.Bool(true), // TODO?
				//PlacementGroup:                   aws.String("XmlStringMaxLen255"), // TODO
			}

			subList := new(Subnets)
			var vpcZones []string

			if cfg.SubnetClass != "" {
				err := GetRegionSubnets(region, subList, "")
				if err != nil {
					return err
				}
			}

			// Set the AZs
			for _, az := range regionAZs {
				if !azs.ValidAZ(az) {
					return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
				}
				terminal.Information("Found Availability Zone [" + az + "]!")

				params.AvailabilityZones = append(params.AvailabilityZones, aws.String(az))

				for _, sub := range *subList {
					if sub.Class == cfg.SubnetClass && sub.AvailabilityZone == az {
						vpcZones = append(vpcZones, sub.SubnetID)
					}
				}

			}

			// Set the VPCZoneIdentifier (SubnetIds seperated by comma)
			params.VPCZoneIdentifier = aws.String(strings.Join(vpcZones, ", "))

			// Set the Termination Policies
			for _, terminationPolicy := range cfg.LoadBalancerNames {
				params.TerminationPolicies = append(params.TerminationPolicies, aws.String(terminationPolicy)) // ??
			}

			// Update it!
			if !dryRun {
				_, err := svc.UpdateAutoScalingGroup(params)
				if err != nil {
					if awsErr, ok := err.(awserr.Error); ok {
						return errors.New(awsErr.Message())
					}
					return err
				}

				terminal.Information("Updated AutoScaling Group [" + asg.Name + "] in [" + region + "]!")

			} else {
				fmt.Println(params)
			}

		}

	}

	return nil
}

// DeleteAutoScaleGroups - Public function with confirmation terminal prompt
func DeleteAutoScaleGroups(name, region string, force, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	asgList := new(AutoScaleGroups)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionAutoScaleGroups(region, asgList, name)
	} else {
		asgList, _ = GetAutoScaleGroups(name)
	}

	if err != nil {
		return errors.New("Error gathering AutoScaling Groups list")
	}

	if len(*asgList) > 0 {
		// Print the table
		asgList.PrintTable()
	} else {
		return errors.New("No AutoScaling Groups found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these AutoScaling Groups?") {
		return errors.New("Aborting!")
	}

	// Delete 'Em

	err = deleteAutoScaleGroups(asgList, force, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func deleteAutoScaleGroups(asgList *AutoScaleGroups, force, dryRun bool) (err error) {
	for _, asg := range *asgList {
		svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(asg.Region)}))

		params := &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(asg.Name),
			ForceDelete:          aws.Bool(force),
		}

		// Delete it!
		if !dryRun {
			_, err := svc.DeleteAutoScalingGroup(params)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					return errors.New(awsErr.Message())
				}
				return err
			}

			terminal.Information("Deleted AutoScaling Group [" + asg.Name + "] in [" + asg.Region + "]!")

		} else {
			fmt.Println(params)
		}

	}

	return nil
}

func SuspendProcesses(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	asgList := new(AutoScaleGroups)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionAutoScaleGroups(region, asgList, search)
	} else {
		asgList, _ = GetAutoScaleGroups(search)
	}

	if err != nil {
		return errors.New("Error gathering Autoscale Group list")
	}

	if len(*asgList) > 0 {
		// Print the table
		asgList.PrintTable()
	} else {
		return errors.New("No Autoscale Groups found!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to suspend these Autoscale Groups?") {
		return errors.New("Aborting!")
	}

	// Suspend 'Em
	if !dryRun {
		err = suspendProcesses(asgList)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

		terminal.Information("Done!")
	}

	return nil
}

func suspendProcesses(asgList *AutoScaleGroups) error {
	for _, asg := range *asgList {
		svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(asg.Region)}))

		params := &autoscaling.ScalingProcessQuery{
			AutoScalingGroupName: aws.String(asg.Name),
		}
		_, err := svc.SuspendProcesses(params)

		if err != nil {
			return err
		}

		terminal.Information("Suspended processes on Autoscale Group [" + asg.Name + "] in [" + asg.Region + "]!")
	}

	return nil
}

func ResumeProcesses(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	asgList := new(AutoScaleGroups)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionAutoScaleGroups(region, asgList, search)
	} else {
		asgList, _ = GetAutoScaleGroups(search)
	}

	if err != nil {
		return errors.New("Error gathering Autoscale Group list")
	}

	if len(*asgList) > 0 {
		// Print the table
		asgList.PrintTable()
	} else {
		return errors.New("No Autoscale Groups found!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to resume these Autoscale Groups?") {
		return errors.New("Aborting!")
	}

	// Resume 'Em
	if !dryRun {
		err = resumeProcesses(asgList)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

		terminal.Information("Done!")
	}

	return
}

func resumeProcesses(asgList *AutoScaleGroups) error {
	for _, asg := range *asgList {
		svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(asg.Region)}))

		params := &autoscaling.ScalingProcessQuery{
			AutoScalingGroupName: aws.String(asg.Name),
		}
		_, err := svc.ResumeProcesses(params)

		if err != nil {
			return err
		}

		terminal.Information("Resumed processes on Autoscale Group [" + asg.Name + "] in [" + asg.Region + "]!")
	}

	return nil
}

func (a *AutoScaleGroups) PrintTable() {
	if len(*a) == 0 {
		terminal.ShowErrorMessage("Warning", "No Autoscale Groups Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*a))

	for index, asg := range *a {
		models.ExtractAwsmTable(index, asg, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
