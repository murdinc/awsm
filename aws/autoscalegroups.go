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
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type AutoScaleGroups []AutoScaleGroup

type AutoScaleGroup struct {
	Name              string
	Class             string
	HealthCheck       string
	LaunchConfig      string
	LoadBalancers     string
	Instances         string
	DesiredCapacity   string
	MinSize           string
	MaxSize           string
	Cooldown          string
	AvailabilityZones string
	VpcName           string
	VpcId             string
	SubnetName        string
	SubnetId          string
	Region            string
}

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
	a.HealthCheck = aws.StringValue(autoscalegroup.HealthCheckType)
	a.LaunchConfig = aws.StringValue(autoscalegroup.LaunchConfigurationName)
	a.LoadBalancers = strings.Join(aws.StringValueSlice(autoscalegroup.LoadBalancerNames), ", ")
	a.Instances = fmt.Sprint(len(autoscalegroup.Instances))
	a.DesiredCapacity = fmt.Sprint(aws.Int64Value(autoscalegroup.DesiredCapacity))
	a.MinSize = fmt.Sprint(aws.Int64Value(autoscalegroup.MinSize))
	a.MaxSize = fmt.Sprint(aws.Int64Value(autoscalegroup.MaxSize))
	a.Cooldown = fmt.Sprint(aws.Int64Value(autoscalegroup.DefaultCooldown))
	a.AvailabilityZones = strings.Join(aws.StringValueSlice(autoscalegroup.AvailabilityZones), ", ")
	a.SubnetId = aws.StringValue(autoscalegroup.VPCZoneIdentifier)
	a.SubnetName = subList.GetSubnetName(a.SubnetId)
	a.VpcId = subList.GetVpcIdBySubnetId(a.SubnetId)
	a.VpcName = subList.GetVpcNameBySubnetId(a.SubnetId)
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
	var cfg config.AutoScaleGroupClassConfig
	err = cfg.LoadConfig(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Autoscaling group class configuration for [" + class + "]")
	}

	// Verify the launchconfig class input
	var launchConfigurationCfg config.LaunchConfigurationClassConfig
	err = launchConfigurationCfg.LoadConfig(cfg.LaunchConfigurationClass)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Launch Configuration class configuration for [" + cfg.LaunchConfigurationClass + "]")
	}

	// Get the AZs
	azs, errs := GetAZs()
	if errs != nil {
		return errors.New("Error gathering region list")
	}

	//for region, regionAZs := range azs.GetRegionMap(cfg.AvailabilityZones) {
	for region, regionAZs := range azs.GetRegionMap(cfg.AvailabilityZones) {

		// Verify that the latest Launch Configuration is available in this region
		lcName := GetLaunchConfigurationName(region, cfg.LaunchConfigurationClass, launchConfigurationCfg.Version)
		if lcName == "" {
			return fmt.Errorf("Launch Configuration [%s] version [%d] is not available in [%s]!", cfg.LaunchConfigurationClass, launchConfigurationCfg.Version, region)
		} else {
			terminal.Information(fmt.Sprintf("Found latest Launch Configuration [%s] version [%d] in [%s]", cfg.LaunchConfigurationClass, launchConfigurationCfg.Version, region))
		}

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
		var vpcs []string

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
			} else {
				terminal.Information("Found Availability Zone [" + az + "]!")
			}

			params.AvailabilityZones = append(params.AvailabilityZones, aws.String(az))

			for _, sub := range *subList {
				if sub.Class == cfg.SubnetClass && sub.AvailabilityZone == az {
					vpcs = append(vpcs, sub.SubnetId)
				}
			}

		}

		// Set the VPCZoneIdentifier (SubnetIds seperated by comma)
		params.VPCZoneIdentifier = aws.String(strings.Join(vpcs, ", "))

		// Set the Load Balancers
		for _, elb := range cfg.LoadBalancerNames {
			params.LoadBalancerNames = append(params.LoadBalancerNames, aws.String(elb))
		}

		// Set the Termination Policies
		for _, terminationPolicy := range cfg.LoadBalancerNames {
			params.TerminationPolicies = append(params.TerminationPolicies, aws.String(terminationPolicy)) // ??
		}

		fmt.Println(params)

		// Create it!
		if !dryRun {
			resp, err := svc.CreateAutoScalingGroup(params)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					return errors.New(awsErr.Message())
				}
				return err
			}

			fmt.Println(resp)

			terminal.Information("Done!")
		}

	}

	return nil

}

// Public function with confirmation terminal prompt
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
	if !dryRun {
		err = updateAutoScaleGroups(asgList, version, double)
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

// Private function without the confirmation terminal prompts
func updateAutoScaleGroups(asgList *AutoScaleGroups, version string, double bool) (err error) {
	for _, asg := range *asgList {
		svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(asg.Region)}))

		params := &autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(asg.Name),
			AvailabilityZones: []*string{
				aws.String("XmlStringMaxLen255"), // Required
				// More values...
			},
			//DefaultCooldown:         aws.Int64(int64(asg.Cooldown)),
			//DesiredCapacity:         aws.Int64(int64(asg.DesiredCapacity)),
			HealthCheckGracePeriod:  aws.Int64(1),
			HealthCheckType:         aws.String("XmlStringMaxLen32"),
			LaunchConfigurationName: aws.String("ResourceName"),
			MaxSize:                 aws.Int64(1),
			MinSize:                 aws.Int64(1),
			NewInstancesProtectedFromScaleIn: aws.Bool(true),
			PlacementGroup:                   aws.String("XmlStringMaxLen255"),
			TerminationPolicies: []*string{
				aws.String("XmlStringMaxLen1600"), // Required
				// More values...
			},
			VPCZoneIdentifier: aws.String("XmlStringMaxLen255"),
		}

		/*

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

		*/

		_, err := svc.UpdateAutoScalingGroup(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

		terminal.Information("Deleted AutoScaling Group [" + asg.Name + "] in [" + asg.Region + "]!")
	}

	return nil
}

// Public function with confirmation terminal prompt
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
	if !dryRun {
		err = deleteAutoScaleGroups(asgList, force)
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

// Private function without the confirmation terminal prompts
func deleteAutoScaleGroups(asgList *AutoScaleGroups, force bool) (err error) {
	for _, asg := range *asgList {
		svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(asg.Region)}))

		params := &autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(asg.Name),
			ForceDelete:          aws.Bool(force),
		}

		_, err := svc.DeleteAutoScalingGroup(params)
		if err != nil {
			return err
		}

		terminal.Information("Deleted AutoScaling Group [" + asg.Name + "] in [" + asg.Region + "]!")
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

func (i *AutoScaleGroups) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.Class,
			val.HealthCheck,
			val.LaunchConfig,
			val.LoadBalancers,
			val.Instances,
			val.DesiredCapacity,
			val.MinSize,
			val.MaxSize,
			val.Cooldown,
			val.AvailabilityZones,
			val.VpcName,
			val.SubnetName,
		}
	}

	table.SetHeader([]string{"Name", "Class", "Health Check", "Launch Config", "Load Balancers", "Instances", "Desired Capacity", "Min Size", "Max Size", "Cooldown", "Availability Zones", "VPC", "Subnet"})

	table.AppendBulk(rows)
	table.Render()
}
