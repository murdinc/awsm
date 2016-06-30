package aws

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type Subnets []Subnet

type Subnet struct {
	Name             string
	Class            string
	SubnetId         string
	VpcName          string
	VpcId            string
	State            string
	AvailabilityZone string
	Default          string
	CIDRBlock        string
	AvailableIPs     string
	MapPublicIp      bool
	Region           string
}

func GetSubnetByTag(region, key, value string) (Subnet, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + key),
				Values: []*string{
					aws.String(value),
				},
			},
		},
	}

	result, err := svc.DescribeSubnets(params)

	if err != nil {
		return Subnet{}, err
	}

	count := len(result.Subnets)

	switch count {
	case 0:
		return Subnet{}, errors.New("No Subnet found with [" + key + "] of [" + value + "] in [" + region + "], Aborting!")
	case 1:
		vpcList := new(Vpcs)
		GetRegionVpcs(region, vpcList, "")

		subnet := new(Subnet)
		subnet.Marshal(result.Subnets[0], region, vpcList)
		return *subnet, nil
	}

	return Subnet{}, errors.New("Found more than one Subnet with [" + key + "] of [" + value + "] in [" + region + "], Aborting!")
}

func GetSubnets(search string) (*Subnets, []error) {
	var wg sync.WaitGroup
	var errs []error

	subList := new(Subnets)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSubnets(*region.RegionName, subList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering subnet list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return subList, errs
}

func (s *Subnet) Marshal(subnet *ec2.Subnet, region string, vpcList *Vpcs) {
	s.Name = GetTagValue("Name", subnet.Tags)
	s.Class = GetTagValue("Class", subnet.Tags)
	s.SubnetId = aws.StringValue(subnet.SubnetId)
	s.VpcId = aws.StringValue(subnet.VpcId)
	s.VpcName = vpcList.GetVpcName(s.VpcId)
	s.State = aws.StringValue(subnet.State)
	s.AvailabilityZone = aws.StringValue(subnet.AvailabilityZone)
	s.Default = fmt.Sprintf("%t", aws.BoolValue(subnet.DefaultForAz))
	s.CIDRBlock = aws.StringValue(subnet.CidrBlock)
	s.AvailableIPs = fmt.Sprint(aws.Int64Value(subnet.AvailableIpAddressCount))
	s.MapPublicIp = aws.BoolValue(subnet.MapPublicIpOnLaunch)
	s.Region = region
}

func GetRegionSubnets(region string, subList *Subnets, search string) error {
	svc := ec2.New(session.New(&aws.Config{Region: &region}))
	result, err := svc.DescribeSubnets(&ec2.DescribeSubnetsInput{})

	if err != nil {
		return err
	}

	vpcList := new(Vpcs)
	GetRegionVpcs(region, vpcList, "")

	subs := make(Subnets, len(result.Subnets))
	for i, subnet := range result.Subnets {
		subs[i].Marshal(subnet, region, vpcList)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, s := range subs {
			rVpc := reflect.ValueOf(s)

			for k := 0; k < rVpc.NumField(); k++ {
				sVal := rVpc.Field(k).String()

				if term.MatchString(sVal) {
					*subList = append(*subList, subs[i])
					continue Loop
				}
			}
		}
	} else {
		*subList = append(*subList, subs[:]...)
	}

	return nil
}

func GetSubnetsByVpcId(vpcId string, region string) (Subnets, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(vpcId),
				},
			},
		},
	}

	result, err := svc.DescribeSubnets(params)

	if err != nil {
		return Subnets{}, err
	}

	vpcList := new(Vpcs)
	GetRegionVpcs(region, vpcList, "")

	subList := make(Subnets, len(result.Subnets))
	for i, subnet := range result.Subnets {
		subList[i].Marshal(subnet, region, vpcList)
	}

	return Subnets{}, nil
}

func (i *Subnets) GetSubnetName(id string) string {
	for _, subnet := range *i {
		if subnet.SubnetId == id && subnet.Name != "" {
			return subnet.Name
		}
	}
	return id
}

func (i *Subnets) GetVpcIdBySubnetId(id string) string {
	for _, subnet := range *i {
		if subnet.SubnetId == id && subnet.VpcName != "" {
			return subnet.VpcName
		} else if subnet.SubnetId == id {
			return subnet.VpcId
		}
	}
	return ""
}

func (i *Subnets) GetVpcNameBySubnetId(id string) string {
	for _, subnet := range *i {
		if subnet.SubnetId == id && subnet.VpcName != "" {
			return subnet.VpcName
		} else if subnet.SubnetId == id {
			return subnet.VpcId
		}
	}
	return ""
}

func CreateSubnet(class, name, vpc, ip, az string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	var cfg config.SubnetClassConfig
	err := cfg.LoadConfig(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Subnet Class Configuration for [" + class + "]!")
	}

	// Verify the az input
	azs, errs := GetAZs()
	if errs != nil {
		return errors.New("Error Verifying Availability Zone input")
	}
	if !azs.ValidAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	} else {
		terminal.Information("Found Availability Zone [" + az + "]!")
	}

	region := azs.GetRegion(az)

	// Verify the vpc input
	targetVpc, err := GetVpcByTag(region, "Class", vpc)
	if err != nil {
		return err
	}
	terminal.Information("Found [" + targetVpc.Name + "] VPC [" + targetVpc.VpcId + "]!")

	// Create the Subnet

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.CreateSubnetInput{
		CidrBlock:        aws.String(ip + cfg.CIDR),
		VpcId:            aws.String(targetVpc.VpcId),
		DryRun:           aws.Bool(dryRun),
		AvailabilityZone: aws.String(az),
	}

	createSubnetResp, err := svc.CreateSubnet(params)

	if err != nil {
		return err
	}

	terminal.Information("Created Subnet [" + *createSubnetResp.Subnet.SubnetId + "] named [" + name + "] in [" + *createSubnetResp.Subnet.AvailabilityZone + "]!")

	// Add Tags
	err = SetEc2NameAndClassTags(createSubnetResp.Subnet.SubnetId, name, class, region)

	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Public function with confirmation terminal prompt
func DeleteSubnets(name, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	subnetList := new(Subnets)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionSubnets(region, subnetList, name)
	} else {
		subnetList, _ = GetSubnets(name)
	}

	if err != nil {
		return errors.New("Error gathering Subnet list")
	}

	if len(*subnetList) > 0 {
		// Print the table
		subnetList.PrintTable()
	} else {
		return errors.New("No Subnets found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Subnets") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = deleteSubnets(subnetList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func deleteSubnets(subnetList *Subnets, dryRun bool) (err error) {
	for _, subnet := range *subnetList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(subnet.Region)}))

		params := &ec2.DeleteSubnetInput{
			SubnetId: aws.String(subnet.SubnetId),
			DryRun:   aws.Bool(dryRun),
		}

		_, err := svc.DeleteSubnet(params)
		if err != nil {
			return err
		}

		terminal.Information("Deleted Subnet [" + subnet.Name + "] in [" + subnet.Region + "]!")
	}

	return nil
}

func (i *Subnets) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.Class,
			val.SubnetId,
			val.VpcName,
			val.VpcId,
			val.State,
			val.AvailabilityZone,
			val.Default,
			val.CIDRBlock,
			val.AvailableIPs,
			fmt.Sprintf("%t", val.MapPublicIp),
		}
	}

	table.SetHeader([]string{"Name", "Class", "Subnet Id", "VPC", "VPC Id", "State", "Availability Zone", "Default for AZ", "CIDR Block", "Available IPs", "Map Public IP"})

	table.AppendBulk(rows)
	table.Render()
}
