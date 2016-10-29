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
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// Subnets represents a slice of Subnets
type Subnets []Subnet

// Subnet represents a single Subnet
type Subnet models.Subnet

// GetSubnetNames returns a slice of Subnet Names given their ID's
func (s *Subnets) GetSubnetNames(ids []string) []string {
	names := make([]string, len(ids))
	for i, id := range ids {
		for _, sub := range *s {
			if sub.SubnetID == id && sub.Name != "" {
				names[i] = sub.Name
			} else if sub.SubnetID == id {
				names[i] = sub.SubnetID
			}
		}
	}
	return names
}

// GetSubnetByTag returns a single Subnet given the provided region and Tag key/value
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

// GetSubnets returns a slice of Subnets that match the provided search term
func GetSubnets(search string) (*Subnets, []error) {
	var wg sync.WaitGroup
	var errs []error

	subList := new(Subnets)
	regions := regions.GetRegionList()

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

// Marshal parses the response from the aws sdk into an awsm Subnet
func (s *Subnet) Marshal(subnet *ec2.Subnet, region string, vpcList *Vpcs) {
	s.Name = GetTagValue("Name", subnet.Tags)
	s.Class = GetTagValue("Class", subnet.Tags)
	s.SubnetID = aws.StringValue(subnet.SubnetId)
	s.VpcID = aws.StringValue(subnet.VpcId)
	s.VpcName = vpcList.GetVpcName(s.VpcID)
	s.State = aws.StringValue(subnet.State)
	s.AvailabilityZone = aws.StringValue(subnet.AvailabilityZone)
	s.Default = aws.BoolValue(subnet.DefaultForAz)
	s.CIDRBlock = aws.StringValue(subnet.CidrBlock)
	s.AvailableIPs = int(aws.Int64Value(subnet.AvailableIpAddressCount))
	s.MapPublicIP = aws.BoolValue(subnet.MapPublicIpOnLaunch)
	s.Region = region
}

// GetRegionSubnets returns a list of Subnets of a region into the provided Subnets slice
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

// GetSubnetsByVpcID returns a slice of Subnets that belong to the provided VPC ID
func GetSubnetsByVpcID(vpcID string, region string) (Subnets, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(vpcID),
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

// GetSubnetName returns the name of a Subnet given its ID
func (s *Subnets) GetSubnetName(id string) string {
	for _, subnet := range *s {
		if subnet.SubnetID == id && subnet.Name != "" {
			return subnet.Name
		}
	}
	return id
}

// GetVpcIDBySubnetID returns the ID of a VPC given the ID of a Subnet
func (s *Subnets) GetVpcIDBySubnetID(id string) string {
	for _, subnet := range *s {
		if subnet.SubnetID == id && subnet.VpcName != "" {
			return subnet.VpcName
		} else if subnet.SubnetID == id {
			return subnet.VpcID
		}
	}
	return ""
}

// GetVpcNameBySubnetID returns the Name of a VPC given the ID of a Subnet
func (s *Subnets) GetVpcNameBySubnetID(id string) string {
	for _, subnet := range *s {
		if subnet.SubnetID == id && subnet.VpcName != "" {
			return subnet.VpcName
		} else if subnet.SubnetID == id {
			return subnet.VpcID
		}
	}
	return ""
}

// CreateSubnet creates a new VPC Subnet
func CreateSubnet(class, name, vpc, ip, az string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	cfg, err := config.LoadSubnetClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found Subnet Class Configuration for [" + class + "]!")

	// Verify the az input
	azs, errs := regions.GetAZs()
	if errs != nil {
		return errors.New("Error Verifying Availability Zone input")
	}
	if !azs.ValidAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	}

	terminal.Information("Found Availability Zone [" + az + "]!")

	region := azs.GetRegion(az)

	// Verify the vpc input
	targetVpc, err := GetVpcByTag(region, "Class", vpc)
	if err != nil {
		return err
	}
	terminal.Information("Found [" + targetVpc.Name + "] VPC [" + targetVpc.VpcID + "]!")

	// Create the Subnet

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.CreateSubnetInput{
		CidrBlock:        aws.String(ip + cfg.CIDR),
		VpcId:            aws.String(targetVpc.VpcID),
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

// DeleteSubnets deletes one or more VPC Subnets based on the given name and optional region input.
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
	if !terminal.PromptBool("Are you sure you want to delete these Subnets?") {
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

// private function without the confirmation terminal prompts
func deleteSubnets(subnetList *Subnets, dryRun bool) (err error) {
	for _, subnet := range *subnetList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(subnet.Region)}))

		params := &ec2.DeleteSubnetInput{
			SubnetId: aws.String(subnet.SubnetID),
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

// PrintTable Prints an ascii table of the list of Subnets
func (s *Subnets) PrintTable() {
	if len(*s) == 0 {
		terminal.ShowErrorMessage("Warning", "No Subnets Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*s))

	for index, subnet := range *s {
		models.ExtractAwsmTable(index, subnet, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
