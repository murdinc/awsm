package aws

import (
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator"
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

// GetSecurityGroupNames returns a slice of the security group names (or ID's if a name is not available)
func (s *SecurityGroups) GetSecurityGroupIDs() []string {
	ids := make([]string, len(*s))
	for i, secGrp := range *s {
		ids[i] = secGrp.GroupID
	}

	return ids
}

// GetSecurityGroupByName returns a single Security Group that matches a provided region and name
func GetSecurityGroupByName(region, name string) (SecurityGroup, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.DescribeSecurityGroupsInput{
		GroupNames: []*string{
			aws.String(name),
		},
	}

	result, err := svc.DescribeSecurityGroups(params)
	if err != nil {
		return SecurityGroup{}, err
	}

	count := len(result.SecurityGroups)

	switch count {
	case 0:
		// Fall back to tag
		return GetSecurityGroupByTag(region, "Name", name)
	case 1:
		vpcList := new(Vpcs)
		GetRegionVpcs(region, vpcList, "")

		sec := new(SecurityGroup)
		sec.Marshal(result.SecurityGroups[0], region, vpcList)
		return *sec, nil
	}

	return SecurityGroup{}, errors.New("Found more than one Security Group named [" + name + "] in [" + region + "], Aborting!")
}

// GetClassicSecurityGroupByName returns a single Security Group that matches a provided region and name, (non-vpc only)
func GetEc2ClassicSecurityGroupByName(region, name string) (SecurityGroup, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.DescribeSecurityGroupsInput{
		GroupNames: []*string{
			aws.String(name),
		},
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(""),
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
		// Fall back to tag
		return GetSecurityGroupByTag(region, "Name", name)
	case 1:
		vpcList := new(Vpcs)
		GetRegionVpcs(region, vpcList, "")

		sec := new(SecurityGroup)
		sec.Marshal(result.SecurityGroups[0], region, vpcList)
		return *sec, nil
	}

	return SecurityGroup{}, errors.New("Found more than one Security Group named [" + name + "] in [" + region + "], Aborting!")
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

	s.Name = GetTagValue("Name", securitygroup.Tags)
	s.Class = GetTagValue("Class", securitygroup.Tags)
	s.GroupID = aws.StringValue(securitygroup.GroupId)
	s.Description = aws.StringValue(securitygroup.Description)
	s.Vpc = vpc
	s.VpcID = aws.StringValue(securitygroup.VpcId)
	s.Region = region

	// Fall back to the Group name
	if s.Name == "" {
		s.Name = aws.StringValue(securitygroup.GroupName)
	}

	// Get the ingress grants
	for _, grant := range securitygroup.IpPermissions {

		var cidr, group []string
		for _, ipRange := range grant.IpRanges {
			cidr = append(cidr, aws.StringValue(ipRange.CidrIp))
		}

		for _, groupPairs := range grant.UserIdGroupPairs {
			group = append(group, aws.StringValue(groupPairs.GroupName))
		}

		s.SecurityGroupGrants = append(s.SecurityGroupGrants, config.SecurityGroupGrant{
			Type:                     "ingress",
			FromPort:                 int(aws.Int64Value(grant.FromPort)),
			ToPort:                   int(aws.Int64Value(grant.ToPort)),
			IPProtocol:               aws.StringValue(grant.IpProtocol),
			CidrIPs:                  cidr,
			SourceSecurityGroupNames: group,
		})
	}

	// Get the egress grants
	for _, grant := range securitygroup.IpPermissionsEgress {

		var cidr, group []string
		for _, ipRange := range grant.IpRanges {
			cidr = append(cidr, aws.StringValue(ipRange.CidrIp))
		}

		for _, groupPairs := range grant.UserIdGroupPairs {
			group = append(group, aws.StringValue(groupPairs.GroupName))
		}

		s.SecurityGroupGrants = append(s.SecurityGroupGrants, config.SecurityGroupGrant{
			Type:                     "egress",
			FromPort:                 int(aws.Int64Value(grant.FromPort)),
			ToPort:                   int(aws.Int64Value(grant.ToPort)),
			IPProtocol:               aws.StringValue(grant.IpProtocol),
			CidrIPs:                  cidr,
			SourceSecurityGroupNames: group,
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
	cfg, err := config.LoadSecurityGroupClass(class, false)
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

		theVpc := (*vpcList)[0]
		terminal.Information("Found VPC [" + theVpc.VpcID + "] named [" + theVpc.Name + "] in [" + region + "]")
		params.VpcId = aws.String(theVpc.VpcID)
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

	// Add Grants
	group := SecurityGroup{
		GroupID: aws.StringValue(createSecGrpResponse.GroupId),
		Class:   class,
		Name:    class,
		Region:  region,
		VpcID:   aws.StringValue(params.VpcId),
	}

	if vpc != "" {
		defaultGrants := []config.SecurityGroupGrant{
			config.SecurityGroupGrant{
				Type:       "egress",
				CidrIPs:    []string{"0.0.0.0/0"},
				IPProtocol: "-1",
				FromPort:   0,
				ToPort:     0,
			},
		}
		group.SecurityGroupGrants = defaultGrants
	}

	groupSlice := make(SecurityGroups, 1)
	groupSlice[0] = group
	changes, err := groupSlice.Diff()
	if err != nil {
		return err
	}

	return updateSecurityGroups(changes, dryRun)
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

	changes, err := secGrpList.Diff()
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		terminal.Information("There are no changes needed on these security groups!")
		return nil
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to update these Security Groups?") {
		return errors.New("Aborting!")
	}

	// Update 'Em
	err = updateSecurityGroups(changes, dryRun)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil
}

type SecurityGroupChange struct {
	Group  SecurityGroup
	Revoke bool
	Type   string
	Grants []config.SecurityGroupGrant
}

func (s SecurityGroups) Diff() ([]SecurityGroupChange, error) {

	terminal.Delta("Comparing awsm Security Group grants...")

	changes := []SecurityGroupChange{}
	cfgHashes := make([]map[uint64]config.SecurityGroupGrant, len(s))

	for i, secGrp := range s {
		cfgHashes[i] = make(map[uint64]config.SecurityGroupGrant)
		// Verify the security group class input
		cfg, err := config.LoadSecurityGroupClass(secGrp.Class, true)
		if err != nil {
			return changes, err
		}

		// cycle through the config grants and generate hashess
		for _, cGrant := range cfg.SecurityGroupGrants {

			for _, ipGrant := range cGrant.CidrIPs {
				cidrIpGrant := cGrant
				cidrIpGrant.CidrIPs = []string{ipGrant}
				cidrIpGrant.SourceSecurityGroupNames = []string{}

				configGrantHash, err := hashstructure.Hash(cidrIpGrant, nil)
				if err != nil {
					return changes, err
				}
				cfgHashes[i][configGrantHash] = cidrIpGrant
			}

			for _, secGrant := range cGrant.SourceSecurityGroupNames {
				secGrpGrant := cGrant
				secGrpGrant.CidrIPs = []string{}
				secGrpGrant.SourceSecurityGroupNames = []string{secGrant}

				configGrantHash, err := hashstructure.Hash(secGrpGrant, nil)
				if err != nil {
					return changes, err
				}
				cfgHashes[i][configGrantHash] = secGrpGrant
			}

		}

	}

	for i, secGrp := range s {

		var removeIngress, removeEgress, addIngress, addEgress []config.SecurityGroupGrant

		// cycle through existing grants and find ones to remove
		for _, grant := range secGrp.SecurityGroupGrants {

			for _, ipGrant := range grant.CidrIPs {
				cidrIpGrant := grant
				cidrIpGrant.CidrIPs = []string{ipGrant}
				cidrIpGrant.SourceSecurityGroupNames = []string{}

				existingGrantHash, err := hashstructure.Hash(cidrIpGrant, nil)
				if err != nil {
					return changes, err
				}

				if _, ok := cfgHashes[i][existingGrantHash]; !ok {
					terminal.Delta(fmt.Sprintf("[%s %s] - Deauthorize - [%s]	[%s :%d-%d]	[%s]", secGrp.Name, secGrp.Region, cidrIpGrant.Type, cidrIpGrant.IPProtocol, cidrIpGrant.FromPort, cidrIpGrant.ToPort, strings.Join(cidrIpGrant.CidrIPs, ", ")))

					if cidrIpGrant.Type == "ingress" {
						removeIngress = append(removeIngress, cidrIpGrant)
					} else if cidrIpGrant.Type == "egress" {
						removeEgress = append(removeEgress, cidrIpGrant)
					}

				} else {
					//terminal.Notice(fmt.Sprintf("[%s %s] - Keeping - [%s]	[%s :%d-%d]	[%s]", secGrp.Name, secGrp.Region, cidrIpGrant.Type, cidrIpGrant.IPProtocol, cidrIpGrant.FromPort, cidrIpGrant.ToPort, strings.Join(cidrIpGrant.CidrIPs, ", ")))
					delete(cfgHashes[i], existingGrantHash)
				}
			}

			for _, secGrant := range grant.SourceSecurityGroupNames {
				secGrpGrant := grant
				secGrpGrant.CidrIPs = []string{}
				secGrpGrant.SourceSecurityGroupNames = []string{secGrant}

				existingGrantHash, err := hashstructure.Hash(secGrpGrant, nil)
				if err != nil {
					return changes, err
				}

				if _, ok := cfgHashes[i][existingGrantHash]; !ok {
					terminal.Delta(fmt.Sprintf("[%s %s] - Deauthorize - [%s]	[%s :%d-%d]	[%s]", secGrp.Name, secGrp.Region, secGrpGrant.Type, secGrpGrant.IPProtocol, secGrpGrant.FromPort, secGrpGrant.ToPort, strings.Join(secGrpGrant.SourceSecurityGroupNames, ", ")))

					if secGrpGrant.Type == "ingress" {
						removeIngress = append(removeIngress, secGrpGrant)
					} else if secGrpGrant.Type == "egress" {
						removeEgress = append(removeEgress, secGrpGrant)
					}

				} else {
					//terminal.Notice(fmt.Sprintf("[%s %s] - Keeping - [%s]	[%s :%d-%d]	[%s]", secGrp.Name, secGrp.Region, secGrpGrant.Type, secGrpGrant.IPProtocol, secGrpGrant.FromPort, secGrpGrant.ToPort, strings.Join(secGrpGrant.SourceSecurityGroupNames, ", ")))
					delete(cfgHashes[i], existingGrantHash)
				}

			}

		}

		// cycle through hashes and find ones to add
		for _, grant := range cfgHashes[i] {

			sGrant := grant

			// Skip egress rules on non vpc security groups
			if secGrp.VpcID == "" && grant.Type == "egress" {
				terminal.Notice(fmt.Sprintf("[%s %s] - Skip - [%s]	[%s :%d-%d]	Egress rules can only be applied to VPC Security Groups", secGrp.Name, secGrp.Region, sGrant.Type, sGrant.IPProtocol, sGrant.FromPort, sGrant.ToPort))
				continue
			}

			for _, ipGrant := range grant.CidrIPs {
				sGrant.CidrIPs = []string{ipGrant}
				sGrant.SourceSecurityGroupNames = []string{}

				terminal.Delta(fmt.Sprintf("[%s %s] - Authorize - [%s]	[%s :%d-%d]	[%s]", secGrp.Name, secGrp.Region, sGrant.Type, sGrant.IPProtocol, sGrant.FromPort, sGrant.ToPort, strings.Join(sGrant.CidrIPs, ", ")))

				if grant.Type == "ingress" {
					addIngress = append(addIngress, sGrant)
				} else if grant.Type == "egress" {
					addEgress = append(addEgress, sGrant)
				}
			}

			for _, secGrant := range grant.SourceSecurityGroupNames {

				sGrant.CidrIPs = []string{}
				sGrant.SourceSecurityGroupNames = []string{secGrant}

				terminal.Delta(fmt.Sprintf("[%s %s] - Authorize - [%s]	[%s :%d-%d]	[%s]", secGrp.Name, secGrp.Region, sGrant.Type, sGrant.IPProtocol, sGrant.FromPort, sGrant.ToPort, strings.Join(sGrant.SourceSecurityGroupNames, ", ")))

				if grant.Type == "ingress" {
					addIngress = append(addIngress, sGrant)
				} else if grant.Type == "egress" {
					addEgress = append(addEgress, sGrant)
				}
			}
		}

		// authorize
		if len(addIngress) > 0 {
			changes = append(changes, SecurityGroupChange{
				Group:  secGrp,
				Revoke: false,
				Type:   "ingress",
				Grants: addIngress,
			})
		}
		if len(addEgress) > 0 {
			changes = append(changes, SecurityGroupChange{
				Group:  secGrp,
				Revoke: false,
				Type:   "egress",
				Grants: addEgress,
			})
		}
		// deauthorize
		if len(removeIngress) > 0 {
			changes = append(changes, SecurityGroupChange{
				Group:  secGrp,
				Revoke: true,
				Type:   "ingress",
				Grants: removeIngress,
			})
		}

		if len(removeEgress) > 0 {
			changes = append(changes, SecurityGroupChange{
				Group:  secGrp,
				Revoke: true,
				Type:   "egress",
				Grants: removeEgress,
			})
		}
	}

	terminal.Information("Comparison complete!")

	return changes, nil

}

// private function without terminal prompts
func updateSecurityGroups(changes []SecurityGroupChange, dryRun bool) error {

	for _, change := range changes {
		if change.Type == "ingress" {
			if change.Revoke {
				// revoke
				err := revokeIngress(change.Group, change.Grants, dryRun)
				if err != nil {
					return err
				}
			} else {
				// authorize
				err := authorizeIngress(change.Group, change.Grants, dryRun)
				if err != nil {
					return err
				}
			}

		} else if change.Type == "egress" {
			if change.Revoke {
				// revoke
				err := revokeEgress(change.Group, change.Grants, dryRun)
				if err != nil {
					return err
				}
			} else {
				// authorize
				err := authorizeEgress(change.Group, change.Grants, dryRun)
				if err != nil {
					return err
				}
			}
		}
	}

	terminal.Information("Done!")

	return nil
}

func authorizeIngress(secGrp SecurityGroup, grants []config.SecurityGroupGrant, dryRun bool) error {

	if len(grants) == 0 {
		return nil
	}

	params := &ec2.AuthorizeSecurityGroupIngressInput{
		DryRun:  aws.Bool(dryRun),
		GroupId: aws.String(secGrp.GroupID),
		//GroupName: aws.String(secGrp.Name),
	}

	ipPermissions := []*ec2.IpPermission{}

	for _, grant := range grants {

		ipRanges := []*ec2.IpRange{}
		ipv6Ranges := []*ec2.Ipv6Range{}
		groupPairs := []*ec2.UserIdGroupPair{}

		for _, ip := range grant.CidrIPs {

			address, _, _ := net.ParseCIDR(ip)

			if govalidator.IsIPv4(address.String()) {
				ipRanges = append(ipRanges,
					&ec2.IpRange{
						CidrIp: aws.String(ip),
					},
				)
			} else if govalidator.IsIPv6(address.String()) {
				ipv6Ranges = append(ipv6Ranges,
					&ec2.Ipv6Range{
						CidrIpv6: aws.String(ip),
					},
				)
			} else {
				return errors.New("IP [" + ip + "] does not appear to be a valid IPv4 or IPv6 Address. Aborting!")
			}
		}

		for _, groupName := range grant.SourceSecurityGroupNames {
			_, err := GetSecurityGroupByName(secGrp.Region, groupName)
			if err != nil {
				return errors.New("Security Group [" + groupName + "] does not appear to exist in [" + secGrp.Region + "]. Aborting!")
			}

			groupPairs = append(groupPairs,
				&ec2.UserIdGroupPair{
					GroupName: aws.String(groupName),
				},
			)
		}

		ipPermission := &ec2.IpPermission{}

		ipPermission.
			SetIpProtocol(grant.IPProtocol).
			SetFromPort(int64(grant.FromPort)).
			SetToPort(int64(grant.ToPort))

		if len(ipRanges) > 0 {
			ipPermission.SetIpRanges(ipRanges)
		}

		if len(ipv6Ranges) > 0 {
			ipPermission.SetIpv6Ranges(ipv6Ranges)
		}

		if len(groupPairs) > 0 {
			ipPermission.SetUserIdGroupPairs(groupPairs)
		}

		ipPermissions = append(ipPermissions, ipPermission)
		params.SetIpPermissions(ipPermissions)

	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(secGrp.Region)}))
	svc := ec2.New(sess)
	_, err := svc.AuthorizeSecurityGroupIngress(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "DryRunOperation" {
				return nil
			}
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil
}

func authorizeEgress(secGrp SecurityGroup, grants []config.SecurityGroupGrant, dryRun bool) error {

	if len(grants) == 0 {
		return nil
	}

	params := &ec2.AuthorizeSecurityGroupEgressInput{
		DryRun:  aws.Bool(dryRun),
		GroupId: aws.String(secGrp.GroupID),
		//GroupName: aws.String(secGrp.Name),
	}

	ipPermissions := []*ec2.IpPermission{}

	for _, grant := range grants {

		ipRanges := []*ec2.IpRange{}
		ipv6Ranges := []*ec2.Ipv6Range{}
		groupPairs := []*ec2.UserIdGroupPair{}

		for _, ip := range grant.CidrIPs {

			address, _, _ := net.ParseCIDR(ip)

			if govalidator.IsIPv4(address.String()) {
				ipRanges = append(ipRanges,
					&ec2.IpRange{
						CidrIp: aws.String(ip),
					},
				)
			} else if govalidator.IsIPv6(address.String()) {
				ipv6Ranges = append(ipv6Ranges,
					&ec2.Ipv6Range{
						CidrIpv6: aws.String(ip),
					},
				)
			} else {
				return errors.New("IP [" + ip + "] does not appear to be a valid IPv4 or IPv6 Address. Aborting!")
			}
		}

		for _, groupName := range grant.SourceSecurityGroupNames {
			_, err := GetSecurityGroupByName(secGrp.Region, groupName)
			if err != nil {
				return errors.New("Security Group [" + groupName + "] does not appear to exist in [" + secGrp.Region + "]. Aborting!")
			}

			groupPairs = append(groupPairs,
				&ec2.UserIdGroupPair{
					GroupName: aws.String(groupName),
				},
			)
		}

		ipPermission := &ec2.IpPermission{}

		ipPermission.
			SetIpProtocol(grant.IPProtocol).
			SetFromPort(int64(grant.FromPort)).
			SetToPort(int64(grant.ToPort))

		if len(ipRanges) > 0 {
			ipPermission.SetIpRanges(ipRanges)
		}

		if len(ipv6Ranges) > 0 {
			ipPermission.SetIpv6Ranges(ipv6Ranges)
		}

		if len(groupPairs) > 0 {
			ipPermission.SetUserIdGroupPairs(groupPairs)
		}

		ipPermissions = append(ipPermissions, ipPermission)
		params.SetIpPermissions(ipPermissions)

	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(secGrp.Region)}))
	svc := ec2.New(sess)
	_, err := svc.AuthorizeSecurityGroupEgress(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "DryRunOperation" {
				return nil
			}
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil
}

func revokeIngress(secGrp SecurityGroup, grants []config.SecurityGroupGrant, dryRun bool) error {

	if len(grants) == 0 {
		return nil
	}

	params := &ec2.RevokeSecurityGroupIngressInput{
		DryRun:  aws.Bool(dryRun),
		GroupId: aws.String(secGrp.GroupID),
		//GroupName: aws.String(secGrp.Name),
	}

	ipPermissions := []*ec2.IpPermission{}

	for _, grant := range grants {

		ipRanges := []*ec2.IpRange{}
		ipv6Ranges := []*ec2.Ipv6Range{}
		groupPairs := []*ec2.UserIdGroupPair{}

		for _, ip := range grant.CidrIPs {

			address, _, _ := net.ParseCIDR(ip)

			if govalidator.IsIPv4(address.String()) {
				ipRanges = append(ipRanges,
					&ec2.IpRange{
						CidrIp: aws.String(ip),
					},
				)
			} else if govalidator.IsIPv6(address.String()) {
				ipv6Ranges = append(ipv6Ranges,
					&ec2.Ipv6Range{
						CidrIpv6: aws.String(ip),
					},
				)
			} else {
				return errors.New("IP [" + ip + "] does not appear to be a valid IPv4 or IPv6 Address. Aborting!")
			}
		}

		for _, groupName := range grant.SourceSecurityGroupNames {
			_, err := GetSecurityGroupByName(secGrp.Region, groupName)
			if err != nil {
				return errors.New("Security Group [" + groupName + "] does not appear to exist in [" + secGrp.Region + "]. Aborting!")
			}

			groupPairs = append(groupPairs,
				&ec2.UserIdGroupPair{
					GroupName: aws.String(groupName),
				},
			)
		}

		ipPermission := &ec2.IpPermission{}

		ipPermission.
			SetIpProtocol(grant.IPProtocol).
			SetFromPort(int64(grant.FromPort)).
			SetToPort(int64(grant.ToPort))

		if len(ipRanges) > 0 {
			ipPermission.SetIpRanges(ipRanges)
		}

		if len(ipv6Ranges) > 0 {
			ipPermission.SetIpv6Ranges(ipv6Ranges)
		}

		if len(groupPairs) > 0 {
			ipPermission.SetUserIdGroupPairs(groupPairs)
		}

		ipPermissions = append(ipPermissions, ipPermission)
		params.SetIpPermissions(ipPermissions)

	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(secGrp.Region)}))
	svc := ec2.New(sess)
	_, err := svc.RevokeSecurityGroupIngress(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "DryRunOperation" {
				return nil
			}
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil
}

