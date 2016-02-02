package aws

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/cli"
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
	KeyPair          string
	AvailabilityZone string
	VPC              string
	Subnet           string
	State            string
}

func GetInstances() (*Instances, error) {
	var wg sync.WaitGroup

	instList := new(Instances)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionInstances(region.RegionName, instList)
			if err != nil {
				cli.ShowErrorMessage("Error gathering instance list", err.Error())
			}
		}(region)
	}

	wg.Wait()

	return instList, nil
}

func GetRegionInstances(region *string, instList *Instances) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return err
	}

	for _, reservation := range result.Reservations {
		inst := make(Instances, len(reservation.Instances))
		for i, instance := range reservation.Instances {
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
				KeyPair:          aws.StringValue(instance.KeyName),
				VPC:              aws.StringValue(instance.VpcId),
				Subnet:           aws.StringValue(instance.SubnetId),
				State:            instance.State.GoString(),
			}
		}
		*instList = append(*instList, inst[:]...)
	}
	return nil
}

func (i *Instances) PrintTable() {
	collumns := []string{"Name", "Private IP", "Public IP", "Instance Id", "AMI", "Root", "Size", "Key Pair", "Availability Zone", "VPC", "Subnet"}

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.PrivateIp,
			val.PublicIp,
			val.InstanceId,
			val.AMI,
			val.Root,
			val.Size,
			val.KeyPair,
			val.AvailabilityZone,
			val.VPC,
			val.Subnet,
		}
	}

	printTable(collumns, rows)
}
