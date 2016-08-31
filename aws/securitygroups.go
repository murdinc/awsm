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
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type SecurityGroups []SecurityGroup

type SecurityGroup struct {
	Name        string
	Class       string
	GroupId     string
	Description string
	Vpc         string
	VpcId       string
	Region      string
	SecuirityGroupPermissions
}

type SecuirityGroupPermissions struct {
	Ingress []*ec2.IpPermission
	Egress  []*ec2.IpPermission
}

func (s *SecurityGroups) GetSecurityGroupNames(ids []string) []string {
	names := make([]string, len(ids))
	for i, id := range ids {
		for _, secGrp := range *s {
			if secGrp.GroupId == id && secGrp.Name != "" {
				names[i] = secGrp.Name
			} else if secGrp.GroupId == id {
				names[i] = secGrp.GroupId
			}
		}
	}
	return names
}

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

func (s *SecurityGroup) Marshal(securitygroup *ec2.SecurityGroup, region string, vpcList *Vpcs) {

	vpc := vpcList.GetVpcName(aws.StringValue(securitygroup.VpcId))

	s.Name = aws.StringValue(securitygroup.GroupName)
	s.Class = GetTagValue("Class", securitygroup.Tags)
	s.GroupId = aws.StringValue(securitygroup.GroupId)
	s.Description = aws.StringValue(securitygroup.Description)
	s.Vpc = vpc
	s.VpcId = aws.StringValue(securitygroup.VpcId)
	s.SecuirityGroupPermissions.Ingress = securitygroup.IpPermissions
	s.SecuirityGroupPermissions.Egress = securitygroup.IpPermissionsEgress
	s.Region = region
}

func (i *SecurityGroups) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.Class,
			val.GroupId,
			val.Description,
			val.Vpc,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Class", "Group Id", "Description", "Vpc", "Region"})

	table.AppendBulk(rows)
	table.Render()
}

func CreateSecurityGroup(class, region, vpcId string, dryRun bool) error {

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
	} else {
		terminal.Information("Found Security Group class configuration for [" + class + "]")
	}

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Create the security group
	params := &ec2.CreateSecurityGroupInput{
		Description: aws.String(cfg.Description),
		GroupName:   aws.String(class),
		DryRun:      aws.Bool(dryRun),
		VpcId:       aws.String(vpcId),
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

		fmt.Println(secGrp)
		fmt.Println(cfg)

	}

	return nil
}

func authorizeIngress() {

}

func authorizeEgress() {

}

func revokeIngress() {

}

func revokeEgress() {

}

func deleteSecurityGroups(secGrpList *SecurityGroups, dryRun bool) error {

	for _, secGrp := range *secGrpList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(secGrp.Region)}))

		params := &ec2.DeleteSecurityGroupInput{
			DryRun:  aws.Bool(true),
			GroupId: aws.String(secGrp.GroupId),
		}

		_, err := svc.DeleteSecurityGroup(params)

		if err != nil {
			return err
		}

		terminal.Information("Deleted Security Group [" + secGrp.Name + "] in [" + secGrp.Region + "]!")
	}
	return nil

}
