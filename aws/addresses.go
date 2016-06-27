package aws

import (
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/terminal"

	"github.com/olekukonko/tablewriter"
)

type Addresses []Address

type Address struct {
	PublicIp   string
	PrivateIp  string
	Domain     string
	InstanceId string
	Region     string
}

func GetAddresses() (*Addresses, []error) {
	var wg sync.WaitGroup
	var errs []error

	ipList := new(Addresses)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionAddresses(region.RegionName, ipList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering address list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return ipList, errs
}

func GetRegionAddresses(region *string, adrList *Addresses) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeAddresses(&ec2.DescribeAddressesInput{})

	if err != nil {
		return err
	}

	adr := make(Addresses, len(result.Addresses))
	for i, address := range result.Addresses {

		adr[i] = Address{
			PublicIp:   aws.StringValue(address.PublicIp),
			PrivateIp:  aws.StringValue(address.PrivateIpAddress),
			InstanceId: aws.StringValue(address.InstanceId),
			Domain:     aws.StringValue(address.Domain),
			Region:     fmt.Sprintf(*region),
		}
	}
	*adrList = append(*adrList, adr[:]...)

	return nil
}

func (i *Addresses) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.PublicIp,
			val.PrivateIp,
			val.Domain,
			val.InstanceId,
			val.Region,
		}
	}

	table.SetHeader([]string{"Public IP", "Private IP", "Domain", "Instance Id", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
