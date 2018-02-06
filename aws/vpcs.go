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
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// Vpcs represents a slice of VPCs
type Vpcs []Vpc

// Vpc represents a single VPC
type Vpc models.Vpc

// InternetGateways represents a slice of InternetGateways
type InternetGateways []InternetGateway

// InternetGateway represents a single InternetGateway
type InternetGateway models.InternetGateway

// RouteTables represents a slice of RouteTables
type RouteTables []RouteTable

// RouteTable represents a single RouteTable
type RouteTable models.RouteTable

// RouteTableAssociations represents a slice of RouteTableAssociations
type RouteTableAssociations []RouteTableAssociation

// RouteTableAssociation represents a single RouteTableAssociation
type RouteTableAssociation models.RouteTableAssociation

// GetRegionVpcByTag returns a single VPC that matches the provided region and Tag key/value
func GetRegionVpcByTag(region, key, value string) (Vpc, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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

// GetVpcByTag returns a single VPC that matches the provided and Tag key/value
func GetVpcByTag(region, key, value string) (Vpc, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(v.Region)}))
	svc := ec2.New(sess)

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

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(v.Region)}))
	svc := ec2.New(sess)

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
	regions := GetRegionListWithoutIgnored()

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

func AttachInternetGateway(gatewaySearch, vpcSearch string, dryRun bool) error {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the Internet Gateway
	gatewayList, errs := GetInternetGateways(gatewaySearch, true)
	if errs != nil {
		return errors.New("Error while trying to find the internet gateway!")
	}

	gatewayCount := len(*gatewayList)

	if gatewayCount < 1 {
		return errors.New("No available Internet Gateways found matching [" + gatewaySearch + "]")
	} else if gatewayCount > 1 {
		gatewayList.PrintTable()
		return errors.New("Please limit your search term to just one Internet Gateway")
	}

	gateway := (*gatewayList)[0]

	terminal.Information("Found Internet Gateway [" + gateway.InternetGatewayID + "] named [" + gateway.Name + "] in [" + gateway.Region + "]!")

	// Get the VPC

	vpcList := new(Vpcs)
	GetRegionVpcs(gateway.Region, vpcList, vpcSearch)

	vpcCount := len(*vpcList)
	if vpcCount == 0 {
		return errors.New("No VPCs found for your search terms.")
	}
	if vpcCount > 1 {
		vpcList.PrintTable()
		return errors.New("Please limit your search to return only one VPC.")
	}
	vpc := (*vpcList)[0]

	terminal.Information("Found VPC [" + vpc.VpcID + "] named [" + vpc.Name + "] with a class of [" + vpc.Class + "] in [" + vpc.Region + "]!")

	// Confirm
	if !terminal.PromptBool("Are you sure you want to attach this Internet Gateway to this VPC?") {
		terminal.ErrorLine("Aborting!")
		return nil
	}

	// Attach it!
	err := attachInternetGateway(vpc.VpcID, gateway.InternetGatewayID, vpc.Region, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

func attachInternetGateway(vpcId, internetGatewayId, region string, dryRun bool) error {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(internetGatewayId),
		VpcId:             aws.String(vpcId),
		DryRun:            aws.Bool(dryRun),
	}

	_, err := svc.AttachInternetGateway(params)
	if err != nil {
		return err
	}

	terminal.Delta("Attached Internet Gateway [" + internetGatewayId + "] to VPC [" + vpcId + "]!")

	return nil
}

func AssociateRouteTable(routeTableSearch, subnetSearch string, dryRun bool) error {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the Subnet
	subnetList, _ := GetSubnets(subnetSearch)
	subCount := len(*subnetList)
	if subCount == 0 {
		return errors.New("No Subnets found for your search terms.")
	}
	if subCount > 1 {
		subnetList.PrintTable()
		return errors.New("Please limit your search to return only one Subnet.")
	}
	subnet := (*subnetList)[0]

	terminal.Information("Found Subnet [" + subnet.SubnetID + "] named [" + subnet.Name + "] with a class of [" + subnet.Class + "] in [" + subnet.Region + "]!")

	// Get the Route Table
	rtList, err := GetVpcRouteTables(routeTableSearch, subnet.VpcID, subnet.Region)
	if err != nil {
		return errors.New("Error while trying to find the route table!")
	}

	rtCount := len(*rtList)

	if rtCount < 1 {
		return errors.New("No Route Tables found matching [" + routeTableSearch + "]")
	} else if rtCount > 1 {
		rtList.PrintTable()
		return errors.New("Please limit your search term to just one Route Table")
	}

	rt := (*rtList)[0]

	terminal.Information("Found Route Table [" + rt.RouteTableID + "] named[" + rt.Name + "] in [" + rt.Region + "]!")

	// Confirm
	if !terminal.PromptBool("Are you sure you want to associate this Route Table to this Subnet?") {
		terminal.ErrorLine("Aborting!")
		return nil
	}

	// Associate it!
	associateRouteTable(rt.RouteTableID, subnet.SubnetID, subnet.Region, dryRun)

	terminal.Information("Done!")

	return nil
}

func associateRouteTable(rtId, subnetId, region string, dryRun bool) error {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(rtId),
		SubnetId:     aws.String(subnetId),
		DryRun:       aws.Bool(dryRun),
	}

	_, err := svc.AssociateRouteTable(params)
	if err != nil {
		return err
	}

	terminal.Delta("Associated Route Table [" + rtId + "] to Subnet [" + subnetId + "] !")

	return nil
}

