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
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/dustin/go-humanize"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type LoadBalancers []LoadBalancer

type LoadBalancer struct {
	Name                string    `json:"name"`
	DNSName             string    `json:"dnsName"`
	Region              string    `json:"region"`
	AvailabilityZones   string    `json:"availabilityZone"`
	CreatedTime         time.Time `json:"createdTime"`
	CreatedHuman        string    `json:"createdHuman"`
	SecurityGroups      string    `json:"securityGroups"`
	Scheme              string    `json:"scheme"`
	VpcId               string    `json:"vpcId"`
	Vpc                 string    `json:"vpc"`
	HealthCheckTarget   string    `json:"healthCheckTarget"`
	HealthCheckInterval string    `json:"healthCheckInterval"`
	Subnets             string    `json:"subnets"`
	SubnetIds           []string  `json:"subnetIds"`
}

func GetLoadBalancers() (*LoadBalancers, []error) {
	var wg sync.WaitGroup
	var errs []error

	lbList := new(LoadBalancers)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionLoadBalancers(*region.RegionName, lbList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering loadbalancer list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return lbList, errs
}

func GetRegionLoadBalancers(region string, lbList *LoadBalancers) error {
	svc := elb.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})

	if err != nil {
		return err
	}

	secGrpList := new(SecurityGroups)
	err = GetRegionSecurityGroups(region, secGrpList, "")

	vpcList := new(Vpcs)
	subList := new(Subnets)
	GetRegionVpcs(region, vpcList, "")
	GetRegionSubnets(region, subList, "")

	lb := make(LoadBalancers, len(result.LoadBalancerDescriptions))
	for i, balancer := range result.LoadBalancerDescriptions {
		lb[i].Marshal(balancer, region, secGrpList, vpcList, subList)
	}
	*lbList = append(*lbList, lb[:]...)

	return nil
}

func (l *LoadBalancer) Marshal(balancer *elb.LoadBalancerDescription, region string, secGrpList *SecurityGroups, vpcList *Vpcs, subList *Subnets) {

	// security groups
	secGroupNames := secGrpList.GetSecurityGroupNames(aws.StringValueSlice(balancer.SecurityGroups))
	secGroupNamesSorted := sort.StringSlice(secGroupNames[0:])
	secGroupNamesSorted.Sort()

	// subnets
	subnetNames := subList.GetSubnetNames(aws.StringValueSlice(balancer.Subnets))
	subnetNamesSorted := sort.StringSlice(subnetNames[0:])
	subnetNamesSorted.Sort()

	l.Name = aws.StringValue(balancer.LoadBalancerName)
	l.DNSName = aws.StringValue(balancer.DNSName)
	l.CreatedTime = aws.TimeValue(balancer.CreatedTime) // robots
	l.CreatedHuman = humanize.Time(l.CreatedTime)       // humans
	l.VpcId = aws.StringValue(balancer.VPCId)
	l.Vpc = vpcList.GetVpcName(l.VpcId)
	l.SubnetIds = aws.StringValueSlice(balancer.Subnets)
	l.Subnets = strings.Join(subnetNamesSorted, ", ")
	l.HealthCheckTarget = aws.StringValue(balancer.HealthCheck.Target)
	l.HealthCheckInterval = fmt.Sprintf("%d seconds", *balancer.HealthCheck.Interval)
	l.Scheme = aws.StringValue(balancer.Scheme)
	l.SecurityGroups = strings.Join(secGroupNamesSorted, ", ")
	l.AvailabilityZones = strings.Join(aws.StringValueSlice(balancer.AvailabilityZones), ", ") // TODO
	l.Region = region
}

func (i *LoadBalancers) PrintTable() {

	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Load Balancers Found!")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.CreatedHuman,
			val.Vpc,
			val.Subnets,
			val.HealthCheckTarget,
			val.HealthCheckInterval,
			val.Scheme,
			val.SecurityGroups,
			val.AvailabilityZones,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Created", "VPC", "Subnets", "Health Check Target", "Interval", "Scheme", "Security Groups", "Availability Zones", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
