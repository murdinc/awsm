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
	"github.com/murdinc/awsm/aws/regions"
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

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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
	regions := regions.GetRegionList()

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

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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
func CreateSecurityGroup(class, region, vpc string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Verify the security group class input
	cfg, err := config.LoadSecurityGroupClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found Security Group class configuration for [" + class + "]")

	// Validate the region
	if !regions.ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	params := &ec2.CreateSecurityGroupInput{
		Description: aws.String(cfg.Description),
		GroupName:   aws.String(class),
		DryRun:      aws.Bool(dryRun),
	}

	vpcList := new(Vpcs)
	if vpc != "" {
		GetRegionVpcs(region, vpcList, vpc)

		count := len(*vpcList)
		if count == 0 {
			return errors.New("No VPC's found matching [" + vpc + "] in [" + region + "], Aborting!")
		}

		if count > 1 {
			vpcList.PrintTable()
			return errors.New("Please limit your Vpc search term to result in only one VPC, Aborting!")
		}

		vpc := (*vpcList)[0]
		terminal.Information("Found VPC [" + vpc.VpcID + "] named [" + vpc.Name + "] in [" + region + "]")
		params.VpcId = aws.String(vpc.VpcID)
	}

	// Create the security group
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	createSecGrpResponse, err := svc.CreateSecurityGroup(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	// Add Tags
	SetEc2NameAndClassTags(createSecGrpResponse.GroupId, class, class, region)
	terminal.Delta("Created Security Group [" + aws.StringValue(createSecGrpResponse.GroupId) + "] in region [" + region + "]")

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
			terminal.Notice("Skipping Security Group [" + secGrp.Name + "]")
			terminal.ErrorLine(err.Error())
			continue
		} else {
			terminal.Information("Found Security Group class configuration for [" + secGrp.Class + "]")
		}

		// cycle through the config grants an generate hashes
		cfgHashes := make(map[uint64]int)
		for i, cGrant := range cfg.SecurityGroupGrants {
			configGrantHash, err := hashstructure.Hash(cGrant, nil)
			if err != nil {
				return err
			}
			cfgHashes[configGrantHash] = i
		}

		var remove, add []config.SecurityGroupGrant

		// cycle through existing grants and find ones to remove
		for _, eGrant := range secGrp.SecurityGroupGrants {

			existingGrantHash, err := hashstructure.Hash(eGrant, nil)
			if err != nil {
				return err
			}

			if _, ok := cfgHashes[existingGrantHash]; !ok {
				fmt.Println("remove")
				fmt.Println(existingGrantHash)
				fmt.Println(eGrant)
				fmt.Println("")

				//cfgHashes[existingGrantHash] = -1

				remove = append(remove, eGrant)
			} else {
				fmt.Println("keep")
				fmt.Println(existingGrantHash)
				fmt.Println(eGrant)
				fmt.Println("")

				delete(cfgHashes, existingGrantHash)
			}
		}

		// cycle through hashes and find ones to add
		for hash, i := range cfgHashes {
			fmt.Println("add")
			fmt.Println(hash)
			fmt.Println(i)
			fmt.Println(cfg.SecurityGroupGrants[i])
			fmt.Println("")
			add = append(add, cfg.SecurityGroupGrants[i])
		}

		fmt.Println("=======================================================")

		fmt.Println("remove:")
		fmt.Println(remove)
		fmt.Println("")

		fmt.Println("add:")
		fmt.Println(add)
		fmt.Println("")

	}

	return nil
}

// TODO
func authorizeIngress(dryRun bool) {

	params := &ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:   aws.String("String"),
		DryRun:   aws.Bool(dryRun),
		FromPort: aws.Int64(1),
		GroupId:  aws.String("String"),
		//GroupName: aws.String("String"),
		IpPermissions: []*ec2.IpPermission{
			{ // Required
				FromPort:   aws.Int64(1),
				IpProtocol: aws.String("String"),
				IpRanges: []*ec2.IpRange{
					{ // Required
						CidrIp: aws.String("String"),
					},
					// More values...
				},
				Ipv6Ranges: []*ec2.Ipv6Range{
					{ // Required
						CidrIpv6: aws.String("String"),
					},
					// More values...
				},
				PrefixListIds: []*ec2.PrefixListId{
					{ // Required
						PrefixListId: aws.String("String"),
					},
					// More values...
				},
				ToPort: aws.Int64(1),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{ // Required
						GroupId:       aws.String("String"),
						GroupName:     aws.String("String"),
						PeeringStatus: aws.String("String"),
						UserId:        aws.String("String"),
						VpcId:         aws.String("String"),
						VpcPeeringConnectionId: aws.String("String"),
					},
					// More values...
				},
			},
			// More values...
		},
		IpProtocol:                 aws.String("String"),
		SourceSecurityGroupName:    aws.String("String"),
		SourceSecurityGroupOwnerId: aws.String("String"),
		ToPort: aws.Int64(1),
	}

	sess := session.Must(session.NewSession())
	svc := ec2.New(sess)

	resp, err := svc.AuthorizeSecurityGroupIngress(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
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
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(secGrp.Region)}))
		svc := ec2.New(sess)

		params := &ec2.DeleteSecurityGroupInput{
			DryRun:  aws.Bool(dryRun),
			GroupId: aws.String(secGrp.GroupID),
		}

		_, err := svc.DeleteSecurityGroup(params)

		if err != nil {
			return err
		}

		terminal.Delta("Deleted Security Group [" + secGrp.Name + "] in [" + secGrp.Region + "]!")
	}
	return nil

}
