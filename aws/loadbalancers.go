package aws

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// LoadBalancers represents a slice of AWS Load Balancers
type LoadBalancers []LoadBalancer

// LoadBalancer represents a single AWS Load Balancer
type LoadBalancer models.LoadBalancer

// GetLoadBalancers returns a slice of AWS Load Balancers
func GetLoadBalancers() (*LoadBalancers, []error) {
	var wg sync.WaitGroup
	var errs []error

	lbList := new(LoadBalancers)
	regions := regions.GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionLoadBalancers(*region.RegionName, lbList)
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
func GetRegionLoadBalancers(region string, lbList *LoadBalancers) error {
	svc := elb.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})

	if err != nil {
		return err
	}

	secGrpList := new(SecurityGroups)
	err = GetRegionSecurityGroups(region, secGrpList, "")

	vpcList := new(Vpcs)
	subList := new(Subnets)
	GetRegionVpcs(region, vpcList, "")
	GetRegionSubnets(region, subList, "")

	lb := make(LoadBalancers, len(result.LoadBalancerDescriptions))
	for i, balancer := range result.LoadBalancerDescriptions {
		lb[i].Marshal(balancer, region, secGrpList, vpcList, subList)
	}
	*lbList = append(*lbList, lb[:]...)

	return nil
}

// Marshal parses the response from the aws sdk into an awsm LoadBalancer
func (l *LoadBalancer) Marshal(balancer *elb.LoadBalancerDescription, region string, secGrpList *SecurityGroups, vpcList *Vpcs, subList *Subnets) {

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
	l.AvailabilityZones = strings.Join(aws.StringValueSlice(balancer.AvailabilityZones), ", ") // TODO
	l.Region = region
}

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

/*
func CreateLoadBalancer(class, name, az string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	elbCfg, err := config.LoadLoadBalancerClass(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Load Balancer Class Configuration for [" + class + "]!")
	}

	// Verify the az input
	azs, errs := regions.GetAZs()
	if errs != nil {
		return errors.New("Error Verifying Availability Zone input")
	}
	if !azs.ValidAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	} else {
		terminal.Information("Found Availability Zone [" + az + "]!")
	}

	region := azs.GetRegion(az)

	svc := elb.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &elb.CreateLoadBalancerInput{
		Listeners: []*elb.Listener{ // Required
			{ // Required
				InstancePort:     aws.Int64(1),           // Required
				LoadBalancerPort: aws.Int64(1),           // Required
				Protocol:         aws.String("Protocol"), // Required
				InstanceProtocol: aws.String("Protocol"),
				SSLCertificateId: aws.String("SSLCertificateId"),
			},
			// More values...
		},
		LoadBalancerName: aws.String(name), // Required
		AvailabilityZones: []*string{
			aws.String("AvailabilityZone"), // Required
			// More values...
		},
		Scheme: aws.String("LoadBalancerScheme"),
		SecurityGroups: []*string{
			aws.String("SecurityGroupId"), // Required
			// More values...
		},
		Subnets: []*string{
			aws.String("SubnetId"), // Required
			// More values...
		},
		Tags: []*elb.Tag{
			{ // Required
				Key:   aws.String("TagKey"), // Required
				Value: aws.String("TagValue"),
			},
			// More values...
		},
	}

	createLoadBalancerResp, err := svc.CreateLoadBalancer(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Created Load Balancer [" + *createLoadBalancerResp.LoadBalancerId + "] named [" + name + "] in [" + region + "]!")

	// Add Tags
	err = SetEc2NameAndClassTags(createLoadBalancerResp.LoadBalancerId, name, class, region)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil

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
		err = GetRegionLoadBalancers(region, elbList, search, true)
	} else {
		elbList, _ = GetLoadBalancers(search, true)
	}

	if err != nil {
		return errors.New("Error gathering Load Balancer list")
	}

	if len(*elbList) > 0 {
		// Print the table
		elbList.PrintTable()
	} else {
		return errors.New("No available Load Balancers found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Load Balancers?") {
		return errors.New("Aborting!")
	}

	if !dryRun {
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
	for _, elb := range *elbList {
		svc := elb.New(session.New(&aws.Config{Region: aws.String(elb.Region)}))

		params := &ec2.DeleteLoadBalancerInput{
			LoadBalancerName: aws.String(elb.Name),
		}

		_, err := svc.DeleteLoadBalancer(params)
		if err != nil {
			return err
		}

		terminal.Delta("Deleted Load Balancer [" + elb.Name + "] in [" + elb.Region + "]!")
	}

	return nil
}
*/
