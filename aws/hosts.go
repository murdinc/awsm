package aws

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// ResourceRecords represents a slice of AWS Route53 Host Records
type ResourceRecords []ResourceRecord

// ResourceRecord represents a single Route53 Host Record
type ResourceRecord models.ResourceRecord

// ResourceRecordChanges represents a slice of AWS Route53 Host Record Changes
type ResourceRecordChanges []ResourceRecordChange

// ResourceRecordChange represents a single Route53 Host Record Change
type ResourceRecordChange models.ResourceRecordChange

// HostedZones represents a slice of AWS Route53 Hosted Zones
type HostedZones []HostedZone

// HostedZone represents a single Route53 Hosted Zone
type HostedZone models.HostedZone

// DeleteResourceRecords deletes AWS Route53 Resource Records
func DeleteResourceRecords(search string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	resourceRecordList, err := GetResourceRecords(search)
	if err != nil {
		return err
	}

	if len(*resourceRecordList) > 0 {
		// Print the table
		resourceRecordList.PrintTable()
	} else {
		return errors.New("No Resource Records found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Resource Records?") {
		return errors.New("Aborting!")
	}

	changeSet := make(map[string][]ResourceRecordChange)

	for _, record := range *resourceRecordList {
		changeSet[record.HostedZoneId] = append(changeSet[record.HostedZoneId],
			ResourceRecordChange{
				Action: "DELETE",
				Name:   record.Name,
				Values: record.Values,
				Type:   record.Type,
				TTL:    record.TTL,
			},
		)
	}

	return changeResourceRecord(changeSet, dryRun)
}

func (h *HostedZone) GetResourceRecords(search string) (*ResourceRecords, error) {
	regions := regions.GetRegionList()
	rand.Seed(time.Now().UnixNano())
	region := regions[rand.Intn(len(regions))] // pick a random region

	resourceRecordList := new(ResourceRecords)

	err := GetRegionResourceRecords(h.Id, *region.RegionName, resourceRecordList, search)
	if err != nil {
		return resourceRecordList, err
	}

	if len(*resourceRecordList) == 0 {
		return resourceRecordList, errors.New("No Resource Records Found!")
	}

	return resourceRecordList, nil
}

// CreateResourceRecord creates an AWS Route53 Resource Record
func CreateResourceRecord(name, value string, ttl string, force, dryRun bool) error {

	// If we were not passed a value, try to get it from the ec2metadata instead
	if value == "" {
		terminal.Notice("No value given, attempting to get value from ec2 meta-data...")
		sess := session.Must(session.NewSession())
		svc := ec2metadata.New(sess)

		terminal.Information("Trying to fine a Public IP Address...")
		publicIp, err := svc.GetMetadata("public-ipv4")
		if publicIp == "" || err != nil {
			terminal.Notice("Unable to find a Public IPv4 Address, looking for a Local IPv4 Address...")
			localIp, err := svc.GetMetadata("local-ipv4")
			if localIp == "" || err != nil {
				terminal.Notice("Unable to find a Local IPv4 Address, please provide a value instead. Aborting!")
				return err
			}
			value = localIp
		} else {
			value = publicIp
		}
		terminal.Delta("Using IP Address: " + value)
	}

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	action := "CREATE"
	// --force flag
	if force {
		terminal.Information("--force flag is set, using UPSERT Action!")
		action = "UPSERT"
	}

	var ttlInt int64
	if ttl == "" {
		ttlInt = 300
		terminal.Information("Using default TTL of [300]")
	} else {
		var err error
		ttlInt, err = strconv.ParseInt(ttl, 10, 64)
		if err != nil {
			return errors.New("Unable to use TTL value [" + ttl + "]")
		}
	}

	if validRecord(name) {
		terminal.Information("Name [" + name + "] appears to be a valid DNS Record.")
	} else {
		return errors.New("Name [" + name + "] appears to be invalid!")
	}

	recordType := "CNAME"

	switch {
	default:
		return errors.New("Value [" + value + "] is of an unknown type!")

	case govalidator.IsIPv4(value):
		terminal.Information("Value [" + value + "] appears to be a valid IPv4 Address.")
		recordType = "A"

	case govalidator.IsIPv6(value):
		terminal.Information("Value [" + value + "] appears to be a valid IPv6 Address.")
		recordType = "AAAA"

	}

	hostedZone, err := findHostedZone(name)
	if err != nil {
		return err
	}

	terminal.Information(fmt.Sprintf("Found Hosted Zone [%s - %s] with [%d] existing records.", hostedZone.Id, hostedZone.Name, hostedZone.ResourceRecordSetCount))

	changeSet := make(map[string][]ResourceRecordChange)

	changeSet[hostedZone.Id] = append(changeSet[hostedZone.Id],
		ResourceRecordChange{
			Action: action,
			Name:   name,
			Values: []string{value},
			Type:   recordType,
			TTL:    int(ttlInt),
		},
	)

	err = changeResourceRecord(changeSet, dryRun)
	if !force && err != nil && strings.Contains(err.Error(), "already exists") {
		terminal.Information(err.Error())
		update := terminal.PromptBool("Do you want to update (UPSERT) it instead?")
		if update {
			changeSet[hostedZone.Id][0].Action = "UPSERT"
			return changeResourceRecord(changeSet, dryRun)
		}

		terminal.ErrorLine("Aborting!")
		return nil
	}

	return err
}

func findHostedZone(name string) (HostedZone, error) {

	// bad assumption stripping the first chunk?
	hostedZoneSearch := strings.Split(name, ".")[1:]
	hostedZoneSearchLen := len(hostedZoneSearch)

	hostedZones, err := GetHostedZones("")
	if err != nil {
		return HostedZone{}, err
	}
	// Chunk the hosted zones
	for _, hostedZone := range *hostedZones {
		// strip trailing dot and split at the other dots
		//hostedZoneChunked := strings.Split(hostedZone.Name, ".")
		hostedZoneChunked := strings.Split(hostedZone.Name[:len(hostedZone.Name)-1], ".")

		for s := 0; s < hostedZoneSearchLen; s++ {
			hostedZoneSearchChunk := hostedZoneSearch[s:]

			if reflect.DeepEqual(hostedZoneChunked, hostedZoneSearchChunk) {
				return hostedZone, nil
			}
		}
	}

	return HostedZone{}, errors.New("Unable to find Hosted Zone for [" + name + "]")
}

// private function without terminal prompts
func changeResourceRecord(changeSet map[string][]ResourceRecordChange, dryRun bool) error {

	for id, changes := range changeSet {

		params := &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(id),
		}

		recordChanges := make([]*route53.Change, len(changes))

		for i, change := range changes {
			recordChanges[i] = &route53.Change{
				Action: aws.String(change.Action),
				ResourceRecordSet: &route53.ResourceRecordSet{
					Name: aws.String(change.Name),
					Type: aws.String(change.Type),
					TTL:  aws.Int64(int64(change.TTL)),
					/*
						AliasTarget: &route53.AliasTarget{
							DNSName:              aws.String("DNSName"),    // Required
							EvaluateTargetHealth: aws.Bool(true),           // Required
							HostedZoneId:         aws.String("ResourceId"), // Required
						},
						SetIdentifier: aws.String("ResourceRecordSetIdentifier"),
						TTL:           aws.Int64(1),
						TrafficPolicyInstanceId: aws.String("TrafficPolicyInstanceId"),
						Failover: aws.String("ResourceRecordSetFailover"),
						GeoLocation: &route53.GeoLocation{
							ContinentCode:   aws.String("GeoLocationContinentCode"),
							CountryCode:     aws.String("GeoLocationCountryCode"),
							SubdivisionCode: aws.String("GeoLocationSubdivisionCode"),
						},
						HealthCheckId: aws.String("HealthCheckId"),
						Region:        aws.String("ResourceRecordSetRegion"),
						Weight:                  aws.Int64(1),
					*/
				},
			}

			resourceRecords := make([]*route53.ResourceRecord, len(change.Values))
			for j, value := range change.Values {
				resourceRecords[j] = new(route53.ResourceRecord)
				resourceRecords[j].SetValue(value)
			}

			terminal.Delta("[" + change.Action + "] - Resource Record [" + change.Name + "] : [" + strings.Join(change.Values, ", ") + "]")

			recordChanges[i].ResourceRecordSet.SetResourceRecords(resourceRecords)

		}

		changeBatch := &route53.ChangeBatch{}
		changeBatch.SetChanges(recordChanges)
		params.SetChangeBatch(changeBatch)

		if !dryRun {
			regions := regions.GetRegionList()

			rand.Seed(time.Now().UnixNano())
			region := regions[rand.Intn(len(regions))] // pick a random region

			sess := session.Must(session.NewSession(&aws.Config{Region: region.RegionName}))
			svc := route53.New(sess)

			_, err := svc.ChangeResourceRecordSets(params)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					return errors.New(awsErr.Message())
				}
				return err
			}
		}
	}

	terminal.Information("Done!")

	return nil
}

