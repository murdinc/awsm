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
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"

	"github.com/olekukonko/tablewriter"
)

// Addresses represents a slice of Elastic IP Addresses
type Addresses []Address

// Address represents a single Elastic IP Address
type Address models.Address

// GetAddresses returns a slice of Elastic IP Addresses based on the given search term and optional available flag
func GetAddresses(search string, available bool) (*Addresses, []error) {
	var wg sync.WaitGroup
	var errs []error

	ipList := new(Addresses)
	regions := regions.GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionAddresses(*region.RegionName, ipList, search, available)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering address list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return ipList, errs
}

// Marshal parses the response from the aws sdk into an awsm Address
func (a *Address) Marshal(address *ec2.Address, region string, instList *Instances) {

	a.AllocationID = aws.StringValue(address.AllocationId)
	a.PublicIP = aws.StringValue(address.PublicIp)
	a.PrivateIP = aws.StringValue(address.PrivateIpAddress)
	a.InstanceID = aws.StringValue(address.InstanceId)
	a.Domain = aws.StringValue(address.Domain)
	a.NetworkInterfaceID = aws.StringValue(address.NetworkInterfaceId)
	a.NetworkInterfaceOwnerID = aws.StringValue(address.NetworkInterfaceOwnerId)
	a.Region = region

	switch a.InstanceID {
	case "":
		a.Status = "available"

	default:
		instance := instList.GetInstanceName(a.InstanceID)
		a.Status = "in-use"
		a.Attachment = instance
	}
}

// GetRegionAddresses returns a list of Elastic IP Addresses for a given region into the provided Addresses slice
func GetRegionAddresses(region string, adrList *Addresses, search string, available bool) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	result, err := svc.DescribeAddresses(&ec2.DescribeAddressesInput{})

	if err != nil {
		return err
	}

	instList := new(Instances)
	GetRegionInstances(region, instList, "", false)

	adr := make(Addresses, len(result.Addresses))
	for i, address := range result.Addresses {
		adr[i].Marshal(address, region, instList)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, ad := range adr {
			rAddr := reflect.ValueOf(ad)

			for k := 0; k < rAddr.NumField(); k++ {
				sVal := rAddr.Field(k).String()

				if term.MatchString(sVal) && ((available && adr[i].Status == "available") || !available) {
					*adrList = append(*adrList, adr[i])
					continue Loop
				}
			}
		}
	} else {
		if available {
			for i, _ := range adr {
				if adr[i].Status == "available" {
					*adrList = append(*adrList, adr[i])
				}
			}
		} else {
			*adrList = append(*adrList, adr[:]...)
		}
	}

	return nil
}

// CreateAddress creates a new Elastic IP Address in the given region and domain
func CreateAddress(region, domain string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Validate the region
	if !regions.ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	// Validate the domain
	if !(domain == "vpc" || domain != "classic") {
		return errors.New("Domain should be either [vpc] or [classic].")
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	// Create the address
	params := &ec2.AllocateAddressInput{
		Domain: aws.String(domain),
		DryRun: aws.Bool(dryRun),
	}
	_, err := svc.AllocateAddress(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil
}

// DeleteAddresses Deletes one or more Elastic IP Addresses based on the given search term and optional region
func DeleteAddresses(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	addrList := new(Addresses)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionAddresses(region, addrList, search, false)
	} else {
		addrList, _ = GetAddresses(search, false)
	}

	if err != nil {
		return errors.New("Error gathering Image list")
	}

	if len(*addrList) > 0 {
		// Print the table
		addrList.PrintTable()
	} else {
		return errors.New("No available Elastic IP Addresses found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Addresses?") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = deleteAddresses(addrList, dryRun)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func deleteAddresses(addrList *Addresses, dryRun bool) (err error) {
	for _, addr := range *addrList {

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(addr.Region)}))
		svc := ec2.New(sess)

		params := &ec2.ReleaseAddressInput{
			AllocationId: aws.String(addr.AllocationID),
			DryRun:       aws.Bool(dryRun),
			//PublicIp: aws.String(addr.PublicIp), // TODO required for ec2 classic
		}

		_, err := svc.ReleaseAddress(params)
		if err != nil {
			return err
		}

		terminal.Delta("Deleted Address [" + addr.AllocationID + "] in [" + addr.Region + "]!")
	}

	return nil
}

// PrintTable Prints an ascii table of the list of Elastic IP Addresses
func (i *Addresses) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Addresses Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, address := range *i {
		models.ExtractAwsmTable(index, address, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
