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
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type SimpleDBDomains []SimpleDBDomain

type SimpleDBDomain struct {
	Name   string
	Region string
}

func GetSimpleDBDomains(search string) (*SimpleDBDomains, []error) {
	var wg sync.WaitGroup
	var errs []error

	domainList := new(SimpleDBDomains)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSimpleDBDomains(region.RegionName, domainList, search)
			if err != nil {
				// TODO handle regions without service endpoints that work
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering simpledb domain list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}

	wg.Wait()

	return domainList, errs
}

func GetRegionSimpleDBDomains(region *string, domainList *SimpleDBDomains, search string) error {
	svc := simpledb.New(session.New(&aws.Config{Region: region}))
	result, err := svc.ListDomains(nil)
	if err != nil {
		//return err // TODO handle regions without services
	}

	domains := make(SimpleDBDomains, len(result.DomainNames))
	for i, domain := range result.DomainNames {

		domains[i] = SimpleDBDomain{
			Name:   aws.StringValue(domain),
			Region: aws.StringValue(region),
		}
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, dn := range domains {
			rDomain := reflect.ValueOf(dn)

			for k := 0; k < rDomain.NumField(); k++ {
				sVal := rDomain.Field(k).String()

				if term.MatchString(sVal) {
					*domainList = append(*domainList, domains[i])
					continue Loop
				}
			}
		}
	} else {
		*domainList = append(*domainList, domains[:]...)
	}

	return nil
}

func (i *SimpleDBDomains) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Domains Found!")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Region"})

	table.AppendBulk(rows)
	table.Render()
}

func CreateSimpleDBDomain(domain, region string) error {

	// Validate the region
	if !ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	svc := simpledb.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &simpledb.CreateDomainInput{
		DomainName: aws.String(domain),
	}

	terminal.Information("Creating SimpleDB Domain [" + domain + "] in [" + region + "]...")

	_, err := svc.CreateDomain(params)
	if err == nil {
		terminal.Information("Done!")
	}

	return err
}

func DeleteSimpleDBDomain(search, region string) (err error) {

	domainList := new(SimpleDBDomains)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionSimpleDBDomains(aws.String(region), domainList, search)
	} else {
		domainList, _ = GetSimpleDBDomains(search)
	}

	if err != nil {
		terminal.ErrorLine("Error gathering SimpleDB domains list")
		return
	}

	if len(*domainList) > 0 {
		// Print the table
		domainList.PrintTable()
	} else {
		terminal.ErrorLine("No SimpleDB Domains found, Aborting!")
		return
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these SimpleDB Domains?") {
		terminal.ErrorLine("Aborting!")
		return
	}

	// Delete 'Em
	for _, domain := range *domainList {
		svc := simpledb.New(session.New(&aws.Config{Region: aws.String(domain.Region)}))

		params := &simpledb.DeleteDomainInput{
			DomainName: aws.String(domain.Name),
		}
		_, err = svc.DeleteDomain(params)
		if err != nil {
			terminal.ErrorLine("Error while deleting SimpleDB Domain [" + domain.Name + "] in [" + domain.Region + "], Aborting!")
			return
		}
		terminal.Information("Deleted SimpleDB Domain [" + domain.Name + "] in [" + domain.Region + "]!")
	}

	terminal.Information("Done!")

	return
}
