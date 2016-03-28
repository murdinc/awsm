package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/cli"
)

type SecurityGroups []SecurityGroup

type SecurityGroup struct {
	Name        string
	GroupId     string
	Description string
	Vpc         string
	Region      string
}

func GetSecurityGroups() (*SecurityGroups, error) {
	var wg sync.WaitGroup

	sgroupList := new(SecurityGroups)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSecurityGroups(region.RegionName, sgroupList)
			if err != nil {
				cli.ShowErrorMessage("Error gathering SecurityGroup list", err.Error())
			}
		}(region)
	}
	wg.Wait()

	return sgroupList, nil
}

func GetRegionSecurityGroups(region *string, sgroupList *SecurityGroups) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})

	if err != nil {
		return err
	}

	sgroup := make(SecurityGroups, len(result.SecurityGroups))
	for i, securitygroup := range result.SecurityGroups {
		sgroup[i] = SecurityGroup{
			Name:        GetTagValue("Name", securitygroup.Tags),
			GroupId:     aws.StringValue(securitygroup.GroupId),
			Description: aws.StringValue(securitygroup.Description),
			Vpc:         aws.StringValue(securitygroup.VpcId),
			Region:      fmt.Sprintf(*region),
		}
	}
	*sgroupList = append(*sgroupList, sgroup[:]...)

	return nil
}

func (i *SecurityGroups) PrintTable() {
	collumns := []string{"Name", "Group Id", "Description", "Vpc", "Region"}

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

	printTable(collumns, rows)
}
