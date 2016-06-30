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
	Region            string
	VpcName           string
	VpcId             string
	SubnetName        string
	SubnetId          string
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
			err := GetRegionAutoScaleGroups(*region.RegionName, asgList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering autoscale group list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return asgList, errs
}

func GetRegionAutoScaleGroups(region string, asgList *AutoScaleGroups) error {
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
	*asgList = append(*asgList, asg[:]...)

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