// GetResourceRecords returns a list of Route53 Resource Records that match the provided search term
func GetResourceRecords(search string) (*ResourceRecords, error) {

	resourceRecordList := new(ResourceRecords)

	hostZones, err := GetHostedZones("")
	if err != nil {
		return resourceRecordList, err
	}

	regions := regions.GetRegionList()

	rand.Seed(time.Now().UnixNano())
	region := regions[rand.Intn(len(regions))] // pick a random region

	for _, host := range *hostZones {
		err = GetRegionResourceRecords(host.Id, *region.RegionName, resourceRecordList, search)
		if err != nil {
			return resourceRecordList, err
		}
	}

	return resourceRecordList, err

}

// GetRegionResourceRecords returns a list of ResourceRecords for a given region into the provided ResourceRecords slice
func GetRegionResourceRecords(hostedZoneId, region string, resourceRecordList *ResourceRecords, search string) error {

	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneId),
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := route53.New(sess)

	var resourceRecordSetsResult []*route53.ResourceRecordSet

ListLoop:
	for {
		result, err := svc.ListResourceRecordSets(params)
		if err != nil {
			return err
		}

		resourceRecordSetsResult = append(resourceRecordSetsResult, result.ResourceRecordSets...)

		// Break when done
		if !aws.BoolValue(result.IsTruncated) {
			break ListLoop
		}

		params.StartRecordName = result.NextRecordName
	}

	resourceRecords := make(ResourceRecords, len(resourceRecordSetsResult))
	for i, b := range resourceRecordSetsResult {
		resourceRecords[i].Marshal(b, hostedZoneId)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, kp := range resourceRecords {
			rhostedZones := reflect.ValueOf(kp)

			for k := 0; k < rhostedZones.NumField(); k++ {
				sVal := rhostedZones.Field(k).String()

				if term.MatchString(sVal) {
					*resourceRecordList = append(*resourceRecordList, resourceRecords[i])
					continue Loop
				}
			}
		}
	} else {
		*resourceRecordList = append(*resourceRecordList, resourceRecords[:]...)
	}

	return nil
}

