package aws

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/asaskevich/govalidator"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/mitchellh/hashstructure"
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// LoadBalancers represents a slice of AWS Load Balancers
type LoadBalancers []LoadBalancer

// LoadBalancer represents a single AWS Load Balancer
type LoadBalancer models.LoadBalancer

// GetLoadBalancers returns a slice of AWS Load Balancers
func GetLoadBalancers(search string) (*LoadBalancers, []error) {
	var wg sync.WaitGroup
	var errs []error

	lbList := new(LoadBalancers)
	regions := regions.GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionLoadBalancers(*region.RegionName, lbList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering loadbalancer list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return lbList, errs
}

// GetRegionLoadBalancers returns a list of Load Balancers in a region into the provided LoadBalancers slice
func GetRegionLoadBalancers(region string, lbList *LoadBalancers, search string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := elb.New(sess)

	result, err := svc.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})

	if err != nil {
		return err
	}

	secGrpList := new(SecurityGroups)
	vpcList := new(Vpcs)
	subList := new(Subnets)
	GetRegionSecurityGroups(region, secGrpList, "")
	GetRegionVpcs(region, vpcList, "")
	GetRegionSubnets(region, subList, "")

	// Get the tags all at once, to save time
	elbNames := []string{}
	for _, lb := range result.LoadBalancerDescriptions {
		elbNames = append(elbNames, aws.StringValue(lb.LoadBalancerName))
	}

	elbTags, err := GetLoadBalancerTags(elbNames, region)
	if err != nil {
		return err
	}

	lb := make(LoadBalancers, len(result.LoadBalancerDescriptions))
	for i, balancer := range result.LoadBalancerDescriptions {
		lb[i].Marshal(balancer, region, secGrpList, vpcList, subList, elbTags)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, g := range lb {
			rAsg := reflect.ValueOf(g)

			for k := 0; k < rAsg.NumField(); k++ {
				sVal := rAsg.Field(k).String()

				if term.MatchString(sVal) {
					*lbList = append(*lbList, lb[i])
					continue Loop
				}
			}
		}
	} else {
		*lbList = append(*lbList, lb[:]...)
	}

	return nil
}

// GetLoadBalancerByName returns a single Load Balancer given the provided region and name
func GetLoadBalancerByName(region, name string) (LoadBalancer, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := elb.New(sess)

	params := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{
			aws.String(name),
		},
	}
	result, err := svc.DescribeLoadBalancers(params)
	if err != nil || len(result.LoadBalancerDescriptions) == 0 {
		return LoadBalancer{}, err
	}

	count := len(result.LoadBalancerDescriptions)

	switch count {
	case 0:
		return LoadBalancer{}, errors.New("No Load Balancers found with name of [" + name + "] in [" + region + "].")
	case 1:
		secGrpList := new(SecurityGroups)
		vpcList := new(Vpcs)
		subList := new(Subnets)
		elbTags, _ := GetLoadBalancerTags([]string{name}, region)
		GetRegionSecurityGroups(region, secGrpList, "")
		GetRegionVpcs(region, vpcList, "")
		GetRegionSubnets(region, subList, "")

		lb := new(LoadBalancer)
		lb.Marshal(result.LoadBalancerDescriptions[0], region, secGrpList, vpcList, subList, elbTags)
		return *lb, nil
	}

	return LoadBalancer{}, errors.New("Found more than one Load Balancer named [" + name + "] in [" + region + "]!")
}

func GetLoadBalancerTags(names []string, region string) (map[string][]*elb.Tag, error) {

	elbTags := make(map[string][]*elb.Tag)

	if len(names) == 0 {
		return elbTags, nil
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := elb.New(sess)

	params := &elb.DescribeTagsInput{
		LoadBalancerNames: aws.StringSlice(names),
	}

	resp, err := svc.DescribeTags(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr)
			return elbTags, errors.New(awsErr.Message())
		}
		return elbTags, err
	}

	for _, tags := range resp.TagDescriptions {
		name := aws.StringValue(tags.LoadBalancerName)
		elbTags[name] = append(elbTags[name], tags.Tags...)
	}

	return elbTags, nil
}

