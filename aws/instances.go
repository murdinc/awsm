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
}

func GetInstances(search string) (*Instances, error) {
	var wg sync.WaitGroup

	instList := new(Instances)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionInstances(region.RegionName, instList, search)
			if err != nil {
				terminal.ShowErrorMessage("Error gathering instance list", err.Error())
			}
		}(region)
	}

	wg.Wait()

	return instList, nil
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

	// Check --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the class config
	config, err := config.GetClassConfig(class, "ec2")
	if err != nil {
		return err
	} else {
		terminal.Information("Found Instance Class Configuration for [" + class + "]!")
	}

	// Check AZ
	azs, _ := GetAZs()
	if !azs.ValidateAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	} else {
		terminal.Information("Found Availability Zone [" + az + "]!")
	}

	fmt.Println(config)

	return nil
}

func GetInstanceClassConfig(class string) (*config.InstanceClassConfig, error) {

	return &config.InstanceClassConfig{}, nil
}