// GetHostedZones returns a list of Route53 Hosted Zones that match the provided search term
func GetHostedZones(search string) (*HostedZones, error) {

	hostedZoneList := new(HostedZones)
	regions := regions.GetRegionList()

	rand.Seed(time.Now().UnixNano())
	region := regions[rand.Intn(len(regions))] // pick a random region

	err := GetRegionHostedZones(*region.RegionName, hostedZoneList, search)

	return hostedZoneList, err
}

// GetRegionHostedZones returns a list of HostedZones for a given region into the provided HostedZones slice
func GetRegionHostedZones(region string, hostedZoneList *HostedZones, search string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := route53.New(sess)

	params := &route53.ListHostedZonesInput{}

	var hostedZonesResult []*route53.HostedZone

HostedZonesLoop:
	for {
		result, err := svc.ListHostedZones(params)
		if err != nil {
			return err
		}

		hostedZonesResult = append(hostedZonesResult, result.HostedZones...)

		// Break when done
		if !aws.BoolValue(result.IsTruncated) {
			break HostedZonesLoop
		}

		params.Marker = result.NextMarker
	}

	hostedZones := make(HostedZones, len(hostedZonesResult))

	for i, b := range hostedZonesResult {
		hostedZones[i].Marshal(b)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, kp := range hostedZones {
			rhostedZones := reflect.ValueOf(kp)

			for k := 0; k < rhostedZones.NumField(); k++ {
				sVal := rhostedZones.Field(k).String()

				if term.MatchString(sVal) {
					*hostedZoneList = append(*hostedZoneList, hostedZones[i])
					continue Loop
				}
			}
		}
	} else {
		*hostedZoneList = append(*hostedZoneList, hostedZones[:]...)
	}

	return nil
}