func DetachInternetGateway(gatewaySearch string, dryRun bool) error {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the Internet Gateway
	gatewayList, err := GetInternetGateways(gatewaySearch, false)
	if err != nil {
		return errors.New("Error while trying to find the internet gateway!")
	}

	gatewayCount := len(*gatewayList)

	if gatewayCount < 1 {
		return errors.New("No Internet Gateways found matching [" + gatewaySearch + "]")
	} else if gatewayCount > 1 {
		gatewayList.PrintTable()
		return errors.New("Please limit your search term to just one Internet Gateway")
	}

	gateway := (*gatewayList)[0]

	terminal.Information("Found Internet Gateway [" + gateway.InternetGatewayID + "] named [" + gateway.Name + "] attached to [" + gateway.Attachment + "] in [" + gateway.Region + "]!")

	if gateway.Attachment == "" {
		terminal.ErrorLine("This Internet Gateway does not appear to be attached to a VPC! Aborting.")
		return nil
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to detach this Internet Gateway from this VPC?") {
		terminal.ErrorLine("Aborting!")
		return nil
	}

	// Detach it!
	detachInternetGateway(gateway, dryRun)

	terminal.Information("Done!")

	return nil
}

func detachInternetGateway(internetGateway InternetGateway, dryRun bool) error {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(internetGateway.Region)}))
	svc := ec2.New(sess)

	params := &ec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(internetGateway.InternetGatewayID),
		VpcId:             aws.String(internetGateway.Attachment),
		DryRun:            aws.Bool(dryRun),
	}

	_, err := svc.DetachInternetGateway(params)
	if err != nil {
		return err
	}

	terminal.Delta("Detached Internet Gateway [" + internetGateway.InternetGatewayID + "] named [" + internetGateway.Name + "] from VPC [" + internetGateway.Attachment + "]!")

	return nil
}

/**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/ /**/