// Marshal parses the response from the aws sdk into an awsm LoadBalancer
func (l *LoadBalancer) Marshal(balancer *elb.LoadBalancerDescription, region string, secGrpList *SecurityGroups, vpcList *Vpcs, subList *Subnets, tags map[string][]*elb.Tag) {

	// security groups
	secGroupNames := secGrpList.GetSecurityGroupNames(aws.StringValueSlice(balancer.SecurityGroups))
	secGroupNamesSorted := sort.StringSlice(secGroupNames[0:])
	secGroupNamesSorted.Sort()

	// subnets
	subnetNames := subList.GetSubnetNames(aws.StringValueSlice(balancer.Subnets))
	subnetNamesSorted := sort.StringSlice(subnetNames[0:])
	subnetNamesSorted.Sort()

	l.Name = aws.StringValue(balancer.LoadBalancerName)
	l.DNSName = aws.StringValue(balancer.DNSName)
	l.CreatedTime = aws.TimeValue(balancer.CreatedTime)
	l.VpcID = aws.StringValue(balancer.VPCId)
	l.Vpc = vpcList.GetVpcName(l.VpcID)
	l.SubnetIDs = aws.StringValueSlice(balancer.Subnets)
	l.Subnets = strings.Join(subnetNamesSorted, ", ")
	l.HealthCheckTarget = aws.StringValue(balancer.HealthCheck.Target)
	l.HealthCheckInterval = fmt.Sprintf("%d seconds", *balancer.HealthCheck.Interval)
	l.Scheme = aws.StringValue(balancer.Scheme)
	l.SecurityGroups = strings.Join(secGroupNamesSorted, ", ")
	l.AvailabilityZones = strings.Join(aws.StringValueSlice(balancer.AvailabilityZones), ", ")
	l.Region = region
	l.Class = GetTagValue("Class", tags[l.Name])

	// Get the listeners
	for _, listenerDesc := range balancer.ListenerDescriptions {
		listener := listenerDesc.Listener
		l.LoadBalancerListeners = append(l.LoadBalancerListeners, config.LoadBalancerListener{
			InstancePort:     int(aws.Int64Value(listener.InstancePort)),
			LoadBalancerPort: int(aws.Int64Value(listener.LoadBalancerPort)),
			Protocol:         aws.StringValue(listener.Protocol),
			InstanceProtocol: aws.StringValue(listener.InstanceProtocol),
			SSLCertificateID: aws.StringValue(listener.SSLCertificateId),
		})
	}

}

/*

// LoadBalancerListener is a Load Balancer Listener
type LoadBalancerListener struct {
	ID               string `json:"id" hash:"ignore" awsm:"ignore"` // Needed?
	InstancePort     int    `json:"instancePort"`
	LoadBalancerPort int    `json:"loadBalancerPort"`
	Protocol         string `json:"protocol"`
	InstanceProtocol string `json:"instanceProtocol"`
	SSLCertificateID string `json:"sslCertificateID"`
}

*/

