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

func (a *AutoScaleGroups) LockedLaunchConfigurations() (ids map[string]bool) {
	for _, asg := range *a {
		ids[asg.LaunchConfig] = true
	}
	return ids
}

func CreateAutoScaleGroups(class string, dryRun bool) error {

	// Verify the class input

	// Verify that the current launch configuration exists in the regions of this ASG class

	/*
		svc := autoscaling.New(session.New())

		params := &autoscaling.CreateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String("XmlStringMaxLen255"), // Required
			MaxSize:              aws.Int64(1),                     // Required
			MinSize:              aws.Int64(1),                     // Required
			AvailabilityZones: []*string{
				aws.String("XmlStringMaxLen255"), // Required
				// More values...
			},
			DefaultCooldown:         aws.Int64(1),
			DesiredCapacity:         aws.Int64(1),
			HealthCheckGracePeriod:  aws.Int64(1),
			HealthCheckType:         aws.String("XmlStringMaxLen32"),
			InstanceId:              aws.String("XmlStringMaxLen19"),
			LaunchConfigurationName: aws.String("ResourceName"),
			LoadBalancerNames: []*string{
				aws.String("XmlStringMaxLen255"), // Required
				// More values...
			},
			NewInstancesProtectedFromScaleIn: aws.Bool(true),
			PlacementGroup:                   aws.String("XmlStringMaxLen255"),
			Tags: []*autoscaling.Tag{
				{ // Required
					Key:               aws.String("TagKey"), // Required
					PropagateAtLaunch: aws.Bool(true),
					ResourceId:        aws.String("XmlString"),
					ResourceType:      aws.String("XmlString"),
					Value:             aws.String("TagValue"),
				},
				// More values...
			},
			TerminationPolicies: []*string{
				aws.String("XmlStringMaxLen1600"), // Required
				// More values...
			},
			VPCZoneIdentifier: aws.String("XmlStringMaxLen255"),
		}
		resp, err := svc.CreateAutoScalingGroup(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return
		}

		// Pretty-print the response data.
		fmt.Println(resp)
	*/
	return nil

}

func UpdateAutoScaleGroups(class string, dryRun bool) error {
	/*
		svc := autoscaling.New(session.New())

		params := &autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String("ResourceName"), // Required
			AvailabilityZones: []*string{
				aws.String("XmlStringMaxLen255"), // Required
				// More values...
			},
			DefaultCooldown:         aws.Int64(1),
			DesiredCapacity:         aws.Int64(1),
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
		resp, err := svc.UpdateAutoScalingGroup(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
			return
		}

		// Pretty-print the response data.
		fmt.Println(resp)
	*/
	return nil
}

func SuspendProcesses(search, region string, dryRun bool) (err error) {
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