func DisassociateRouteTable(routeTableSearch, subnetSearch string, dryRun bool) error {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the Subnet
	subnetList, _ := GetSubnets(subnetSearch)
	subCount := len(*subnetList)
	if subCount == 0 {
		return errors.New("No Subnets found for your search terms.")
	}
	if subCount > 1 {
		subnetList.PrintTable()
		return errors.New("Please limit your search to return only one Subnet.")
	}
	subnet := (*subnetList)[0]

	terminal.Information("Found Subnet [" + subnet.SubnetID + "] named [" + subnet.Name + "] with a class of [" + subnet.Class + "] in [" + subnet.Region + "]!")

	// Get the Route Table
	rtList, err := GetVpcRouteTables(routeTableSearch, subnet.VpcID, subnet.Region)
	if err != nil {
		return errors.New("Error while trying to find the route table!")
	}

	rtCount := len(*rtList)

	if rtCount < 1 {
		return errors.New("No Route Tables found matching [" + routeTableSearch + "]")
	} else if rtCount > 1 {
		rtList.PrintTable()
		return errors.New("Please limit your search term to just one Route Table")
	}

	rt := (*rtList)[0]

	terminal.Information("Found Route Table [" + rt.RouteTableID + "] named [" + rt.Name + "] in [" + rt.Region + "]!")

	associationId := ""

Loop:
	for _, association := range rt.Associations {
		if association.SubnetID == subnet.SubnetID {
			associationId = association.AssociationID
			break Loop
		}
	}

	if associationId == "" {
		return errors.New("Unable to locate the association id between this route table and subnet!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to disassociate this Route Table from this Subnet?") {
		terminal.ErrorLine("Aborting!")
		return nil
	}

	// Associate it!
	disassociateRouteTable(associationId, rt, subnet, dryRun)

	terminal.Information("Done!")

	return nil
}

func disassociateRouteTable(associationId string, rt RouteTable, subnet Subnet, dryRun bool) error {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(subnet.Region)}))
	svc := ec2.New(sess)

	params := &ec2.DisassociateRouteTableInput{
		AssociationId: aws.String(associationId),
		DryRun:        aws.Bool(dryRun),
	}

	_, err := svc.DisassociateRouteTable(params)
	if err != nil {
		return err
	}

	terminal.Delta("Disassociated Route Table [" + rt.RouteTableID + "] named [" + rt.Name + "] from Subnet [" + subnet.SubnetID + "] named [" + subnet.Name + "]!")

	return nil
}

func DeleteInternetGateway(gatewaySearch string, dryRun bool) error {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the Internet Gateway
	gatewayList, err := GetInternetGateways(gatewaySearch, false)
	if err != nil {
		return errors.New("Error while trying to find the internet gateway!")
	}

	gatewayCount := len(*gatewayList)

	if gatewayCount < 1 {
		return errors.New("No Internet Gateways found matching [" + gatewaySearch + "]")
	} else if gatewayCount > 1 {
		gatewayList.PrintTable()
		return errors.New("Please limit your search term to just one Internet Gateway")
	}

	gateway := (*gatewayList)[0]

	terminal.Information("Found Internet Gateway [" + gateway.InternetGatewayID + "] named [" + gateway.Name + "] in [" + gateway.Region + "]!")

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete this Internet Gateway?") {
		terminal.ErrorLine("Aborting!")
		return nil
	}

	// Delete it!
	deleteInternetGateway(gateway, dryRun)

	terminal.Information("Done!")

	return nil
}

func deleteInternetGateway(internetGateway InternetGateway, dryRun bool) error {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(internetGateway.Region)}))
	svc := ec2.New(sess)

	params := &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(internetGateway.InternetGatewayID),
		DryRun:            aws.Bool(dryRun),
	}

	_, err := svc.DeleteInternetGateway(params)
	if err != nil {
		return err
	}

	terminal.Delta("Deleted Internet Gateway [" + internetGateway.InternetGatewayID + "] named [" + internetGateway.Name + "]!")

	return nil
}