func revokeEgress(secGrp SecurityGroup, grants []config.SecurityGroupGrant, dryRun bool) error {

	if len(grants) == 0 {
		return nil
	}

	params := &ec2.RevokeSecurityGroupEgressInput{
		DryRun:  aws.Bool(dryRun),
		GroupId: aws.String(secGrp.GroupID),
		//GroupName: aws.String(secGrp.Name),
	}

	ipPermissions := []*ec2.IpPermission{}

	for _, grant := range grants {

		ipRanges := []*ec2.IpRange{}
		ipv6Ranges := []*ec2.Ipv6Range{}
		groupPairs := []*ec2.UserIdGroupPair{}

		for _, ip := range grant.CidrIPs {

			address, _, _ := net.ParseCIDR(ip)

			if govalidator.IsIPv4(address.String()) {
				ipRanges = append(ipRanges,
					&ec2.IpRange{
						CidrIp: aws.String(ip),
					},
				)
			} else if govalidator.IsIPv6(address.String()) {
				ipv6Ranges = append(ipv6Ranges,
					&ec2.Ipv6Range{
						CidrIpv6: aws.String(ip),
					},
				)
			} else {
				return errors.New("IP [" + ip + "] does not appear to be a valid IPv4 or IPv6 Address. Aborting!")
			}
		}

		for _, groupName := range grant.SourceSecurityGroupNames {
			_, err := GetSecurityGroupByName(secGrp.Region, groupName)
			if err != nil {
				return errors.New("Security Group [" + groupName + "] does not appear to exist in [" + secGrp.Region + "]. Aborting!")
			}

			groupPairs = append(groupPairs,
				&ec2.UserIdGroupPair{
					GroupName: aws.String(groupName),
				},
			)
		}

		ipPermission := &ec2.IpPermission{}

		ipPermission.
			SetIpProtocol(grant.IPProtocol).
			SetFromPort(int64(grant.FromPort)).
			SetToPort(int64(grant.ToPort))

		if len(ipRanges) > 0 {
			ipPermission.SetIpRanges(ipRanges)
		}

		if len(ipv6Ranges) > 0 {
			ipPermission.SetIpv6Ranges(ipv6Ranges)
		}

		if len(groupPairs) > 0 {
			ipPermission.SetUserIdGroupPairs(groupPairs)
		}

		ipPermissions = append(ipPermissions, ipPermission)
		params.SetIpPermissions(ipPermissions)

	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(secGrp.Region)}))
	svc := ec2.New(sess)
	_, err := svc.RevokeSecurityGroupEgress(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "DryRunOperation" {
				return nil
			}
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil
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
