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
	humanize "github.com/dustin/go-humanize"
	"github.com/hashicorp/hil"
	"github.com/hashicorp/hil/ast"
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// Instances represents a slice of EC2 Instances
type Instances []Instance

// Instance represents a single EC2 Instace
type Instance models.Instance

// GetInstances returns a list of EC2 Instances that match the provided search term and optional running flag
func GetInstances(search string, running bool) (*Instances, []error) {
	var wg sync.WaitGroup
	var errs []error

	instList := new(Instances)
	regions := GetRegionListWithoutIgnored()

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

// GetInstanceName returns the the name of an EC2 Instance given an EC2 Instance ID
func (i *Instances) GetInstanceName(id string) string {
	for _, instance := range *i {
		if instance.InstanceID == id && instance.Name != "" {
			return instance.Name
		}
	}
	return id
}

// GetInstanceClass returns the the class of an EC2 Instance given an EC2 Instance ID
func (i *Instances) GetInstanceClass(id string) string {
	for _, instance := range *i {
		if instance.InstanceID == id && instance.Name != "" {
			return instance.Class
		}
	}
	return id
}

// Marshal parses the response from the aws sdk into an awsm Instance
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

	if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
		i.IamInstanceProfileArn = aws.StringValue(instance.IamInstanceProfile.Arn)
		iamInstanceProfileName, _ := ParseArn(i.IamInstanceProfileArn)
		i.IamInstanceProfileName = iamInstanceProfileName.ProfileName
	}

	// TODO
	//instance.SecurityGroups
}

// GetRegionInstances returns a slice of Instances into the passed Instances slice based on the provided region and search term, and optional running flag
func GetRegionInstances(region string, instList *Instances, search string, running bool) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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
			if running {
				for i, _ := range inst {
					if inst[i].State == "running" {
						*instList = append(*instList, inst[i])
					}
				}
			} else {
				*instList = append(*instList, inst[:]...)
			}
		}
	}
	return nil
}

// PrintTable Prints an ascii table of the list of Instances
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

