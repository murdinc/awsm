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
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// SimpleDBDomains represents a slice of SimpleDB Domains
type SimpleDBDomains []SimpleDBDomain

// SimpleDBDomain represents a single SimpleDB Domain
type SimpleDBDomain models.SimpleDBDomain

// GetSimpleDBDomains returns a slice of SimpleDB Domains that match the provided search term
func GetSimpleDBDomains(search string) (*SimpleDBDomains, []error) {
	var wg sync.WaitGroup
	var errs []error

	domainList := new(SimpleDBDomains)
	regions := GetRegionListWithoutIgnored()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSimpleDBDomains(*region.RegionName, domainList, search)
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

// GetRegionSimpleDBDomains returns a slice of a regions SimpleDB Domains into the provided SimpleDBDomains slice
func GetRegionSimpleDBDomains(region string, domainList *SimpleDBDomains, search string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := simpledb.New(sess)

	result, err := svc.ListDomains(nil)
	if err != nil {
		//return err // TODO handle regions without services
	}

	domains := make(SimpleDBDomains, len(result.DomainNames))
	for i, domain := range result.DomainNames {

		domains[i] = SimpleDBDomain{
			Name:   aws.StringValue(domain),
			Region: region,
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

// PrintTable Prints an ascii table of the list of SimpleDB Domains
func (i *SimpleDBDomains) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No SimpleDB Domains Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, domain := range *i {
		models.ExtractAwsmTable(index, domain, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

// CreateSimpleDBDomain creates a new SimpleDB Domain
func CreateSimpleDBDomain(domain, region string) error {

	// Validate the region
	if !regions.ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := simpledb.New(sess)

	params := &simpledb.CreateDomainInput{
		DomainName: aws.String(domain),
	}

	terminal.Delta("Creating SimpleDB Domain [" + domain + "] in [" + region + "]...")

	_, err := svc.CreateDomain(params)
	if err == nil {
		terminal.Information("Done!")
	}

	return err
}

// DeleteSimpleDBDomains deletes one or more SimpleDB Domains
func DeleteSimpleDBDomains(search, region string) (err error) {

	domainList := new(SimpleDBDomains)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionSimpleDBDomains(region, domainList, search)
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
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(domain.Region)}))
		svc := simpledb.New(sess)

		params := &simpledb.DeleteDomainInput{
			DomainName: aws.String(domain.Name),
		}
		_, err = svc.DeleteDomain(params)
		if err != nil {
			terminal.ErrorLine("Error while deleting SimpleDB Domain [" + domain.Name + "] in [" + domain.Region + "], Aborting!")
			return
		}
		terminal.Delta("Deleted SimpleDB Domain [" + domain.Name + "] in [" + domain.Region + "]!")
	}

	terminal.Information("Done!")

	return
}
