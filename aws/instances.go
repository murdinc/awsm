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
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type Instances []Instance

type Instance models.Instance

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

func (i *Instances) GetInstanceName(id string) string {
	for _, instance := range *i {
		if instance.InstanceID == id && instance.Name != "" {
			return instance.Name
		}
	}
	return id
}

func (i *Instance) Marshal(instance *ec2.Instance, region string, subList *Subnets, vpcList *Vpcs, imgList *Images) {

	subnet := subList.GetSubnetName(aws.StringValue(instance.SubnetId))
	vpc := vpcList.GetVpcName(aws.StringValue(instance.VpcId))

	i.Name = GetTagValue("Name", instance.Tags)
	i.Class = GetTagValue("Class", instance.Tags)
	i.InstanceID = aws.StringValue(instance.InstanceId)
	i.AvailabilityZone = aws.StringValue(instance.Placement.AvailabilityZone)
	i.PrivateIP = aws.StringValue(instance.PrivateIpAddress)
	i.PublicIP = aws.StringValue(instance.PublicIpAddress)
	i.AMIID = aws.StringValue(instance.ImageId)
	i.AMIName = imgList.GetImageName(i.AMIID)
	i.Root = aws.StringValue(instance.RootDeviceType)
	i.Size = aws.StringValue(instance.InstanceType)
	i.Virtualization = aws.StringValue(instance.VirtualizationType)
	i.State = aws.StringValue(instance.State.Name)
	i.KeyPair = aws.StringValue(instance.KeyName)
	i.VPCID = aws.StringValue(instance.VpcId)
	i.VPC = vpc
	i.SubnetID = aws.StringValue(instance.SubnetId)
	i.Subnet = subnet
	i.Region = region

	// TODO
	//instance.SecurityGroups
}

