package aws

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

// GetSubnetClasses returns a slice of Subnet Classes given their ID's
func (s *Subnets) GetSubnetClasses(ids []string) []string {
	classes := make([]string, len(ids))
	for i, id := range ids {
		for _, sub := range *s {
			if sub.SubnetID == id && sub.Class != "" {
				classes[i] = sub.Class
			} else if sub.SubnetID == id {
				classes[i] = sub.SubnetID
			}
		}
	}
	return classes
}

// GetSubnetByTag returns a single Subnet given the provided region and Tag key/value
func GetSubnetByTag(region, key, value string) (Subnet, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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
func CreateSubnet(class, name, vpcSearch, ip, az string, dryRun bool) error {

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

	// Verify the VPC input
	vpcs, _ := GetVpcs(vpcSearch)
	vpcCount := len(*vpcs)
	if vpcCount == 0 {
		return errors.New("No VPCs found for your search terms.")
	}
	if vpcCount > 1 {
		vpcs.PrintTable()
		return errors.New("Please limit your search to return only one VPC.")
	}
	vpc := (*vpcs)[0]
	region := vpc.Region

	terminal.Information("Found VPC [" + vpc.VpcID + "] named [" + vpc.Name + "] with a class of [" + vpc.Class + "] in [" + region + "]!")

	// Verify the az input
	if az != "" {
		azs, errs := regions.GetAZs()
		if errs != nil {
			return errors.New("Error Verifying Availability Zone input")
		}
		if !azs.ValidAZ(az) {
			return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
		}
		terminal.Information("Found Availability Zone [" + az + "]!")

		azRegion := azs.GetRegion(az)

		if azRegion != vpc.Region {
			return cli.NewExitError("Availability Zone ["+az+"] is not in the same region as the VPC ["+vpc.Region+"]!", 1)
		}
	}

	// Create the Subnet
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(vpc.Region)}))
	svc := ec2.New(sess)

	params := &ec2.CreateSubnetInput{
		CidrBlock: aws.String(ip + cfg.CIDR),
		VpcId:     aws.String(vpc.VpcID),
		DryRun:    aws.Bool(dryRun),
	}

	if az != "" {
		params.SetAvailabilityZone(az)
	}

	createSubnetResp, err := svc.CreateSubnet(params)
	if err != nil {
		return err
	}

	subnetId := *createSubnetResp.Subnet.SubnetId
	subnetAz := *createSubnetResp.Subnet.AvailabilityZone

	terminal.Delta("Created Subnet [" + subnetId + "] named [" + name + "] in [" + subnetAz + "]!")

	terminal.Notice("Waiting to tag Subnet...")

	err = svc.WaitUntilSubnetAvailable(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{
			aws.String(subnetId),
		},
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
	}

	terminal.Delta("Adding Subnet Tags...")

	// Add Tags
	err = SetEc2NameAndClassTags(createSubnetResp.Subnet.SubnetId, name, class, vpc.Region)
	if err != nil {
		return err
	}

	namePrefix := vpc.Name
	if namePrefix == "" {
		namePrefix = vpc.VpcID
	}

	if cfg.CreateInternetGateway {
		terminal.Notice("Creating a new Internet Gateway...")

		// Create a new Internet Gateway
		internetGatewayId, err := CreateInternetGateway(namePrefix+"-internet-gateway", region, dryRun)
		if err != nil {
			return err
		}

		terminal.Notice("Attaching Internet Gateway to VPC...")
		// Attach the Internet Gateway to the VPC
		err = attachInternetGateway(vpc.VpcID, internetGatewayId, region, dryRun)
		if err != nil {
			return err
		}

		// Add to main route table, if specified
		if cfg.AddInternetGatewayToMainRouteTable {
			terminal.Notice("Adding Internet Gateway to Main Route Table...")

			routeTable, err := GetVpcMainRouteTable(vpc.VpcID, region)
			if err != nil {
				return err
			}

			err = createRoute(routeTable.RouteTableID, region, "0.0.0.0/0", internetGatewayId, "", dryRun)
			if err != nil {
				return err
			}
		}

		// Create a new route table and add to that, if specified
		if cfg.AddInternetGatewayToNewRouteTable {
			terminal.Notice("Creating a new Route Table...")

			routeTableId, err := createRouteTable(namePrefix+"-"+name+"-route-table", vpc.VpcID, region, dryRun)

			terminal.Notice("Associating Route Table to Subnet...")

			err = associateRouteTable(routeTableId, subnetId, region, dryRun)
			if err != nil {
				return err
			}

			terminal.Notice("Adding Internet Gateway to New Route Table...")

			err = createRoute(routeTableId, region, "0.0.0.0/0", internetGatewayId, "", dryRun)
			if err != nil {
				return err
			}
		}

	}

	if cfg.CreateNatGateway {

		terminal.Notice("Creating a new NAT Gateway...")

		// Create a new NAT Gateway
		natGatewayId, err := CreateNatGateway(namePrefix+"-nat-gateway", "", subnetId, region, dryRun)
		if err != nil {
			return err
		}

		// Add to main route table, if specified
		if cfg.AddNatGatewayToMainRouteTable {
			terminal.Notice("Adding NAT Gateway to Main Route Table...")

			routeTable, err := GetVpcMainRouteTable(vpc.VpcID, region)
			if err != nil {
				return err
			}

			err = createRoute(routeTable.RouteTableID, region, "0.0.0.0/0", "", natGatewayId, dryRun)
			if err != nil {
				return err
			}
		}

		// Create a new route table and add to that, if specified
		if cfg.AddNatGatewayToNewRouteTable {
			terminal.Notice("Creating a new Route Table...")

			routeTableId, err := createRouteTable(namePrefix+"-"+name+"-route-table", vpc.VpcID, region, dryRun)

			terminal.Notice("Associating Route Table to Subnet...")

			err = associateRouteTable(routeTableId, subnetId, region, dryRun)
			if err != nil {
				return err
			}

			terminal.Notice("Adding NAT Gateway to New Route Table...")

			err = createRoute(routeTableId, region, "0.0.0.0/0", "", natGatewayId, dryRun)
			if err != nil {
				return err
			}
		}
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
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(subnet.Region)}))
		svc := ec2.New(sess)

		params := &ec2.DeleteSubnetInput{
			SubnetId: aws.String(subnet.SubnetID),
			DryRun:   aws.Bool(dryRun),
		}

		_, err := svc.DeleteSubnet(params)
		if err != nil {
			return err
		}

		terminal.Delta("Deleted Subnet [" + subnet.Name + "] in [" + subnet.Region + "]!")
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
