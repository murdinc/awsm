package aws

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/terminal"
	"github.com/olekukonko/tablewriter"
)

type SecurityGroups []SecurityGroup

type SecurityGroup struct {
	Name        string
	GroupId     string
	Description string
	Vpc         string
	Region      string
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
		sec := new(SecurityGroup)
		sec.Marshall(result.SecurityGroups[0], region)
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

func GetSecurityGroups() (*SecurityGroups, []error) {
	var wg sync.WaitGroup
	var errs []error

	sgroupList := new(SecurityGroups)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSecurityGroups(*region.RegionName, sgroupList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering security group list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return sgroupList, errs
}

func (s *SecurityGroup) Marshall(securitygroup *ec2.SecurityGroup, region string) {

	s.Name = aws.StringValue(securitygroup.GroupName)
	s.GroupId = aws.StringValue(securitygroup.GroupId)
	s.Description = aws.StringValue(securitygroup.Description)
	s.Vpc = aws.StringValue(securitygroup.VpcId)
	s.Region = region
}

func GetRegionSecurityGroups(region string, sgroupList *SecurityGroups) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})

	if err != nil {
		return err
	}

	sgroup := make(SecurityGroups, len(result.SecurityGroups))
	for i, securitygroup := range result.SecurityGroups {
		sgroup[i].Marshall(securitygroup, region)
	}
	*sgroupList = append(*sgroupList, sgroup[:]...)

	return nil
}

func (i *SecurityGroups) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.GroupId,
			val.Description,
			val.Vpc,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Group Id", "Description", "Vpc", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
