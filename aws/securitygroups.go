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
	"github.com/mitchellh/hashstructure"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// SecurityGroups represents a slice of Security Groups
type SecurityGroups []SecurityGroup

// SecurityGroup represents a single Security Group
type SecurityGroup models.SecurityGroup

// GetSecurityGroupNames returns a slice of the security group names (or ID's if a name is not available)
func (s *SecurityGroups) GetSecurityGroupNames(ids []string) []string {
	names := make([]string, len(ids))
	for i, id := range ids {
		for _, secGrp := range *s {
			if secGrp.GroupID == id && secGrp.Name != "" {
				names[i] = secGrp.Name
			} else if secGrp.GroupID == id {
				names[i] = secGrp.GroupID
			}
		}
	}
	return names
}

// GetSecurityGroupByTag returns a single Security Group that matches a provided region and key/value tag
func GetSecurityGroupByTag(region, key, value string) (SecurityGroup, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
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
		return SecurityGroup{}, errors.New("No Security Group found with [" + key + "] of [" + value + "] in [" + region + "], Aborting!")
	case 1:
		vpcList := new(Vpcs)
		GetRegionVpcs(region, vpcList, "")

		sec := new(SecurityGroup)
		sec.Marshal(result.SecurityGroups[0], region, vpcList)
		return *sec, nil
	}

	return SecurityGroup{}, errors.New("Found more than one Security Group with [" + key + "] of [" + value + "] in [" + region + "], Aborting!")
}

// GetSecurityGroupByTagMulti returns a slice of Security Groups that matches a provided region and key/value tag. Accepts multiple tag values for a single key.
func GetSecurityGroupByTagMulti(region, key string, value []string) (SecurityGroups, error) {
	var secList SecurityGroups
	for _, v := range value {
		secgroup, err := GetSecurityGroupByTag(region, key, v)
		if err != nil {
			return SecurityGroups{}, err
		}

		secList = append(secList, secgroup)
	}

	return secList, nil
}

// GetSecurityGroups returns a slice of Security Groups given a provided search term
func GetSecurityGroups(search string) (*SecurityGroups, []error) {
	var wg sync.WaitGroup
	var errs []error

	secGrpList := new(SecurityGroups)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSecurityGroups(*region.RegionName, secGrpList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering security group list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return secGrpList, errs
}

// GetRegionSecurityGroups returns a regions Security Groups into the provided SecurityGroups slice
func GetRegionSecurityGroups(region string, secGrpList *SecurityGroups, search string) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})

	if err != nil {
		return err
	}

	vpcList := new(Vpcs)
	GetRegionVpcs(region, vpcList, "")

	sgroup := make(SecurityGroups, len(result.SecurityGroups))
	for i, securitygroup := range result.SecurityGroups {
		sgroup[i].Marshal(securitygroup, region, vpcList)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, in := range sgroup {
			rSec := reflect.ValueOf(in)

			for k := 0; k < rSec.NumField(); k++ {
				sVal := rSec.Field(k).String()

				if term.MatchString(sVal) {
					*secGrpList = append(*secGrpList, sgroup[i])
					continue Loop
				}
			}
		}
	} else {
		*secGrpList = append(*secGrpList, sgroup[:]...)
	}

	return nil
}

// Marshal parses the response from the aws sdk into an awsm Security Group
func (s *SecurityGroup) Marshal(securitygroup *ec2.SecurityGroup, region string, vpcList *Vpcs) {

	vpc := vpcList.GetVpcName(aws.StringValue(securitygroup.VpcId))

	s.Name = aws.StringValue(securitygroup.GroupName)
	s.Class = GetTagValue("Class", securitygroup.Tags)
	s.GroupID = aws.StringValue(securitygroup.GroupId)
	s.Description = aws.StringValue(securitygroup.Description)
	s.Vpc = vpc
	s.VpcID = aws.StringValue(securitygroup.VpcId)
	s.Region = region

	// Get the ingress grants
	for _, grant := range securitygroup.IpPermissions {

		var cidr []string
		for _, ipRange := range grant.IpRanges {
			cidr = append(cidr, aws.StringValue(ipRange.CidrIp))
		}

		s.SecurityGroupGrants = append(s.SecurityGroupGrants, config.SecurityGroupGrant{
			Type:       "ingress",
			FromPort:   int(aws.Int64Value(grant.FromPort)),
			ToPort:     int(aws.Int64Value(grant.ToPort)),
			IPProtocol: aws.StringValue(grant.IpProtocol),
			CidrIP:     cidr,
		})
	}

	// Get the egress grants
	for _, grant := range securitygroup.IpPermissionsEgress {

		var cidr []string
		for _, ipRange := range grant.IpRanges {
			cidr = append(cidr, aws.StringValue(ipRange.CidrIp))
		}

		s.SecurityGroupGrants = append(s.SecurityGroupGrants, config.SecurityGroupGrant{
			Type:       "egress",
			FromPort:   int(aws.Int64Value(grant.FromPort)),
			ToPort:     int(aws.Int64Value(grant.ToPort)),
			IPProtocol: aws.StringValue(grant.IpProtocol),
			CidrIP:     cidr,
		})
	}

}

