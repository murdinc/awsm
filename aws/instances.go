package aws

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/terminal"
	"github.com/murdinc/cli"
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
}

func GetInstances(search string) (*Instances, []error) {
	var wg sync.WaitGroup
	var errs []error

	instList := new(Instances)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionInstances(region.RegionName, instList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering instance list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}

	wg.Wait()

	return instList, errs
}

func GetRegionInstances(region *string, instList *Instances, search string) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return err
	}

	subList := new(Subnets)
	vpcList := new(Vpcs)
	GetRegionSubnets(region, subList)
	GetRegionVpcs(region, vpcList)

	for _, reservation := range result.Reservations {
		inst := make(Instances, len(reservation.Instances))
		for i, instance := range reservation.Instances {

			subnet := subList.GetSubnetName(aws.StringValue(instance.SubnetId))
			vpc := vpcList.GetVpcName(aws.StringValue(instance.VpcId))

			inst[i] = Instance{
				Name:             GetTagValue("Name", instance.Tags),
				Class:            GetTagValue("Class", instance.Tags),
				InstanceId:       aws.StringValue(instance.InstanceId),
				AvailabilityZone: aws.StringValue(instance.Placement.AvailabilityZone),
				PrivateIp:        aws.StringValue(instance.PrivateIpAddress),
				PublicIp:         aws.StringValue(instance.PublicIpAddress),
				AMI:              aws.StringValue(instance.ImageId),
				Root:             aws.StringValue(instance.RootDeviceType),
				Size:             aws.StringValue(instance.InstanceType),
				Virtualization:   aws.StringValue(instance.VirtualizationType),
				State:            aws.StringValue(instance.State.Name),
				KeyPair:          aws.StringValue(instance.KeyName),
				VPCId:            aws.StringValue(instance.VpcId),
				VPC:              vpc,
				SubnetId:         aws.StringValue(instance.SubnetId),
				Subnet:           subnet,
			}
		}

		if search != "" {
			term := regexp.MustCompile(search)
		Loop:
			for i, in := range inst {
				rInst := reflect.ValueOf(in)

				for k := 0; k < rInst.NumField(); k++ {
					sVal := rInst.Field(k).String()

					if term.MatchString(sVal) {
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

	// Class Config
	var cfg config.InstanceClassConfig
	err := cfg.LoadConfig(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Instance Class Configuration for [" + class + "]!")
	}

	// AZ
	azs, _ := GetAZs()
	if !azs.ValidateAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	} else {
		terminal.Information("Found Availability Zone [" + az + "]!")
	}

	region := azs.GetRegion(az)

	// AMI Image
	ami, err := GetLatestImage(cfg.AMI, region)
	if err != nil {
		return err
	} else {
		terminal.Information("Found AMI [" + ami.ImageId + "] with class [" + ami.Class + "] created on [" + ami.CreationDate + "]!")
	}

	// EBS Volumes
	snapshots, err := GetLatestSnapshotMulti(cfg.EBSVolumes, region)
	ebsVolumes := make([]*ec2.BlockDeviceMapping, len(cfg.EBSVolumes))
	if err != nil {
		return err
	} else {
		for i, snapshot := range snapshots {
			terminal.Information("Found Snapshot [" + snapshot.SnapshotId + "] with class [" + snapshot.Class + "] created on [" + snapshot.StartTime + "]!")

			ebsVolumes[i] = &ec2.BlockDeviceMapping{
				DeviceName: aws.String("String"),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool("String"),
					Encrypted:           aws.Bool(true),
					Iops:                aws.Int64(1),
					SnapshotId:          aws.String(snapshot.SnapshotId),
					VolumeSize:          aws.Int64(1),
					VolumeType:          aws.String("VolumeType"),
				},
				NoDevice:    aws.String("String"),
				VirtualName: aws.String("String"),
			}
		}
	}

	// EBS Optimized
	if cfg.EbsOptimized {
		terminal.Information("Launching as EBS Optimized")
	}

	// IAM Profile
	var iam IAM
	if len(cfg.IAMUser) > 0 {
		iam, err := GetIAMUser(cfg.IAMUser)
		if err != nil {
			return err
		} else {
			terminal.Information("Found IAM User [" + iam.UserName + "]!")
		}
	}

	// KeyName
	/*
		keyName, err := GetKeyName(cfg.KeyName, region)
		if err != nil {
			return err
		} else {
			terminal.Information("Found Key [" + keyName. + "]!")
		}
	*/

	// Network Interfaces

	// Placement ??

	// Security Groups
	secGroups, err := GetSecurityGroupIds(cfg.SecurityGroups, region)
	if err != nil {
		return err
	} else {
		for _, secGroup := range secGroups {
			terminal.Information("Found Security Group [" + secGroup.GroupId + "] with name [" + secGroup.Name + "]!")
		}
	}

	// Subnet
	subnet, err := GetSubnetId(cfg.Subnet)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Subnet [" + subnet.SubnetId + "] in VPC [" + subnet.VpcId + "]!")
	}

	// User Data

	fmt.Println(fmt.Sprintf("%v", cfg))

	// ================================================================
	// ================================================================
	// ================================================================

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.RunInstancesInput{
		ImageId:             aws.String(ami.ImageId),
		MaxCount:            aws.Int64(1),
		MinCount:            aws.Int64(1),
		BlockDeviceMappings: ebsVolumes,
		/*
			BlockDeviceMappings: []*ec2.BlockDeviceMapping{
				{ // Required
					DeviceName: aws.String("String"),
					Ebs: &ec2.EbsBlockDevice{
						DeleteOnTermination: aws.Bool(true),
						Encrypted:           aws.Bool(true),
						Iops:                aws.Int64(1),
						SnapshotId:          aws.String("String"),
						VolumeSize:          aws.Int64(1),
						VolumeType:          aws.String("VolumeType"),
					},
					NoDevice:    aws.String("String"),
					VirtualName: aws.String("String"),
				},
				// More values...
			},
		*/
		DryRun:       aws.Bool(dryRun),
		EbsOptimized: aws.Bool(cfg.EbsOptimized),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn:  aws.String(iam.Arn),
			Name: aws.String(iam.UserName),
		},
		InstanceInitiatedShutdownBehavior: aws.String(cfg.ShutdownBehavior),
		InstanceType:                      aws.String(cfg.InstanceType),
		//KernelId:                          aws.String("String"),
		KeyName: aws.String("String"),
		Monitoring: &ec2.RunInstancesMonitoringEnabled{
			Enabled: aws.Bool(cfg.Monitoring),
		},
		NetworkInterfaces: []*ec2.InstanceNetworkInterfaceSpecification{
			{ // Required
				AssociatePublicIpAddress: aws.Bool(cfg.PublicIpAddress),
				DeleteOnTermination:      aws.Bool(true),
				Description:              aws.String("String"),
				DeviceIndex:              aws.Int64(1),
				Groups: []*string{
					aws.String("String"), // Required
					// More values...
				},
				NetworkInterfaceId: aws.String("String"),
				PrivateIpAddress:   aws.String("String"),
				PrivateIpAddresses: []*ec2.PrivateIpAddressSpecification{
					{ // Required
						PrivateIpAddress: aws.String("String"), // Required
						Primary:          aws.Bool(true),
					},
					// More values...
				},
				SecondaryPrivateIpAddressCount: aws.Int64(1),
				SubnetId:                       aws.String("String"),
			},
			// More values...
		},
		/*
			Placement: &ec2.Placement{
				Affinity:         aws.String("String"),
				AvailabilityZone: aws.String("String"),
				GroupName:        aws.String("String"),
				HostId:           aws.String("String"),
				Tenancy:          aws.String("Tenancy"),
			},
		*/
		PrivateIpAddress: aws.String("String"),
		RamdiskId:        aws.String("String"),
		SecurityGroupIds: []*string{
			aws.String("String"), // Required
			// More values...
		},
		SecurityGroups: []*string{
			aws.String("String"), // Required
			// More values...
		},
		SubnetId: aws.String(subnet.SubnetId),
		UserData: aws.String("String"),
	}

	fmt.Println(params)

	resp, err := svc.RunInstances(params)

	if err != nil {
		return err
	}

	// Pretty-print the response data.
	fmt.Println(resp)

	// ================================================================
	// ================================================================
	// ================================================================

	return nil
}
