package aws

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type AutoScaleGroups []AutoScaleGroup

type AutoScaleGroup struct {
	Name             string
	Class            string
	HealthCheck      string
	LaunchConfig     string
	LoadBalancer     string
	Instances        string
	DesiredCapacity  string
	MinSize          string
	MaxSize          string
	Cooldown         string
	AvailabilityZone string
	Subnets          string
}

func GetAutoScaleGroups() (*AutoScaleGroups, []error) {
	var wg sync.WaitGroup
	var errs []error

	asgList := new(AutoScaleGroups)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionAutoScaleGroups(region.RegionName, asgList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering autoscale group list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return asgList, errs
}

func GetRegionAutoScaleGroups(region *string, asgList *AutoScaleGroups) error {
	svc := autoscaling.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})

	if err != nil {
		return err
	}

	asg := make(AutoScaleGroups, len(result.AutoScalingGroups))
	for i, autoscalegroup := range result.AutoScalingGroups {
		asg[i] = AutoScaleGroup{
			Name:             aws.StringValue(autoscalegroup.AutoScalingGroupName),
			Class:            GetTagValue("Class", autoscalegroup.Tags),
			HealthCheck:      aws.StringValue(autoscalegroup.HealthCheckType),
			LaunchConfig:     aws.StringValue(autoscalegroup.LaunchConfigurationName),
			LoadBalancer:     strings.Join(aws.StringValueSlice(autoscalegroup.LoadBalancerNames), ","),
			Instances:        fmt.Sprint(len(autoscalegroup.Instances)),
			DesiredCapacity:  fmt.Sprint(aws.Int64Value(autoscalegroup.DesiredCapacity)),
			MinSize:          fmt.Sprint(aws.Int64Value(autoscalegroup.MinSize)),
			MaxSize:          fmt.Sprint(aws.Int64Value(autoscalegroup.MaxSize)),
			Cooldown:         fmt.Sprint(aws.Int64Value(autoscalegroup.DefaultCooldown)),
			AvailabilityZone: strings.Join(aws.StringValueSlice(autoscalegroup.AvailabilityZones), ","),
			//Subnets: aws.StringValue(autoscalegroup.PlacementGroup), // TODO
		}
	}
	*asgList = append(*asgList, asg[:]...)

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
			val.LoadBalancer,
			val.Instances,
			val.DesiredCapacity,
			val.MinSize,
			val.MaxSize,
			val.Cooldown,
			val.AvailabilityZone,
			val.Subnets,
		}
	}

	table.SetHeader([]string{"Name", "Class", "Health Check", "Launch Config", "Load Balancers", "Instances", "Desired Capacity", "Min Size", "Max Size", "Cooldown", "Availability Zone", "Subnets"})

	table.AppendBulk(rows)
	table.Render()
}
