package aws

import (
	"encoding/base64"
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
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type Instances []Instance

type Instance struct {
	Name             string
	Class            string
	PrivateIp        string
	PublicIp         string
	InstanceId       string
	AMI              string
	Root             string
	Size             string
	Virtualization   string
	State            string
	KeyPair          string
	AvailabilityZone string
	VPC              string
	VPCId            string
	Subnet           string
	SubnetId         string
	IAMUser          string
	ShutdownBehavior string
	EbsOptimized     bool // TODO
	Monitoring       bool // TODO
	Region           string
}

func GetInstances(search string, running bool) (*Instances, []error) {
	var wg sync.WaitGroup
	var errs []error

	instList := new(Instances)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionInstances(*region.RegionName, instList, search, running)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering instance list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}

	wg.Wait()

	return instList, errs
}

func (i *Instance) Marshall(instance *ec2.Instance, region string, subList *Subnets, vpcList *Vpcs) {

	subnet := subList.GetSubnetName(aws.StringValue(instance.SubnetId))
	vpc := vpcList.GetVpcName(aws.StringValue(instance.VpcId))

	i.Name = GetTagValue("Name", instance.Tags)
	i.Class = GetTagValue("Class", instance.Tags)
	i.InstanceId = aws.StringValue(instance.InstanceId)
	i.AvailabilityZone = aws.StringValue(instance.Placement.AvailabilityZone)
	i.PrivateIp = aws.StringValue(instance.PrivateIpAddress)
	i.PublicIp = aws.StringValue(instance.PublicIpAddress)
	i.AMI = aws.StringValue(instance.ImageId)
	i.Root = aws.StringValue(instance.RootDeviceType)
	i.Size = aws.StringValue(instance.InstanceType)
	i.Virtualization = aws.StringValue(instance.VirtualizationType)
	i.State = aws.StringValue(instance.State.Name)
	i.KeyPair = aws.StringValue(instance.KeyName)
	i.VPCId = aws.StringValue(instance.VpcId)
	i.VPC = vpc
	i.SubnetId = aws.StringValue(instance.SubnetId)
	i.Subnet = subnet
	i.Region = region
}

func GetRegionInstances(region string, instList *Instances, search string, running bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: &region}))
	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return err
	}

	subList := new(Subnets)
	vpcList := new(Vpcs)
	GetRegionSubnets(region, subList, "")
	GetRegionVpcs(region, vpcList, "")

	for _, reservation := range result.Reservations {
		inst := make(Instances, len(reservation.Instances))
		for i, instance := range reservation.Instances {
			inst[i].Marshall(instance, region, subList, vpcList)
		}

		if search != "" {
			term := regexp.MustCompile(search)
		Loop:
			for i, in := range inst {
				rInst := reflect.ValueOf(in)

				for k := 0; k < rInst.NumField(); k++ {
					sVal := rInst.Field(k).String()

					if term.MatchString(sVal) && ((running && inst[i].State == "running") || !running) {
						*instList = append(*instList, inst[i])
						continue Loop
					}
				}
			}
		} else {
			*instList = append(*instList, inst[:]...)
		}

	}
	return nil
}

func (i *Instances) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Instances Found!")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.Class,
			val.PrivateIp,
			val.PublicIp,
			val.InstanceId,
			val.AMI,
			val.Root,
			val.Size,
			val.Virtualization,
			val.State,
			val.KeyPair,
			val.AvailabilityZone,
			val.VPC,
			val.Subnet,
		}
	}

	table.SetHeader([]string{"Name", "Class", "Private IP", "Public IP", "Instance Id", "AMI", "Root", "Size", "Virtualization", "State", "Key Pair", "Availability Zone", "VPC", "Subnet"})

	table.AppendBulk(rows)
	table.Render()
}

