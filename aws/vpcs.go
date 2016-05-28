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
	"github.com/murdinc/awsm/terminal"
	"github.com/olekukonko/tablewriter"
)

type Vpcs []Vpc

type Vpc struct {
	Name      string
	Class     string
	VpcId     string
	State     string
	Default   string
	CIDRBlock string
	DHCPOptId string
	Tenancy   string
	Region    string
}

func GetVpcByName(region, search string) (Vpc, error) {
	vpcList := new(Vpcs)
	err := GetRegionVpcs(region, vpcList, search)
	if err != nil {
		return Vpc{}, err
	}

	count := len(*vpcList)

	switch count {
	case 0:
		return Vpc{}, errors.New("No VPC found, Aborting!")
	case 1:
		vpcs := *vpcList
		return vpcs[0], nil
	}

	return Vpc{}, errors.New("Please limit your search term to return only one VPC")
}

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

func GetRegionVpcs(region string, vpcList *Vpcs, search string) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{})

	if err != nil {
		return err
	}

	vpcs := make(Vpcs, len(result.Vpcs))
	for i, vpc := range result.Vpcs {
		vpcs[i] = Vpc{
			Name:      GetTagValue("Name", vpc.Tags),
			Class:     GetTagValue("Class", vpc.Tags),
			VpcId:     aws.StringValue(vpc.VpcId),
			State:     aws.StringValue(vpc.State),
			Default:   fmt.Sprintf("%t", aws.BoolValue(vpc.IsDefault)),
			CIDRBlock: aws.StringValue(vpc.CidrBlock),
			DHCPOptId: aws.StringValue(vpc.DhcpOptionsId),
			Tenancy:   aws.StringValue(vpc.InstanceTenancy),
			Region:    region,
		}
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

func (i *Vpcs) GetVpcName(id string) string {
	for _, vpc := range *i {
		if vpc.VpcId == id && vpc.Name != "" {
			return vpc.Name
		}
	}
	return id
}

func (i *Vpcs) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.Class,
			val.VpcId,
			val.State,
			val.Default,
			val.CIDRBlock,
			val.DHCPOptId,
			val.Tenancy,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Class", "VPC Id", "State", "Default", "CIDR Block", "DHCP Options ID", "Tenancy", "Region"})

	table.AppendBulk(rows)
	table.Render()
}

func CreateVpc(class, name, ip, region string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	var cfg config.VpcClassConfig
	err := cfg.LoadConfig(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found VPC Class Configuration for [" + class + "]!")
	}

	// Validate the region
	if !ValidateRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

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
	vpcTagsParams := &ec2.CreateTagsInput{
		Resources: []*string{
			createVpcResp.Vpc.VpcId,
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
			{
				Key:   aws.String("Class"),
				Value: aws.String(class),
			},
		},
		DryRun: aws.Bool(dryRun),
	}
	_, err = svc.CreateTags(vpcTagsParams)

	if err != nil {
		return err
	}

	return nil

}

// Public function with confirmation terminal prompt
func DeleteVpcs(name, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	vpcList := new(Vpcs)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionVpcs(region, vpcList, name)
	} else {
		vpcList, _ = GetVpcs(name)
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
	if !terminal.PromptBool("Are you sure you want to delete these VPCs") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = deleteVpcs(vpcList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func deleteVpcs(vpcList *Vpcs, dryRun bool) (err error) {
	for _, vpc := range *vpcList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(vpc.Region)}))

		params := &ec2.DeleteVpcInput{
			VpcId:  aws.String(vpc.VpcId),
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
