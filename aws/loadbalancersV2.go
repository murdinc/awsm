package aws

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// LoadBalancersV2 represents a slice of Application Load Balancers
type LoadBalancersV2 []LoadBalancerV2

// LoadBalancerV2 represents a single Application Load Balancer
type LoadBalancerV2 models.LoadBalancerV2

// GetLoadBalancersV2 returns a slice of Application Load Balancers
func GetLoadBalancersV2() (*LoadBalancersV2, []error) {
	var wg sync.WaitGroup
	var errs []error

	lbList := new(LoadBalancersV2)
	regions := regions.GetRegionList()

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

// GetRegionLoadBalancersV2 returns a slice of Application Load Balancers in the region into the provided LoadBalancersV2 slice
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

// Marshal parses the response from the aws sdk into an awsm LoadBalancerV2
func (l *LoadBalancerV2) Marshal(balancer *elbv2.LoadBalancer, region string, secGrpList *SecurityGroups) {

	secGroupNames := secGrpList.GetSecurityGroupNames(aws.StringValueSlice(balancer.SecurityGroups))
	secGroupNamesSorted := sort.StringSlice(secGroupNames[0:])
	secGroupNamesSorted.Sort()

	l.Name = aws.StringValue(balancer.LoadBalancerName)
	l.DNSName = aws.StringValue(balancer.DNSName)
	l.CreatedTime = aws.TimeValue(balancer.CreatedTime)
	l.VpcID = aws.StringValue(balancer.VpcId)
	l.Type = aws.StringValue(balancer.Type)
	l.State = balancer.State.String()
	l.Scheme = aws.StringValue(balancer.Scheme)
	l.CanonicalHostedZoneID = aws.StringValue(balancer.CanonicalHostedZoneId)
	l.LoadBalancerArn = aws.StringValue(balancer.LoadBalancerArn)
	l.SecurityGroups = strings.Join(secGroupNamesSorted, ", ")
	//	l.AvailabilityZones = strings.Join(aws.StringValueSlice(balancer.AvailabilityZones), ", ") // TODO
	l.Region = region

}

// PrintTable Prints an ascii table of the list of Application Load Balancers
func (i *LoadBalancersV2) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Application Load Balancers Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, lb := range *i {
		models.ExtractAwsmTable(index, lb, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