// LaunchInstance Launches a new EC2 Instance
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
	azs, _ := regions.GetAZs()
	if !azs.ValidAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	}

	terminal.Information("Found Availability Zone [" + az + "]!")

	region := azs.GetRegion(az)

	// AMI
	var ami Image

	if instanceCfg.AMI == "" {
		amiId := terminal.PromptString("There is no AMI class configured for this Instance class, please provide an AMI to use:")
		ami, err = GetImageById(region, amiId)
		if err != nil {
			return err
		}

		terminal.Information("Found AMI [" + ami.ImageID + "] with name [" + ami.AmiName + "] created [" + humanize.Time(ami.CreationDate) + "]!")

	} else {
		ami, err = GetLatestImageByTag(region, "Class", instanceCfg.AMI)
		if err != nil {
			return err
		}

		terminal.Information("Found AMI [" + ami.ImageID + "] with class [" + ami.Class + "] created [" + humanize.Time(ami.CreationDate) + "]!")
	}

	// EBS
	ebsVolumes := make([]*ec2.BlockDeviceMapping, len(instanceCfg.EBSVolumes))
	ebsVolumeNames := make(map[string]string)
	ebsVolumeClasses := make(map[string]string)
	for i, ebsClass := range instanceCfg.EBSVolumes {
		volCfg, err := config.LoadVolumeClass(ebsClass)
		if err != nil {
			return err
		}

		terminal.Information("Found Volume Class Configuration for [" + ebsClass + "]!")

		var snapshotId string
		if volCfg.Snapshot == "" {
			terminal.Information("No snapshot configured for [" + ebsClass + "]! Creating a fresh volume instead.")

		} else {
			latestSnapshot, err := GetLatestSnapshotByTag(region, "Class", volCfg.Snapshot)
			if err != nil {
				return err
			} else {
				terminal.Information("Found Snapshot [" + latestSnapshot.SnapshotID + "] with class [" + latestSnapshot.Class + "] created [" + humanize.Time(latestSnapshot.StartTime) + "]!")
				snapshotId = latestSnapshot.SnapshotID
			}
		}

		ebsVolumes[i] = &ec2.BlockDeviceMapping{
			DeviceName: aws.String(volCfg.DeviceName),
			Ebs: &ec2.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(volCfg.DeleteOnTermination),
				Encrypted:           aws.Bool(volCfg.Encrypted),
				VolumeSize:          aws.Int64(int64(volCfg.VolumeSize)),
				VolumeType:          aws.String(volCfg.VolumeType),
			},
			//NoDevice:    aws.String("String"),
			//VirtualName: aws.String("String"),
		}

		if volCfg.VolumeType == "io1" {
			ebsVolumes[i].Ebs.Iops = aws.Int64(int64(volCfg.Iops))
		}

		if snapshotId != "" {
			ebsVolumes[i].Ebs.SnapshotId = aws.String(snapshotId)
			ebsVolumes[i].Ebs.Encrypted = nil // You cannot specify the encrypted flag if specifying a snapshot id in a block device mapping
		}

		ebsVolumeNames[volCfg.DeviceName] = class + sequence + "-" + ebsClass
		ebsVolumeClasses[volCfg.DeviceName] = ebsClass

	}

	// EBS Optimized
	if instanceCfg.EbsOptimized {
		terminal.Information("Launching as EBS Optimized")
	}

	// IAM Instance Profile
	var iam IAMInstanceProfile
	if len(instanceCfg.IAMInstanceProfile) > 0 {
		iam, err = GetIAMInstanceProfile(instanceCfg.IAMInstanceProfile)
		if err != nil {
			return err
		}

		terminal.Information("Found IAM Instance Profile [" + iam.ProfileName + "]!")
	}

	// KeyPair
	keyPair, err := GetKeyPairByName(region, instanceCfg.KeyName)
	if err != nil {
		// Try to create it?
		terminal.Information("Unable to find KeyPair [" + instanceCfg.KeyName + "] in [" + region + "], trying to create it...")

		err = CreateKeyPair(instanceCfg.KeyName, region, dryRun)
		if err != nil {
			return err
		}

		keyPair, err = GetKeyPairByName(region, instanceCfg.KeyName)
		if err != nil {
			return err
		}
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
		vpc, err = GetRegionVpcByTag(region, "Class", instanceCfg.Vpc)
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

	// Parse Userdata
	tree, err := hil.Parse(instanceCfg.UserData)
	if err != nil {
		return err
	}

	config := &hil.EvalConfig{
		GlobalScope: &ast.BasicScope{
			VarMap: map[string]ast.Variable{
				"var.class": {
					Type:  ast.TypeString,
					Value: class,
				},
				"var.sequence": {
					Type:  ast.TypeString,
					Value: sequence,
				},
				"var.locale": {
					Type:  ast.TypeString,
					Value: region,
				},
			},
		},
	}

	result, err := hil.Eval(tree, config)
	if err != nil {
		return err
	}

	parsedUserData := result.Value.(string)

	if dryRun {
		terminal.Notice("User Data:")
		fmt.Println(parsedUserData)
	}

	params := &ec2.RunInstancesInput{
		ImageId:      aws.String(ami.ImageID),
		MaxCount:     aws.Int64(1),
		MinCount:     aws.Int64(1),
		DryRun:       aws.Bool(dryRun),
		EbsOptimized: aws.Bool(instanceCfg.EbsOptimized),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(iam.Arn),
		},
		InstanceInitiatedShutdownBehavior: aws.String(instanceCfg.ShutdownBehavior),
		InstanceType:                      aws.String(instanceCfg.InstanceType),
		KeyName:                           aws.String(keyPair.KeyName),
		Monitoring: &ec2.RunInstancesMonitoringEnabled{
			Enabled: aws.Bool(instanceCfg.Monitoring),
		},
		UserData: aws.String(base64.StdEncoding.EncodeToString([]byte(parsedUserData))),
		/*
			Placement: &ec2.Placement{ // havent played around with placements yet, TODO?
				Affinity:         aws.String("String"),
				AvailabilityZone: aws.String("String"),
				GroupName:        aws.String("String"),
				HostId:           aws.String("String"),
				Tenancy:          aws.String("Tenancy"),
			},
			PrivateIpAddress: aws.String("String"),
			KernelId:         aws.String("String"),
			RamdiskId:        aws.String("String"),
		*/
	}

	if instanceCfg.PublicIPAddress {
		params.SetNetworkInterfaces([]*ec2.InstanceNetworkInterfaceSpecification{
			{
				AssociatePublicIpAddress: aws.Bool(true),
				DeleteOnTermination:      aws.Bool(true), // TODO link up to cfg
				DeviceIndex:              aws.Int64(0),
				SubnetId:                 aws.String(subnetID),
				Groups:                   secGroupIds,
			}})
	} else {
		params.SetSubnetId(subnetID)
		params.SetSecurityGroupIds(secGroupIds)
	}

	if len(ebsVolumes) > 0 {
		params.BlockDeviceMappings = ebsVolumes
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	if dryRun {
		terminal.Notice("Params:")
		fmt.Println(params.String())
	}

	launchInstanceResp, err := svc.RunInstances(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
	}

	instance := launchInstanceResp.Instances[0]

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Delta("Launching Instance:")

	inst := make(Instances, 1)
	inst[0].Marshal(instance, region, &Subnets{subnet}, &Vpcs{vpc}, &Images{ami})

	inst[0].Name = class + sequence
	inst[0].Class = class
	inst[0].AMIName = ami.ImageID

	inst.PrintTable()

	terminal.Notice("Waiting to tag Instance...")

	// Wait to tag it
	err = svc.WaitUntilInstanceExists(&ec2.DescribeInstancesInput{
		DryRun: aws.Bool(dryRun),
		InstanceIds: []*string{
			launchInstanceResp.Instances[0].InstanceId,
		},
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
	}

	terminal.Delta("Adding EC2 Tags...")

	// Add Instance Tags
	instanceTagsParams := &ec2.CreateTagsInput{
		Resources: []*string{instance.InstanceId},
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

	if len(ebsVolumes) > 0 {
		terminal.Notice("Waiting to tag EBS Volumes...")

		// Wait to tag it
		err = svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
			DryRun: aws.Bool(dryRun),
			InstanceIds: []*string{
				launchInstanceResp.Instances[0].InstanceId,
			},
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
		}

		terminal.Delta("Adding EBS Tags...")

		// Add EBS Volume Tags
		ebsVols, err := GetVolumesByInstanceID(region, aws.StringValue(instance.InstanceId))
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return errors.New(awsErr.Message())
			}
			return err
		}

		for _, ebsVol := range ebsVols {
			if ebsVolumeNames[ebsVol.Device] != "" || ebsVolumeClasses[ebsVol.Device] != "" {
				// Add Tags
				err = SetEc2NameAndClassTags(aws.String(ebsVol.VolumeID), ebsVolumeNames[ebsVol.Device], ebsVolumeClasses[ebsVol.Device], region)
				if err != nil {
					if awsErr, ok := err.(awserr.Error); ok {
						return errors.New(awsErr.Message())
					}
					return err
				}
			}
		}
	}

	terminal.Information("Finished Launching Instance!")

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
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(instance.Region)}))
		svc := ec2.New(sess)

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

		terminal.Delta("Terminated Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
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

	for _, instance := range *instList {

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(instance.Region)}))
		svc := ec2.New(sess)

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

		terminal.Delta("Stopped Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
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

	for _, instance := range *instList {

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(instance.Region)}))
		svc := ec2.New(sess)

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

		terminal.Delta("Started Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
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

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(instance.Region)}))
		svc := ec2.New(sess)

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

		terminal.Delta("Rebooted Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + instance.AvailabilityZone + "]!")
	}

	return nil
}
