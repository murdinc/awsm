package aws

import (
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/murdinc/awsm/terminal"
	"github.com/olekukonko/tablewriter"
)

type LoadBalancers []LoadBalancer

type LoadBalancer struct {
	Name    string
	DNSName string
	Region  string
}

func GetLoadBalancers() (*LoadBalancers, error) {
	var wg sync.WaitGroup

	lbList := new(LoadBalancers)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionLoadBalancers(region.RegionName, lbList)
			if err != nil {
				terminal.ShowErrorMessage("Error gathering launch config list", err.Error())
			}
		}(region)
	}
	wg.Wait()

	return lbList, nil
}

func GetRegionLoadBalancers(region *string, lbList *LoadBalancers) error {
	svc := elb.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})

	if err != nil {
		return err
	}

	lb := make(LoadBalancers, len(result.LoadBalancerDescriptions))
	for i, balancer := range result.LoadBalancerDescriptions {

		lb[i] = LoadBalancer{
			Name:    *balancer.LoadBalancerName,
			DNSName: *balancer.DNSName,
			Region:  fmt.Sprintf(*region),
		}
	}
	*lbList = append(*lbList, lb[:]...)

	return nil
}

func (i *LoadBalancers) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.DNSName,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "DNS Name", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