// PrintTable Prints an ascii table of the list of Load Balancers
func (i *LoadBalancers) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Load Balancers Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, lb := range *i {
		models.ExtractAwsmTable(index, lb, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

func CreateLoadBalancer(class, region, vpcName string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Bail if it already exists. For some reason there is no error when creating an ELB that already exists?
	lb, _ := GetLoadBalancerByName(region, class)
	if lb.Name == class {
		return errors.New("Load Balancer [" + class + "] already exists in [" + region + "]")
	}

	// Class Config
	elbCfg, err := config.LoadLoadBalancerClass(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Load Balancer Class Configuration for [" + class + "]!")
	}

	// Validate the region
	if !regions.ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	// placeholders
	secGrpIds := []*string{}
	subnetIds := []*string{}

	// Validate the vpc if passed one - with security groups, and get the matching security groups
	if vpcName != "" {
		vpc, err := GetVpcByTag(region, "Name", vpcName)
		if err != nil {
			return err
		}

		// Add Subnets
		for _, sn := range elbCfg.Subnets {
			subnet, err := vpc.GetVpcSubnetByTag("Name", sn)
			if err != nil {
				return err
			}

			subnetIds = append(subnetIds, aws.String(subnet.SubnetID))
		}

		// Get the vpc security groups while we are at it.
		secGroups, err := vpc.GetVpcSecurityGroupByTagMulti("Class", elbCfg.Subnets)
		if err != nil {
			return err
		}
		for _, secGroup := range secGroups {
			secGrpIds = append(secGrpIds, aws.String(secGroup.GroupID))
		}
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := elb.New(sess)

	params := &elb.CreateLoadBalancerInput{
		LoadBalancerName: aws.String(class),
		Scheme:           aws.String(elbCfg.Scheme),

		Tags: []*elb.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(class),
			},
			{
				Key:   aws.String("Class"),
				Value: aws.String(class),
			},
		},
	}

	// Add Subnets
	if len(subnetIds) > 0 {
		params.SetSubnets(subnetIds)
	}

	// Add Security Groups
	if len(secGrpIds) > 0 {
		params.SetSecurityGroups(secGrpIds)
	}

	// Add Availability Zones that are in this region
	azs := []*string{}
	regionAzs := new(regions.AZs)
	regions.GetRegionAZs(region, regionAzs)
	for _, az := range elbCfg.AvailabilityZones {
		if regionAzs.ValidAZ(az) {
			terminal.Information("Found Availability Zone [" + az + "]!")
			azs = append(azs, aws.String(az))
		}
	}
	if len(azs) > 0 {
		params.SetAvailabilityZones(azs)
	}

	// Add Listeners
	if len(elbCfg.LoadBalancerListeners) > 0 {
		listeners := []*elb.Listener{}
		for _, l := range elbCfg.LoadBalancerListeners {

			if !govalidator.IsPort(fmt.Sprint(l.InstancePort)) {
				return errors.New("Instance Port [" + fmt.Sprint(l.InstancePort) + "] is invalid!")
			}

			if !govalidator.IsPort(fmt.Sprint(l.LoadBalancerPort)) {
				return errors.New("Load Balancer Port [" + fmt.Sprint(l.LoadBalancerPort) + "] is invalid!")
			}

			listeners = append(listeners, &elb.Listener{
				InstancePort:     aws.Int64(int64(l.InstancePort)),
				LoadBalancerPort: aws.Int64(int64(l.LoadBalancerPort)),
				Protocol:         aws.String(l.Protocol),
				InstanceProtocol: aws.String(l.InstanceProtocol),
				//SSLCertificateId: aws.String("SSLCertificateId"),
			},
			)
		}
		params.SetListeners(listeners)
	}

	createLoadBalancerResp, err := svc.CreateLoadBalancer(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Created Load Balancer [" + *createLoadBalancerResp.DNSName + "] named [" + class + "] in [" + region + "]!")

	return nil

}

// UpdateLoadBalancers updates one or more Load Balancers that match the provided search term and optional region
func UpdateLoadBalancers(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	lbList := new(LoadBalancers)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionLoadBalancers(region, lbList, search)
	} else {
		lbList, _ = GetLoadBalancers(search)
	}

	if err != nil {
		return errors.New("Error gathering Security Groups list")
	}

	if len(*lbList) > 0 {
		// Print the table
		lbList.PrintTable()
	} else {
		return errors.New("No Load Balancers found, Aborting!")
	}

	changes, err := lbList.Diff()
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		terminal.Information("There are no changes needed on these Load Balancers!")
		return nil
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to update these Load Balancers?") {
		return errors.New("Aborting!")
	}

	// Update 'Em
	err = updateLoadBalancers(changes, dryRun)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Done!")

	return nil
}

func updateLoadBalancers(changes []LoadBalancerChange, dryRun bool) error {

	if !dryRun {
		for _, change := range changes {
			if change.Remove {
				// remove
				err := removeListener(change.LoadBalancer, change.Listeners)
				if err != nil {
					return err
				}
			} else {
				// add
				err := addListener(change.LoadBalancer, change.Listeners)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func addListener(lb LoadBalancer, listeners []config.LoadBalancerListener) error {

	params := &elb.CreateLoadBalancerListenersInput{
		LoadBalancerName: aws.String(lb.Name),
	}

	elbListeners := []*elb.Listener{}

	for _, list := range listeners {

		elbListener := &elb.Listener{}

		elbListener.SetInstancePort(int64(list.InstancePort)).
			SetLoadBalancerPort(int64(list.LoadBalancerPort)).
			SetProtocol(list.Protocol).
			SetInstanceProtocol(list.InstanceProtocol).
			SetSSLCertificateId(list.SSLCertificateID)

		elbListeners = append(elbListeners, elbListener)
	}

	params.SetListeners(elbListeners)

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(lb.Region)}))
	svc := elb.New(sess)

	_, err := svc.CreateLoadBalancerListeners(params)
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

func removeListener(lb LoadBalancer, listeners []config.LoadBalancerListener) error {

	params := &elb.DeleteLoadBalancerListenersInput{
		LoadBalancerName: aws.String(lb.Name),
	}

	elbPorts := []*int64{}

	for _, list := range listeners {
		elbPorts = append(elbPorts, aws.Int64(int64(list.LoadBalancerPort)))
	}

	params.SetLoadBalancerPorts(elbPorts)

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(lb.Region)}))
	svc := elb.New(sess)

	_, err := svc.DeleteLoadBalancerListeners(params)
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