func GetRegionInstances(region string, instList *Instances, search string, running bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: &region}))
	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return err
	}

	subList := new(Subnets)
	vpcList := new(Vpcs)
	imgList := new(Images)
	GetRegionSubnets(region, subList, "")
	GetRegionVpcs(region, vpcList, "")
	GetRegionImages(region, imgList, "", false)

	for _, reservation := range result.Reservations {
		inst := make(Instances, len(reservation.Instances))
		for i, instance := range reservation.Instances {
			inst[i].Marshal(instance, region, subList, vpcList, imgList)
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

	var header []string
	rows := make([][]string, len(*i))

	for index, instance := range *i {
		models.ExtractAwsmTable(index, instance, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

func LaunchInstance(class, sequence, az string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Instance Class Config
	instanceCfg, err := config.LoadInstanceClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found Instance class configuration for [" + class + "]!")

	// AZ
	azs, _ := GetAZs()
	if !azs.ValidAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	}

	terminal.Information("Found Availability Zone [" + az + "]!")

	region := azs.GetRegion(az)

	// AMI
	ami, err := GetLatestImageByTag(region, "Class", instanceCfg.AMI)
	if err != nil {
		return err
	}

	terminal.Information("Found AMI [" + ami.ImageID + "] with class [" + ami.Class + "] created [" + ami.CreatedHuman + "]!")

	// EBS
	ebsVolumes := make([]*ec2.BlockDeviceMapping, len(instanceCfg.EBSVolumes))
	for i, ebsClass := range instanceCfg.EBSVolumes {
		volCfg, err := config.LoadVolumeClass(ebsClass)
		if err != nil {
			return err
		}

		terminal.Information("Found Volume Class Configuration for [" + ebsClass + "]!")

		latestSnapshot, err := GetLatestSnapshotByTag(region, "Class", volCfg.Snapshot)
		if err != nil {
			return err
		}

		terminal.Information("Found Snapshot [" + latestSnapshot.SnapshotID + "] with class [" + latestSnapshot.Class + "] created [" + latestSnapshot.CreatedHuman + "]!")

		ebsVolumes[i] = &ec2.BlockDeviceMapping{
			DeviceName: aws.String(volCfg.DeviceName),
			Ebs: &ec2.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(volCfg.DeleteOnTermination),
				//Encrypted:           aws.Bool(volCfg.Encrypted),
				SnapshotId: aws.String(latestSnapshot.SnapshotID),
				VolumeSize: aws.Int64(int64(volCfg.VolumeSize)),
				VolumeType: aws.String(volCfg.VolumeType),
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
	var iam IAMUser
	if len(instanceCfg.IAMUser) > 0 {
		iam, err := GetIAMUser(instanceCfg.IAMUser)
		if err != nil {
			return err
		}

		terminal.Information("Found IAM User [" + iam.UserName + "]!")

	}

	// KeyPair
	keyPair, err := GetKeyPairByName(region, instanceCfg.KeyName)
	if err != nil {
		return err
	}

	terminal.Information("Found KeyPair [" + keyPair.KeyName + "] in [" + keyPair.Region + "]!")

	// Network Interfaces

	// Placement ??

	// VPC / Subnet
	var vpc Vpc
	var subnet Subnet
	var subnetID string
	secGroupIds := make([]*string, len(instanceCfg.SecurityGroups))
	if instanceCfg.Vpc != "" && instanceCfg.Subnet != "" {
		// VPC
		vpc, err = GetVpcByTag(region, "Class", instanceCfg.Vpc)
		if err != nil {
			return err
		}

		terminal.Information("Found VPC [" + vpc.VpcID + "] in Region [" + region + "]!")

		// Subnet
		subnet, err = vpc.GetVpcSubnetByTag("Class", instanceCfg.Subnet)
		if err != nil {
			return err
		}

		subnetID = subnet.SubnetID
		terminal.Information("Found Subnet [" + subnet.SubnetID + "] in VPC [" + subnet.VpcID + "]!")

		// VPC Security Groups
		secGroups, err := vpc.GetVpcSecurityGroupByTagMulti("Class", instanceCfg.SecurityGroups)
		if err != nil {
			return err
		}

		for i, secGroup := range secGroups {
			terminal.Information("Found VPC Security Group [" + secGroup.GroupID + "] with name [" + secGroup.Name + "]!")
			secGroupIds[i] = aws.String(secGroup.GroupID)
		}

	} else {
		terminal.Information("No VPC and/or Subnet specified for instance Class [" + class + "]!")

		// EC2-Classic security groups
		secGroups, err := GetSecurityGroupByTagMulti(region, "Class", instanceCfg.SecurityGroups)
		if err != nil {
			return err
		}

		for i, secGroup := range secGroups {
			terminal.Information("Found Security Group [" + secGroup.GroupID + "] with name [" + secGroup.Name + "]!")
			secGroupIds[i] = aws.String(secGroup.GroupID)
		}

	}

	// User Data

	// ================================================================
	// ================================================================
	// ================================================================

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.RunInstancesInput{
		ImageId:             aws.String(ami.ImageID),
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
		SubnetId:         aws.String(subnetID),
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
	inst[1].Marshal(launchInstanceResp.Instances[0], region, &Subnets{subnet}, &Vpcs{vpc}, &Images{ami})

	inst.PrintTable()

	// ================================================================
	// ================================================================
	// ================================================================

	return nil
}

// TerminateInstances terminates EC2 instances based on the given search term and optional region input
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
		return errors.New("No Instances found!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to terminate these Instances?") {
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
				aws.String(instance.InstanceID),
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

		terminal.Information("Terminated Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
	}

	return nil
}

// StopInstances stops an EC2 instances based on the given search term and optional region input. A third option "force" will force stop the instance(s)
func StopInstances(search, region string, force, dryRun bool) (err error) {

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
	if !terminal.PromptBool("Are you sure you want to stop these Instances?") {
		return errors.New("Aborting!")
	}

	// Stop 'Em
	err = stopInstances(instList, force, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func stopInstances(instList *Instances, force, dryRun bool) (err error) {
	azs, _ := GetAZs()

	for _, instance := range *instList {

		svc := ec2.New(session.New(&aws.Config{Region: aws.String(azs.GetRegion(instance.AvailabilityZone))}))

		params := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(instance.InstanceID),
			},
			Force:  aws.Bool(force),
			DryRun: aws.Bool(dryRun),
		}

		_, err := svc.StopInstances(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

		terminal.Information("Stopped Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
	}

	return nil
}

// StartInstances starts one or more instances based on the given search term and optional region
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
	if !terminal.PromptBool("Are you sure you want to start these Instances?") {
		return errors.New("Aborting!")
	}

	// Start 'Em
	err = startInstances(instList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func startInstances(instList *Instances, dryRun bool) (err error) {

	azs, _ := GetAZs()

	for _, instance := range *instList {

		svc := ec2.New(session.New(&aws.Config{Region: aws.String(azs.GetRegion(instance.AvailabilityZone))}))

		params := &ec2.StartInstancesInput{
			InstanceIds: []*string{
				aws.String(instance.InstanceID),
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

		terminal.Information("Started Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
	}

	return nil
}

// RebootInstances reboots one or more instances based on the given search term an optional region input
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
	if !terminal.PromptBool("Are you sure you want to reboot these Instances?") {
		return errors.New("Aborting!")
	}

	// Reboot 'Em
	err = rebootInstances(instList, dryRun)
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
				aws.String(instance.InstanceID),
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

		terminal.Information("Rebooted Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
	}

	return nil
}