func LaunchInstance(class, sequence, az string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Instance Class Config
	var instanceCfg config.InstanceClassConfig
	err := instanceCfg.LoadConfig(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Instance Class Configuration for [" + class + "]!")
	}

	// AZ
	azs, _ := GetAZs()
	if !azs.ValidAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	} else {
		terminal.Information("Found Availability Zone [" + az + "]!")
	}

	region := azs.GetRegion(az)

	// AMI Image
	ami, err := GetLatestImageByTag(region, "Class", instanceCfg.AMI)
	if err != nil {
		return err
	} else {
		terminal.Information("Found AMI [" + ami.ImageId + "] with class [" + ami.Class + "] created on [" + ami.CreationDate + "]!")
	}

	// EBS Volumes
	ebsVolumes := make([]*ec2.BlockDeviceMapping, len(instanceCfg.EBSVolumes))
	for i, ebsClass := range instanceCfg.EBSVolumes {
		var volCfg config.VolumeClassConfig
		err := volCfg.LoadConfig(ebsClass)
		if err != nil {
			return err
		} else {
			terminal.Information("Found Volume Class Configuration for [" + ebsClass + "]!")
		}

		latestSnapshot, err := GetLatestSnapshotByTag(region, "Class", volCfg.Snapshot)
		if err != nil {
			return err
		} else {
			terminal.Information("Found Snapshot [" + latestSnapshot.SnapshotId + "] with class [" + latestSnapshot.Class + "] created [" + latestSnapshot.CreatedHuman + "]!")
		}

		ebsVolumes[i] = &ec2.BlockDeviceMapping{
			DeviceName: aws.String(volCfg.DeviceName),
			Ebs: &ec2.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(volCfg.DeleteOnTermination),
				//Encrypted:           aws.Bool(volCfg.Encrypted),
				SnapshotId: aws.String(latestSnapshot.SnapshotId),
				VolumeSize: aws.Int64(int64(volCfg.VolumeSize)),

				VolumeType: aws.String(volCfg.VolumeType),
				//Iops:       aws.Int64(int64(volCfg.Iops)),
			},
			//NoDevice:    aws.String("String"),
			//VirtualName: aws.String("String"),
		}

		if volCfg.VolumeType == "io1" {
			ebsVolumes[i].Ebs.Iops = aws.Int64(int64(volCfg.Iops))
		}

	}

	// EBS Optimized
	if instanceCfg.EbsOptimized {
		terminal.Information("Launching as EBS Optimized")
	}

	// IAM Profile
	var iam IAM
	if len(instanceCfg.IAMUser) > 0 {
		iam, err := GetIAMUser(instanceCfg.IAMUser)
		if err != nil {
			return err
		} else {
			terminal.Information("Found IAM User [" + iam.UserName + "]!")
		}
	}

	// KeyPair
	keyPair, err := GetKeyPairByName(region, instanceCfg.KeyName)
	if err != nil {
		return err
	} else {
		terminal.Information("Found KeyPair [" + keyPair.KeyName + "] in [" + keyPair.Region + "]!")
	}

	// Network Interfaces

	// Placement ??

	// VPC / Subnet
	var vpc Vpc
	var subnet Subnet
	var subnetId string
	secGroupIds := make([]*string, len(instanceCfg.SecurityGroups))
	if instanceCfg.Vpc != "" && instanceCfg.Subnet != "" {
		// VPC
		vpc, err = GetVpcByTag(region, "Class", instanceCfg.Vpc)
		if err != nil {
			return err
		} else {
			terminal.Information("Found VPC [" + vpc.VpcId + "] in Region [" + region + "]!")
		}

		// Subnet
		subnet, err = vpc.GetVpcSubnetByTag("Class", instanceCfg.Subnet)
		if err != nil {
			return err
		} else {
			subnetId = subnet.SubnetId
			terminal.Information("Found Subnet [" + subnet.SubnetId + "] in VPC [" + subnet.VpcId + "]!")
		}

		// VPC Security Groups
		secGroups, err := vpc.GetVpcSecurityGroupByTagMulti("Class", instanceCfg.SecurityGroups)
		if err != nil {
			return err
		} else {
			for i, secGroup := range secGroups {
				terminal.Information("Found VPC Security Group [" + secGroup.GroupId + "] with name [" + secGroup.Name + "]!")
				secGroupIds[i] = aws.String(secGroup.GroupId)
			}
		}

	} else {
		terminal.Information("No VPC and/or Subnet specified for instance Class [" + class + "]!")

		// EC2-Classic security groups
		secGroups, err := GetSecurityGroupByTagMulti(region, "Class", instanceCfg.SecurityGroups)
		if err != nil {
			return err
		} else {
			for i, secGroup := range secGroups {
				terminal.Information("Found Security Group [" + secGroup.GroupId + "] with name [" + secGroup.Name + "]!")
				secGroupIds[i] = aws.String(secGroup.GroupId)
			}
		}
	}

	// User Data

	// ================================================================
	// ================================================================
	// ================================================================

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.RunInstancesInput{
		ImageId:             aws.String(ami.ImageId),
		MaxCount:            aws.Int64(1),
		MinCount:            aws.Int64(1),
		BlockDeviceMappings: ebsVolumes,
		DryRun:              aws.Bool(dryRun),
		EbsOptimized:        aws.Bool(instanceCfg.EbsOptimized),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn:  aws.String(iam.Arn),
			Name: aws.String(iam.UserName),
		},
		InstanceInitiatedShutdownBehavior: aws.String(instanceCfg.ShutdownBehavior),
		InstanceType:                      aws.String(instanceCfg.InstanceType),
		KeyName:                           aws.String(keyPair.KeyName),
		Monitoring: &ec2.RunInstancesMonitoringEnabled{
			Enabled: aws.Bool(instanceCfg.Monitoring),
		},
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{ // only needed when we launch with a public ip. TODO
			{
			/*
				AssociatePublicIpAddress: aws.Bool(instanceCfg.PublicIpAddress),
				DeleteOnTermination:      aws.Bool(true),
				//Description:              aws.String("String"),
				DeviceIndex: aws.Int64(0),
				Groups: []*string{
					aws.String("String"), // Required
				},

					PrivateIpAddress:   aws.String("String"),
					PrivateIpAddresses: []*ec2.PrivateIpAddressSpecification{
						{ // Required
							PrivateIpAddress: aws.String("String"), // Required
							Primary:          aws.Bool(true),
						},
					},
					SecondaryPrivateIpAddressCount: aws.Int64(1),

					SubnetId:                       aws.String("String"),
			*/
			},
		},
		/*
			Placement: &ec2.Placement{ // havent played around with placements yet, TODO?
				Affinity:         aws.String("String"),
				AvailabilityZone: aws.String("String"),
				GroupName:        aws.String("String"),
				HostId:           aws.String("String"),
				Tenancy:          aws.String("Tenancy"),
			},
		*/
		// PrivateIpAddress: aws.String("String"),
		SecurityGroupIds: secGroupIds,
		SubnetId:         aws.String(subnetId),
		UserData:         aws.String(base64.StdEncoding.EncodeToString([]byte(instanceCfg.UserData))),

		//KernelId:         aws.String("String"),
		//RamdiskId:        aws.String("String"),
	}

	launchInstanceResp, err := svc.RunInstances(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	// Add Tags
	instanceTagsParams := &ec2.CreateTagsInput{
		Resources: []*string{
			launchInstanceResp.Instances[0].InstanceId,
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(class + sequence),
			},
			{
				Key:   aws.String("Sequence"),
				Value: aws.String(sequence),
			},
			{
				Key:   aws.String("Class"),
				Value: aws.String(class),
			},
		},
		DryRun: aws.Bool(dryRun),
	}
	_, err = svc.CreateTags(instanceTagsParams)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Finished Launching Instance!")

	inst := make(Instances, 1)
	inst[1].Marshall(launchInstanceResp.Instances[0], region, &Subnets{subnet}, &Vpcs{vpc})

	inst.PrintTable()

	// ================================================================
	// ================================================================
	// ================================================================

	return nil
}