func DeleteRouteTable(routeTableSearch string, dryRun bool) error {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the Route Table
	rtList, err := GetRouteTables(routeTableSearch)
	if err != nil {
		return errors.New("Error while trying to find the route table!")
	}

	rtCount := len(*rtList)

	if rtCount < 1 {
		return errors.New("No Route Tables found matching [" + routeTableSearch + "]")
	} else if rtCount > 1 {
		rtList.PrintTable()
		return errors.New("Please limit your search term to just one Route Table")
	}

	rt := (*rtList)[0]

	terminal.Information("Found Route Table [" + rt.RouteTableID + "] named [" + rt.Name + "] in [" + rt.Region + "]!")

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete this Route Table?") {
		terminal.ErrorLine("Aborting!")
		return nil
	}

	// Delete it!
	deleteRouteTable(rt, dryRun)

	terminal.Information("Done!")

	return nil
}

func deleteRouteTable(rt RouteTable, dryRun bool) error {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(rt.Region)}))
	svc := ec2.New(sess)

	params := &ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(rt.RouteTableID),
		DryRun:       aws.Bool(dryRun),
	}

	_, err := svc.DeleteRouteTable(params)
	if err != nil {
		return err
	}

	terminal.Delta("Deleted Route Table [" + rt.RouteTableID + "] named [" + rt.Name + "]!")

	return nil
}

// GetInternetGateways returns a slice of Internet Gateways that match the provided search term
func GetInternetGateways(search string, available bool) (*InternetGateways, []error) {
	var wg sync.WaitGroup
	var errs []error

	igList := new(InternetGateways)
	regions := GetRegionListWithoutIgnored()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionInternetGateways(*region.RegionName, igList, search, available)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering internet gateway list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return igList, errs
}

// GetRegionInternetGateways returns a list of a regions InternetGateways that match the provided search term
func GetRegionInternetGateways(region string, igList *InternetGateways, search string, available bool) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	result, err := svc.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})

	if err != nil {
		return err
	}

	igs := make(InternetGateways, len(result.InternetGateways))
	for i, ig := range result.InternetGateways {
		igs[i].Marshal(ig, region)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, v := range igs {
			rIg := reflect.ValueOf(v)

			for k := 0; k < rIg.NumField(); k++ {
				sVal := rIg.Field(k).String()

				if term.MatchString(sVal) && ((available && igs[i].State == "detached") || !available) {
					*igList = append(*igList, igs[i])
					continue Loop
				}
			}
		}
	} else {
		if available {
			for i, _ := range igs {
				if igs[i].State == "detached" {
					*igList = append(*igList, igs[i])
				}
			}
		} else {
			*igList = append(*igList, igs[:]...)
		}
	}

	return nil
}

// Marshal parses the response from the aws sdk into an awsm Vpc
func (i *InternetGateway) Marshal(ig *ec2.InternetGateway, region string) {
	i.Name = GetTagValue("Name", ig.Tags)
	i.InternetGatewayID = aws.StringValue(ig.InternetGatewayId)
	i.State = "detached"
	i.Region = region

	if len(ig.Attachments) > 0 {
		i.State = "attached"
		i.Attachment = aws.StringValue(ig.Attachments[0].VpcId)
	}
}

func GetVpcMainRouteTable(vpcId, region string) (RouteTable, error) {

	// Get the Vpc Route Tables
	rtList, err := GetVpcRouteTables("", vpcId, region)
	if err != nil {
		return RouteTable{}, errors.New("Error while trying to find the route table!")
	}

	rtCount := len(*rtList)

	if rtCount < 1 {
		return RouteTable{}, errors.New("No Route Tables found!")
	}

	for _, rt := range *rtList {
		if rt.VpcID == vpcId {
			for _, assoc := range rt.Associations {
				if assoc.Main {
					return rt, nil
				}
			}
		}
	}

	return RouteTable{}, errors.New("Unable to locate the Main Route Table!")
}

func GetVpcRouteTables(search, vpcId, region string) (*RouteTables, error) {

	rtList := new(RouteTables)

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	result, err := svc.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(vpcId),
				},
			},
		},
	})
	if err != nil {
		return rtList, err
	}

	rts := make(RouteTables, len(result.RouteTables))
	for i, rt := range result.RouteTables {
		rts[i].Marshal(rt, region)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, v := range rts {
			rIg := reflect.ValueOf(v)

			for k := 0; k < rIg.NumField(); k++ {
				sVal := rIg.Field(k).String()

				if term.MatchString(sVal) {
					*rtList = append(*rtList, rts[i])
					continue Loop
				}
			}
		}
	} else {
		*rtList = append(*rtList, rts[:]...)
	}

	return rtList, nil
}