// PrintTable Prints an ascii table of the list of Security Groups
func (s *SecurityGroups) PrintTable() {
	if len(*s) == 0 {
		terminal.ShowErrorMessage("Warning", "No Security Groups Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*s))

	for index, sg := range *s {
		models.ExtractAwsmTable(index, sg, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

// CreateSecurityGroup creates a new Security Group based on the provided class, region, and VPC ID
func CreateSecurityGroup(class, region, vpcID string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Validate the region
	if !ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	// Verify the security group class input
	cfg, err := config.LoadSecurityGroupClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found Security Group class configuration for [" + class + "]")

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Create the security group
	params := &ec2.CreateSecurityGroupInput{
		Description: aws.String(cfg.Description),
		GroupName:   aws.String(class),
		DryRun:      aws.Bool(dryRun),
		VpcId:       aws.String(vpcID),
	}

	_, err = svc.CreateSecurityGroup(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil
}

// DeleteSecurityGroups deletes one or more Security Groups that match the provided search term and optional region
func DeleteSecurityGroups(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	secGrpList := new(SecurityGroups)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionSecurityGroups(region, secGrpList, search)
	} else {
		secGrpList, _ = GetSecurityGroups(search)
	}

	if err != nil {
		return errors.New("Error gathering Security Groups list")
	}

	if len(*secGrpList) > 0 {
		// Print the table
		secGrpList.PrintTable()
	} else {
		return errors.New("No Security Groups found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Security Groups?") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = deleteSecurityGroups(secGrpList, dryRun)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Done!")

	return nil
}

// UpdateSecurityGroups updates one or more Security Groups that match the provided search term and optional region
func UpdateSecurityGroups(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	secGrpList := new(SecurityGroups)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionSecurityGroups(region, secGrpList, search)
	} else {
		secGrpList, _ = GetSecurityGroups(search)
	}

	if err != nil {
		return errors.New("Error gathering Security Groups list")
	}

	if len(*secGrpList) > 0 {
		// Print the table
		secGrpList.PrintTable()
	} else {
		return errors.New("No Security Groups found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to update these Security Groups?") {
		return errors.New("Aborting!")
	}

	// Update 'Em
	err = updateSecurityGroups(secGrpList, dryRun)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Done!")

	return nil
}

// private function without terminal prompts
func updateSecurityGroups(secGrpList *SecurityGroups, dryRun bool) error {

	for _, secGrp := range *secGrpList {
		// Verify the security group class input
		cfg, err := config.LoadSecurityGroupClass(secGrp.Class)
		if err != nil {
			terminal.Information("Skipping Security Group [" + secGrp.Name + "]")
			terminal.ErrorLine(err.Error())
			continue
		} else {
			terminal.Information("Found Security Group class configuration for [" + secGrp.Class + "]")
		}

		// TODO
		fmt.Println("\n\n")
		fmt.Println("aws\n")
		fmt.Println(secGrp.SecurityGroupGrants)
		fmt.Println("awsm\n")
		fmt.Println(cfg.SecurityGroupGrants)
		fmt.Println("\n\n")

		hash1, err := hashstructure.Hash(secGrp.SecurityGroupGrants, nil)
		if err != nil {
			panic(err)
		}

		fmt.Printf("\n\n%d\n\n", hash1)

		hash2, err := hashstructure.Hash(cfg.SecurityGroupGrants, nil)
		if err != nil {
			panic(err)
		}

		fmt.Printf("\n\n%d\n\n", hash2)

	}

	return nil
}

// TODO
func authorizeIngress() {

}

// TODO
func authorizeEgress() {

}

// TODO
func revokeIngress() {

}

// TODO
func revokeEgress() {

}

// private function without terminal prompts
func deleteSecurityGroups(secGrpList *SecurityGroups, dryRun bool) error {

	for _, secGrp := range *secGrpList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(secGrp.Region)}))

		params := &ec2.DeleteSecurityGroupInput{
			DryRun:  aws.Bool(true),
			GroupId: aws.String(secGrp.GroupID),
		}

		_, err := svc.DeleteSecurityGroup(params)

		if err != nil {
			return err
		}

		terminal.Information("Deleted Security Group [" + secGrp.Name + "] in [" + secGrp.Region + "]!")
	}
	return nil

}