// Public function with confirmation terminal prompt
func TerminateInstances(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	instList := new(Instances)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionInstances(region, instList, search, true)
	} else {
		instList, _ = GetInstances(search, true)
	}

	if err != nil {
		return errors.New("Error gathering Instance list")
	}

	if len(*instList) > 0 {
		// Print the table
		instList.PrintTable()
	} else {
		return errors.New("No Instances found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to terminate these Instances") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = terminateInstances(instList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func terminateInstances(instList *Instances, dryRun bool) (err error) {
	for _, instance := range *instList {
		azs, _ := GetAZs()

		svc := ec2.New(session.New(&aws.Config{Region: aws.String(azs.GetRegion(instance.AvailabilityZone))}))

		params := &ec2.TerminateInstancesInput{
			InstanceIds: []*string{
				aws.String(instance.InstanceId),
			},
			DryRun: aws.Bool(dryRun),
		}

		_, err := svc.TerminateInstances(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

		terminal.Information("Terminated Instance [" + instance.InstanceId + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
	}

	return nil
}

// Public function with confirmation terminal prompt
func StopInstances(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	instList := new(Instances)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionInstances(region, instList, search, true)
	} else {
		instList, _ = GetInstances(search, true)
	}

	if err != nil {
		return errors.New("Error gathering Instance list")
	}

	if len(*instList) > 0 {
		// Print the table
		instList.PrintTable()
	} else {
		return errors.New("No Instances found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to stop these Instances") {
		return errors.New("Aborting!")
	}

	// Stop 'Em
	err = stopInstances(instList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func stopInstances(instList *Instances, dryRun bool) (err error) {
	for _, instance := range *instList {
		azs, _ := GetAZs()

		svc := ec2.New(session.New(&aws.Config{Region: aws.String(azs.GetRegion(instance.AvailabilityZone))}))

		params := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(instance.InstanceId),
			},
			DryRun: aws.Bool(dryRun),
		}

		_, err := svc.StopInstances(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

		terminal.Information("Stopped Instance [" + instance.InstanceId + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
	}

	return nil
}

// Public function with confirmation terminal prompt
func StartInstances(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	instList := new(Instances)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionInstances(region, instList, search, false)
	} else {
		instList, _ = GetInstances(search, false)
	}

	if err != nil {
		return errors.New("Error gathering Instance list")
	}

	if len(*instList) > 0 {
		// Print the table
		instList.PrintTable()
	} else {
		return errors.New("No Instances found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to start these Instances") {
		return errors.New("Aborting!")
	}

	// Stop 'Em
	err = startInstances(instList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func startInstances(instList *Instances, dryRun bool) (err error) {
	for _, instance := range *instList {
		azs, _ := GetAZs()

		svc := ec2.New(session.New(&aws.Config{Region: aws.String(azs.GetRegion(instance.AvailabilityZone))}))

		params := &ec2.StartInstancesInput{
			InstanceIds: []*string{
				aws.String(instance.InstanceId),
			},
			DryRun: aws.Bool(dryRun),
		}

		_, err := svc.StartInstances(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

		terminal.Information("Started Instance [" + instance.InstanceId + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
	}

	return nil
}

// Public function with confirmation terminal prompt
func RebootInstances(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	instList := new(Instances)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionInstances(region, instList, search, true)
	} else {
		instList, _ = GetInstances(search, true)
	}

	if err != nil {
		return errors.New("Error gathering Instance list")
	}

	if len(*instList) > 0 {
		// Print the table
		instList.PrintTable()
	} else {
		return errors.New("No Instances found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to reboot these Instances") {
		return errors.New("Aborting!")
	}

	// Stop 'Em
	err = stopInstances(instList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func rebootInstances(instList *Instances, dryRun bool) (err error) {
	for _, instance := range *instList {
		azs, _ := GetAZs()

		svc := ec2.New(session.New(&aws.Config{Region: aws.String(azs.GetRegion(instance.AvailabilityZone))}))

		params := &ec2.RebootInstancesInput{
			InstanceIds: []*string{
				aws.String(instance.InstanceId),
			},
			DryRun: aws.Bool(dryRun),
		}

		_, err := svc.RebootInstances(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

		terminal.Information("Rebooted Instance [" + instance.InstanceId + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
	}

	return nil
}
