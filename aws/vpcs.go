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
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// Vpcs represents a slice of VPCs
type Vpcs []Vpc

// Vpc represents a single VPC
type Vpc models.Vpc

// GetVpcByTag returns a single VPC that matches the provided region and Tag key/value
func GetVpcByTag(region, key, value string) (Vpc, error) {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + key),
				Values: []*string{
					aws.String(value),
				},
			},
		},
	}

	result, err := svc.DescribeVpcs(params)
	if err != nil {
		return Vpc{}, err
	}

	count := len(result.Vpcs)

	switch count {
	case 0:
		return Vpc{}, errors.New("No VPC found with [" + key + "] of [" + value + "] in [" + region + "], Aborting!")
	case 1:
		vpc := new(Vpc)
		vpc.Marshal(result.Vpcs[0], region)
		return *vpc, nil
	}

	return Vpc{}, errors.New("Found more than one VPC with [" + key + "] of [" + value + "] in [" + region + "], Aborting!")
}

// GetVpcSecurityGroupByTag returns a VPC Security Group that matches the provided Tag key/value
func (v *Vpc) GetVpcSecurityGroupByTag(key, value string) (SecurityGroup, error) {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(v.Region)}))

	params := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(v.VpcID),
				},
			},
			{
				Name: aws.String("tag:" + key),
				Values: []*string{
					aws.String(value),
				},
			},
		},
	}

	result, err := svc.DescribeSecurityGroups(params)
	if err != nil {
		return SecurityGroup{}, err
	}

	count := len(result.SecurityGroups)

	switch count {
	case 0:
		return SecurityGroup{}, errors.New("No VPC Security Group found with [" + key + "] of [" + value + "] in [" + v.Region + "], Aborting!")
	case 1:
		sec := new(SecurityGroup)
		sec.Marshal(result.SecurityGroups[0], v.Region, &Vpcs{*v})
		return *sec, nil
	}

	return SecurityGroup{}, errors.New("Found more than one VPC Security Group with [" + key + "] of [" + value + "] in [" + v.Region + "], Aborting!")
}

// GetVpcSecurityGroupByTagMulti returns a slice VPC Security Groups that matches the provided Tag key/value. Multiple values can be passed
func (v *Vpc) GetVpcSecurityGroupByTagMulti(key string, value []string) (SecurityGroups, error) {
	var secList SecurityGroups
	for _, val := range value {
		secgroup, err := v.GetVpcSecurityGroupByTag(key, val)
		if err != nil {
			return SecurityGroups{}, err
		}

		secList = append(secList, secgroup)
	}

	return secList, nil
}

// GetVpcSubnetByTag Gets a single VPC Subnet that matches the provided Tag key/value
func (v *Vpc) GetVpcSubnetByTag(key, value string) (Subnet, error) {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(v.Region)}))

	params := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + key),
				Values: []*string{
					aws.String(value),
				},
			},
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(v.VpcID),
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
		return Subnet{}, errors.New("No Subnet found with [" + key + "] of [" + value + "] in [" + v.Region + "] VPC [" + v.Name + "], Aborting!")
	case 1:
		subnet := new(Subnet)
		subnet.Marshal(result.Subnets[0], v.Region, &Vpcs{*v})
		return *subnet, nil
	}

	return Subnet{}, errors.New("Please limit your request to return only one Subnet")
}

// GetVpcs returns a slice of VPCs that match the provided search term
func GetVpcs(search string) (*Vpcs, []error) {
	var wg sync.WaitGroup
	var errs []error

	vpcList := new(Vpcs)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionVpcs(*region.RegionName, vpcList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering vpc list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return vpcList, errs
}

// Marshal parses the response from the aws sdk into an awsm Vpc
func (v *Vpc) Marshal(vpc *ec2.Vpc, region string) {
	v.Name = GetTagValue("Name", vpc.Tags)
	v.Class = GetTagValue("Class", vpc.Tags)
	v.VpcID = aws.StringValue(vpc.VpcId)
	v.State = aws.StringValue(vpc.State)
	v.Default = aws.BoolValue(vpc.IsDefault)
	v.CIDRBlock = aws.StringValue(vpc.CidrBlock)
	v.DHCPOptID = aws.StringValue(vpc.DhcpOptionsId)
	v.Tenancy = aws.StringValue(vpc.InstanceTenancy)
	v.Region = region
}

// GetRegionVpcs returns a list of a regions VPCs that match the provided search term
func GetRegionVpcs(region string, vpcList *Vpcs, search string) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{})

	if err != nil {
		return err
	}

	vpcs := make(Vpcs, len(result.Vpcs))
	for i, vpc := range result.Vpcs {
		vpcs[i].Marshal(vpc, region)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, v := range vpcs {
			rVpc := reflect.ValueOf(v)

			for k := 0; k < rVpc.NumField(); k++ {
				sVal := rVpc.Field(k).String()

				if term.MatchString(sVal) {
					*vpcList = append(*vpcList, vpcs[i])
					continue Loop
				}
			}
		}
	} else {
		*vpcList = append(*vpcList, vpcs[:]...)
	}

	return nil
}

// GetVpcName returns the name of a VPC given its ID
func (i *Vpcs) GetVpcName(id string) string {
	for _, vpc := range *i {
		if vpc.VpcID == id && vpc.Name != "" {
			return vpc.Name
		}
	}
	return id
}

// PrintTable Prints an ascii table of the list of VPCs
func (i *Vpcs) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No VPCs Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, vpc := range *i {
		models.ExtractAwsmTable(index, vpc, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

// CreateVpc creates a new VPC
func CreateVpc(class, name, ip, region string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	cfg, err := config.LoadVpcClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found VPC Class Configuration for [" + class + "]!")

	// Validate the region
	if !ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	// TODO limit to one VPC of a class per region, so that we can target VPCs by class instead of name.

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Create the VPC
	vpcParams := &ec2.CreateVpcInput{
		CidrBlock:       aws.String(ip + cfg.CIDR),
		DryRun:          aws.Bool(dryRun),
		InstanceTenancy: aws.String(cfg.Tenancy),
	}
	createVpcResp, err := svc.CreateVpc(vpcParams)

	if err != nil {
		return err
	}

	terminal.Information("Created VPC [" + *createVpcResp.Vpc.VpcId + "] named [" + name + "] in [" + region + "]!")

	// Add Tags
	err = SetEc2NameAndClassTags(createVpcResp.Vpc.VpcId, name, class, region)

	if err != nil {
		return err
	}

	return nil

}

// DeleteVpcs deletes one or more VPCs given the search term and optional region input
func DeleteVpcs(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	vpcList := new(Vpcs)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionVpcs(region, vpcList, search)
	} else {
		vpcList, _ = GetVpcs(search)
	}

	if err != nil {
		return errors.New("Error gathering VPC list")
	}

	if len(*vpcList) > 0 {
		// Print the table
		vpcList.PrintTable()
	} else {
		return errors.New("No VPCs found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these VPCs?") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = deleteVpcs(vpcList, dryRun)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Done!")

	return nil
}

// private function without the confirmation terminal prompts
func deleteVpcs(vpcList *Vpcs, dryRun bool) (err error) {
	for _, vpc := range *vpcList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(vpc.Region)}))

		params := &ec2.DeleteVpcInput{
			VpcId:  aws.String(vpc.VpcID),
			DryRun: aws.Bool(dryRun),
		}

		_, err := svc.DeleteVpc(params)
		if err != nil {
			return err
		}

		terminal.Information("Deleted VPC [" + vpc.Name + "] in [" + vpc.Region + "]!")
	}

	return nil
}
