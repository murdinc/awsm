package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/cli"
)

type Subnets []Subnet

type Subnet struct {
	Name             string
	SubnetId         string
	VpcId            string
	State            string
	AvailabilityZone string
	Default          string
	CIDRBlock        string
	AvailableIPs     string
	MapPublicIp      string
}

func GetSubnets() (*Subnets, error) {
	var wg sync.WaitGroup

	subList := new(Subnets)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSubnets(region.RegionName, subList)
			if err != nil {
				cli.ShowErrorMessage("Error gathering Subnet list", err.Error())
			}
		}(region)
	}
	wg.Wait()

	return subList, nil
}

func GetRegionSubnets(region *string, subList *Subnets) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeSubnets(&ec2.DescribeSubnetsInput{})

	if err != nil {
		return err
	}

	sub := make(Subnets, len(result.Subnets))
	for i, subnet := range result.Subnets {
		//"Name", "Subnet Id", "VPC Id", "State", "Availability Zone", "Default for AZ", "CIDR Block", "Available IPs", "Map Public IP"
		sub[i] = Subnet{
			Name:             GetTagValue("Name", subnet.Tags),
			SubnetId:         aws.StringValue(subnet.SubnetId),
			VpcId:            aws.StringValue(subnet.VpcId),
			State:            aws.StringValue(subnet.State),
			AvailabilityZone: aws.StringValue(subnet.AvailabilityZone),
			Default:          fmt.Sprintf("%t", aws.BoolValue(subnet.DefaultForAz)),
			CIDRBlock:        aws.StringValue(subnet.CidrBlock),
			AvailableIPs:     fmt.Sprint(aws.Int64Value(subnet.AvailableIpAddressCount)),
		}
	}
	*subList = append(*subList, sub[:]...)

	return nil
}

func (i *Subnets) PrintTable() {
	collumns := []string{"Name", "Subnet Id", "VPC Id", "State", "Availability Zone", "Default for AZ", "CIDR Block", "Available IPs", "Map Public IP"}

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.SubnetId,
			val.VpcId,
			val.State,
			val.AvailabilityZone,
			val.Default,
			val.CIDRBlock,
			val.AvailableIPs,
			val.MapPublicIp,
		}
	}

	printTable(collumns, rows)
}
