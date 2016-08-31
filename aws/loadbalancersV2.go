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
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/dustin/go-humanize"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type LoadBalancersV2 []LoadBalancerV2

type LoadBalancerV2 struct {
	Name                  string    `json:"name"`
	DNSName               string    `json:"dnsName"`
	Type                  string    `json:"type"`
	State                 string    `json:"state"`
	Region                string    `json:"region"`
	AvailabilityZones     string    `json:"availabilityZone"`
	CreatedTime           time.Time `json:"createdTime"`
	CreatedHuman          string    `json:"createdHuman"`
	SecurityGroups        string    `json:"securityGroups"`
	Scheme                string    `json:"scheme"`
	CanonicalHostedZoneId string    `json:"canonicalHostedZoneId"`
	LoadBalancerArn       string    `json:"loadBalancerArn"`
	VpcId                 string    `json:"vpcId"`
}

func GetLoadBalancersV2() (*LoadBalancersV2, []error) {
	var wg sync.WaitGroup
	var errs []error

	lbList := new(LoadBalancersV2)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionLoadBalancersV2(*region.RegionName, lbList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering loadbalancer list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return lbList, errs
}

func GetRegionLoadBalancersV2(region string, lbList *LoadBalancersV2) error {
	svc := elbv2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{})

	fmt.Println(result)

	if err != nil {
		return err
	}

	secGrpList := new(SecurityGroups)
	err = GetRegionSecurityGroups(region, secGrpList, "")

	lb := make(LoadBalancersV2, len(result.LoadBalancers))
	for i, balancer := range result.LoadBalancers {
		lb[i].Marshal(balancer, region, secGrpList)
	}
	*lbList = append(*lbList, lb[:]...)

	return nil
}

func (l *LoadBalancerV2) Marshal(balancer *elbv2.LoadBalancer, region string, secGrpList *SecurityGroups) {

	secGroupNames := secGrpList.GetSecurityGroupNames(aws.StringValueSlice(balancer.SecurityGroups))
	secGroupNamesSorted := sort.StringSlice(secGroupNames[0:])
	secGroupNamesSorted.Sort()

	l.Name = aws.StringValue(balancer.LoadBalancerName)
	l.DNSName = aws.StringValue(balancer.DNSName)
	l.CreatedTime = aws.TimeValue(balancer.CreatedTime) // robots
	l.CreatedHuman = humanize.Time(l.CreatedTime)       // humans
	l.VpcId = aws.StringValue(balancer.VpcId)
	l.Type = aws.StringValue(balancer.Type)
	l.State = balancer.State.String()
	l.Scheme = aws.StringValue(balancer.Scheme)
	l.CanonicalHostedZoneId = aws.StringValue(balancer.CanonicalHostedZoneId)
	l.LoadBalancerArn = aws.StringValue(balancer.LoadBalancerArn)
	l.SecurityGroups = strings.Join(secGroupNamesSorted, ", ")
	//	l.AvailabilityZones = strings.Join(aws.StringValueSlice(balancer.AvailabilityZones), ", ") // TODO
	l.Region = region

}

func (i *LoadBalancersV2) PrintTable() {

	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Application Load Balancers Found!")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.DNSName,
			val.CreatedHuman,
			val.VpcId,
			val.Type,
			val.State,
			val.Scheme,
			val.CanonicalHostedZoneId,
			val.LoadBalancerArn,
			val.SecurityGroups,
			val.AvailabilityZones,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "DNS Name", "Created", "VPC Id", "Type", "State", "Scheme", "Canonical Hosted Zone Id", "Load Balancer ARN", "Security Groups", "Availability Zones", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
