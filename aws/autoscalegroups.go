package aws

import (
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/cli"
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

func GetAutoScaleGroups() (*AutoScaleGroups, error) {
	var wg sync.WaitGroup

	asgList := new(AutoScaleGroups)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionAutoScaleGroups(region.RegionName, asgList)
			if err != nil {
				cli.ShowErrorMessage("Error gathering AutoScaleGroup list", err.Error())
			}
		}(region)
	}
	wg.Wait()

	return asgList, nil
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
			Name:             GetTagValue("Name", autoscalegroup.Tags),
			Class:            GetTagValue("Class", autoscalegroup.Tags),
			HealthCheck:      aws.StringValue(autoscalegroup.HealthCheckType),
			LaunchConfig:     aws.StringValue(autoscalegroup.LaunchConfigurationName),
			LoadBalancer:     strings.Join(aws.StringValueSlice(autoscalegroup.LoadBalancerNames), ","),
			Instances:        string(len(autoscalegroup.Instances)),
			DesiredCapacity:  string(aws.Int64Value(autoscalegroup.DesiredCapacity)),
			MinSize:          string(aws.Int64Value(autoscalegroup.MinSize)),
			MaxSize:          string(aws.Int64Value(autoscalegroup.MaxSize)),
			Cooldown:         string(aws.Int64Value(autoscalegroup.DefaultCooldown)),
			AvailabilityZone: strings.Join(aws.StringValueSlice(autoscalegroup.AvailabilityZones), ","),
			//Subnets: aws.StringValue(autoscalegroup.PlacementGroup), // TODO
		}
	}
	*asgList = append(*asgList, asg[:]...)

	return nil
}

func (i *AutoScaleGroups) PrintTable() {
	collumns := []string{"Name", "Class", "Health Check", "Launch Config", "Load Balancers", "Instances", "Desired Capacity", "Min Size", "Max Size", "Cooldown", "Availability Zone", "Subnets"}

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

	printTable(collumns, rows)
}