// GetRouteTables returns a slice of Route Tables that match the provided search term
func GetRouteTables(search string) (*RouteTables, []error) {
	var wg sync.WaitGroup
	var errs []error

	rtList := new(RouteTables)
	regions := GetRegionListWithoutIgnored()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionRouteTables(*region.RegionName, rtList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering route tables list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return rtList, errs
}

// GetRegionRouteTables returns a list of a regions RouteTables that match the provided search term
func GetRegionRouteTables(region string, rtList *RouteTables, search string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	result, err := svc.DescribeRouteTables(&ec2.DescribeRouteTablesInput{})

	if err != nil {
		return err
	}

	rts := make(RouteTables, len(result.RouteTables))
	for i, rt := range result.RouteTables {
		rts[i].Marshal(rt, region)
		/*
			mainRt, err := GetVpcMainRouteTable(*rt.VpcId, region)
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println(mainRt)
		*/
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, v := range rts {
			rIg := reflect.ValueOf(v)

			for k := 0; k < rIg.NumField(); k++ {
				sVal := rIg.Field(k).String()

				if term.MatchString(sVal) {
					*rtList = append(*rtList, rts[i])
					continue Loop
				}
			}
		}
	} else {
		*rtList = append(*rtList, rts[:]...)
	}

	return nil
}

// Marshal parses the response from the aws sdk into an awsm Vpc
func (r *RouteTable) Marshal(rt *ec2.RouteTable, region string) {
	r.Name = GetTagValue("Name", rt.Tags)
	r.RouteTableID = aws.StringValue(rt.RouteTableId)
	r.VpcID = aws.StringValue(rt.VpcId)
	r.Region = region
	for _, association := range rt.Associations {
		r.Associations = append(r.Associations, models.RouteTableAssociation{
			Main:          aws.BoolValue(association.Main),
			AssociationID: aws.StringValue(association.RouteTableAssociationId),
			SubnetID:      aws.StringValue(association.SubnetId),
		})
	}
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

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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

// PrintTable Prints an ascii table of the list of VPCs
func (i *InternetGateways) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Internet Gateways Found!")
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

// PrintTable Prints an ascii table of the list of VPCs
func (i *RouteTables) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Route Tables Found!")
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
	if !regions.ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	// TODO limit to one VPC of a class per region, so that we can target VPCs by class instead of name?

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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

	vpcId := *createVpcResp.Vpc.VpcId

	terminal.Delta("Created VPC [" + vpcId + "] named [" + name + "] in [" + region + "]!")

	terminal.Notice("Waiting to tag VPC...")

	err = svc.WaitUntilVpcExists(&ec2.DescribeVpcsInput{
		VpcIds: []*string{
			aws.String(vpcId),
		},
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
	}

	terminal.Delta("Adding VPC Tags...")

	// Add Tags
	err = SetEc2NameAndClassTags(&vpcId, name, class, region)
	if err != nil {
		return err
	}

	return nil

}

// CreateInternetGateway creates a new VPC Internet Gateway
func CreateInternetGateway(name, region string, dryRun bool) (string, error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Validate the region
	if !regions.ValidRegion(region) {
		return "", errors.New("Region [" + region + "] is Invalid!")
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	// Create the Internet Gateway
	params := &ec2.CreateInternetGatewayInput{
		DryRun: aws.Bool(dryRun),
	}

	createIGResp, err := svc.CreateInternetGateway(params)
	if err != nil {
		return "", err
	}

	gatewayId := *createIGResp.InternetGateway.InternetGatewayId
	terminal.Delta("Created VPC Internet Gateway [" + gatewayId + "] named [" + name + "] in [" + region + "]!")

	terminal.Delta("Adding Internet Gateway Tags...")

	// Tag it
	err = SetEc2NameAndClassTags(&gatewayId, name, "", region)
	if err != nil {
		return gatewayId, err
	}

	return gatewayId, nil
}

func CreateNatGateway(name, allocationId, subnetId, region string, dryRun bool) (string, error) {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	if allocationId == "" {
		allocationId, _ = CreateAddress(region, "vpc", dryRun)
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.CreateNatGatewayInput{
		AllocationId: aws.String(allocationId),
		SubnetId:     aws.String(subnetId),
	}

	if !dryRun {
		createNGResp, err := svc.CreateNatGateway(params)
		if err != nil {
			return "", err
		}

		gatewayId := *createNGResp.NatGateway.NatGatewayId
		terminal.Delta("Created VPC NAT Gateway [" + gatewayId + "] named [" + name + "] in [" + region + "]!")

		terminal.Notice("Waiting until the NAT Gateway is available...")

		err = svc.WaitUntilNatGatewayAvailable(&ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []*string{
				aws.String(gatewayId),
			},
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return gatewayId, errors.New(awsErr.Message())
			}
		}

		return gatewayId, err
	}

	return "", nil
}

// CreateRouteTable creates a new VPC Route Table
func CreateRouteTable(name, vpcSearch string, dryRun bool) (string, error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	vpcList, _ := GetVpcs(vpcSearch)

	vpcCount := len(*vpcList)
	if vpcCount == 0 {
		return "", errors.New("No VPCs found for your search terms.")
	}
	if vpcCount > 1 {
		vpcList.PrintTable()
		return "", errors.New("Please limit your search to return only one VPC.")
	}
	vpc := (*vpcList)[0]

	// Create it
	return createRouteTable(name, vpc.VpcID, vpc.Region, dryRun)
}

func createRouteTable(name, vpcId, region string, dryRun bool) (string, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	// Create the Route Table
	params := &ec2.CreateRouteTableInput{
		VpcId:  aws.String(vpcId),
		DryRun: aws.Bool(dryRun),
	}

	createRTResp, err := svc.CreateRouteTable(params)
	if err != nil {
		return "", err
	}

	rtId := *createRTResp.RouteTable.RouteTableId
	terminal.Delta("Created VPC Route Table [" + rtId + "] named [" + name + "] in [" + region + "]!")

	// Tag it
	err = SetEc2NameAndClassTags(&rtId, name, "", region)
	if err != nil {
		return rtId, err
	}

	return rtId, nil
}

func createRoute(routeTableId, region, destinationCidr, gatewayId, natGatewayId string, dryRun bool) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.CreateRouteInput{
		RouteTableId: aws.String(routeTableId),
		DryRun:       aws.Bool(dryRun),
		// NetworkInterfaceId:   			aws.String("String"),
		// DestinationIpv6CidrBlock:		aws.String("String"),
		// EgressOnlyInternetGatewayId:		aws.String("String"),
		// InstanceId:						aws.String("String"),
		// VpcPeeringConnectionId:			aws.String("String"),
	}

	if destinationCidr != "" {
		params.SetDestinationCidrBlock(destinationCidr)
	}

	if gatewayId != "" {
		params.SetGatewayId(gatewayId)
	}

	if natGatewayId != "" {
		params.SetNatGatewayId(natGatewayId)
	}

	_, err := svc.CreateRoute(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Delta("Added route to [" + routeTableId + "]!")

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
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(vpc.Region)}))
		svc := ec2.New(sess)

		params := &ec2.DeleteVpcInput{
			VpcId:  aws.String(vpc.VpcID),
			DryRun: aws.Bool(dryRun),
		}

		_, err := svc.DeleteVpc(params)
		if err != nil {
			return err
		}

		terminal.Delta("Deleted VPC [" + vpc.Name + "] in [" + vpc.Region + "]!")
	}

	return nil
}