// Marshal parses the response from the aws sdk into an awsm ResourceRecord
func (h *ResourceRecord) Marshal(resourceRecordSet *route53.ResourceRecordSet, hostedZoneId string) {

	h.Name, _ = strconv.Unquote("\"" + aws.StringValue(resourceRecordSet.Name) + "\"")
	h.Type = aws.StringValue(resourceRecordSet.Type)
	h.TTL = int(aws.Int64Value(resourceRecordSet.TTL))
	h.Region = aws.StringValue(resourceRecordSet.Region)
	h.Failover = aws.StringValue(resourceRecordSet.Failover)
	h.HostedZoneId = hostedZoneId

	h.HealthCheckId = aws.StringValue(resourceRecordSet.HealthCheckId)

	if resourceRecordSet.AliasTarget != nil {
		aliasTarget := models.AliasTarget{
			DNSName:              aws.StringValue(resourceRecordSet.AliasTarget.DNSName),
			EvaluateTargetHealth: aws.BoolValue(resourceRecordSet.AliasTarget.EvaluateTargetHealth),
			HostedZoneId:         aws.StringValue(resourceRecordSet.AliasTarget.HostedZoneId),
		}

		h.AliasTarget = aliasTarget
	}

	for _, record := range resourceRecordSet.ResourceRecords {
		value := aws.StringValue(record.Value)

		if len(value) > 50 {
			h.TableValues = append(h.TableValues, value[:50]+"...")
		} else {
			h.TableValues = append(h.TableValues, value)
		}

		h.Values = append(h.Values, value)
	}

	if h.AliasTarget.DNSName != "" {
		h.TableValues = append(h.TableValues, h.AliasTarget.DNSName)
	}

}

// Marshal parses the response from the aws sdk into an awsm HostedZone
func (h *HostedZone) Marshal(hostedZone *route53.HostedZone) {
	h.Name = aws.StringValue(hostedZone.Name)
	h.Id = aws.StringValue(hostedZone.Id)
	h.Comment = aws.StringValue(hostedZone.Config.Comment)
	h.PrivateZone = aws.BoolValue(hostedZone.Config.PrivateZone)
	h.ResourceRecordSetCount = int(aws.Int64Value(hostedZone.ResourceRecordSetCount))
}

// PrintTable Prints an ascii table of the list of KeyPairs
func (h *HostedZones) PrintTable() {
	if len(*h) == 0 {
		terminal.ShowErrorMessage("Warning", "No Route53 Hosted Zones Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*h))

	for index, key := range *h {
		models.ExtractAwsmTable(index, key, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

// PrintTable Prints an ascii table of the list of ResourceRecords
func (h *ResourceRecords) PrintTable() {
	if len(*h) == 0 {
		terminal.ShowErrorMessage("Warning", "No Route53 Resource Records Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*h))

	for index, key := range *h {
		models.ExtractAwsmTable(index, key, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

// mostly yanked from pkg/net isDomainName()
func validRecord(s string) bool {
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}

	last := byte('.')
	ok := false // Ok once we've seen a letter.
	partlen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		default:
			return false
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			ok = true
			partlen++
		case '0' <= c && c <= '9':
			// fine
			partlen++
		case c == '*':
			if partlen > 0 {
				return false
			}
			partlen++

		case c == '-':
			// Byte before dash cannot be dot.
			if last == '.' {
				return false
			}
			partlen++
		case c == '.':
			// Byte before dot cannot be dot, dash.
			if last == '.' || last == '-' {
				return false
			}
			if partlen > 63 || partlen == 0 {
				return false
			}
			partlen = 0
		}
		last = c
	}
	if last == '-' || partlen > 63 {
		return false
	}

	return ok
}