type LoadBalancerChange struct {
	LoadBalancer LoadBalancer
	Remove       bool
	Listeners    []config.LoadBalancerListener
}

func (s LoadBalancers) Diff() ([]LoadBalancerChange, error) {

	terminal.Delta("Comparing awsm load balancer listeners...")

	changes := []LoadBalancerChange{}
	cfgHashes := make([]map[uint64]config.LoadBalancerListener, len(s))

	for i, lb := range s {
		cfgHashes[i] = make(map[uint64]config.LoadBalancerListener)
		// Verify the security group class input
		cfg, err := config.LoadLoadBalancerClass(lb.Class)
		if err != nil {
			return changes, err
		}

		// cycle through the config listeners and generate hashess
		for _, cListener := range cfg.LoadBalancerListeners {

			configListenerHash, err := hashstructure.Hash(cListener, nil)
			if err != nil {
				return changes, err
			}
			cfgHashes[i][configListenerHash] = cListener

		}

	}

	for i, lb := range s {

		var remove, add []config.LoadBalancerListener

		// cycle through existing grants and find ones to remove
		for _, listener := range lb.LoadBalancerListeners {

			existingListenerHash, err := hashstructure.Hash(listener, nil)
			if err != nil {
				return changes, err
			}

			if _, ok := cfgHashes[i][existingListenerHash]; !ok {
				terminal.Delta(fmt.Sprintf("[%s %s] - Remove -	[%s:%d	-	%s:%d]", lb.Name, lb.Region, listener.Protocol, listener.LoadBalancerPort, listener.InstanceProtocol, listener.InstancePort))
				remove = append(remove, listener)

			} else {
				//terminal.Notice(fmt.Sprintf("[%s %s] - Keeping -	[%s:%d	-	%s:%d]", lb.Name, lb.Region, listener.Protocol, listener.LoadBalancerPort, listener.InstanceProtocol, listener.InstancePort))
				delete(cfgHashes[i], existingListenerHash)
			}
		}

		// cycle through hashes and find ones to add
		for _, listener := range cfgHashes[i] {
			terminal.Delta(fmt.Sprintf("[%s %s] - Add -	[%s:%d	-	%s:%d]", lb.Name, lb.Region, listener.Protocol, listener.LoadBalancerPort, listener.InstanceProtocol, listener.InstancePort))
			add = append(add, listener)

		}

		// remove
		if len(remove) > 0 {
			changes = append(changes, LoadBalancerChange{
				LoadBalancer: lb,
				Listeners:    remove,
				Remove:       true,
			})
		}

		// add
		if len(add) > 0 {
			changes = append(changes, LoadBalancerChange{
				LoadBalancer: lb,
				Listeners:    add,
			})
		}
	}

	terminal.Information("Comparison complete!")
	return changes, nil
}

// Public function with confirmation terminal prompt
func DeleteLoadBalancers(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	elbList := new(LoadBalancers)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionLoadBalancers(region, elbList, search)
	} else {
		elbList, _ = GetLoadBalancers(search)
	}

	if err != nil {
		return errors.New("Error gathering Load Balancer list")
	}

	if len(*elbList) > 0 {
		// Print the table
		elbList.PrintTable()
	} else {
		return errors.New("No Load Balancers found matching your search term, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Load Balancers?") {
		return errors.New("Aborting!")
	}

	if !dryRun { // no dryRun param on this aws operation
		// Delete 'Em
		err = deleteLoadBalancers(elbList)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func deleteLoadBalancers(elbList *LoadBalancers) (err error) {
	for _, lb := range *elbList {
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(lb.Region)}))
		svc := elb.New(sess)

		params := &elb.DeleteLoadBalancerInput{
			LoadBalancerName: aws.String(lb.Name),
		}

		_, err := svc.DeleteLoadBalancer(params)
		if err != nil {
			return err
		}

		terminal.Delta("Deleted Load Balancer [" + lb.Name + "] in [" + lb.Region + "]!")
	}

	return nil
}
