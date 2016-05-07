package aws

import (
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/terminal"
	"github.com/olekukonko/tablewriter"
)

type Vpcs []Vpc

// "Name", "VPC Id", "State", "Default", "CIDR Block", "DHCP Options ID", "Tenancy"

type Vpc struct {
	Name      string
	VpcId     string
	State     string
	Default   string
	CIDRBlock string
	DHCPOptId string
	Tenancy   string
	Region    string
}

func GetVpcs() (*Vpcs, error) {
	var wg sync.WaitGroup

	vpcList := new(Vpcs)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionVpcs(region.RegionName, vpcList)
			if err != nil {
				terminal.ShowErrorMessage("Error gathering Vpc list", err.Error())
			}
		}(region)
	}
	wg.Wait()

	return vpcList, nil
}

func GetRegionVpcs(region *string, vpcList *Vpcs) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{})

	if err != nil {
		return err
	}

	v := make(Vpcs, len(result.Vpcs))
	for i, vpc := range result.Vpcs {
		v[i] = Vpc{
			Name:      GetTagValue("Name", vpc.Tags),
			VpcId:     aws.StringValue(vpc.VpcId),
			State:     aws.StringValue(vpc.State),
			Default:   fmt.Sprintf("%t", aws.BoolValue(vpc.IsDefault)),
			CIDRBlock: aws.StringValue(vpc.CidrBlock),
			DHCPOptId: aws.StringValue(vpc.DhcpOptionsId),
			Tenancy:   aws.StringValue(vpc.InstanceTenancy),
			Region:    *region,
		}
	}
	*vpcList = append(*vpcList, v[:]...)

	return nil
}

func (i *Vpcs) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.VpcId,
			val.State,
			val.Default,
			val.CIDRBlock,
			val.DHCPOptId,
			val.Tenancy,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "VPC Id", "State", "Default", "CIDR Block", "DHCP Options ID", "Tenancy"})

	table.AppendBulk(rows)
	table.Render()
}
