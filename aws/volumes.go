package aws

import (
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/terminal"
	"github.com/olekukonko/tablewriter"
)

type Volumes []Volume

type Volume struct {
	Name             string
	VolumeId         string
	Size             string
	State            string
	Iops             string
	Attachments      string
	CreationTime     string
	VolumeType       string
	SnapshoId        string
	DeleteOnTerm     string
	AvailabilityZone string
}

func GetVolumes() (*Volumes, error) {
	var wg sync.WaitGroup

	volList := new(Volumes)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionVolumes(region.RegionName, volList)
			if err != nil {
				terminal.ShowErrorMessage("Error gathering Volume list", err.Error())
			}
		}(region)
	}
	wg.Wait()

	return volList, nil
}

func GetRegionVolumes(region *string, volList *Volumes) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeVolumes(&ec2.DescribeVolumesInput{})

	if err != nil {
		return err
	}

	vol := make(Volumes, len(result.Volumes))
	for i, volume := range result.Volumes {

		vol[i] = Volume{
			Name:     GetTagValue("Name", volume.Tags),
			VolumeId: GetTagValue("Class", volume.Tags),
			Size:     fmt.Sprint(aws.Int64Value(volume.Size)),
			State:    aws.StringValue(volume.State),
			Iops:     fmt.Sprint(aws.Int64Value(volume.Iops)),
			//Attachments:  aws.StringValue(volume.Attachments), // TODO
			CreationTime: aws.TimeValue(volume.CreateTime).String(),
			VolumeType:   aws.StringValue(volume.VolumeType),
			SnapshoId:    aws.StringValue(volume.SnapshotId),
			//DeleteOnTerm:     aws.StringValue(volume.), // TODO
			AvailabilityZone: aws.StringValue(volume.AvailabilityZone),
		}
	}
	*volList = append(*volList, vol[:]...)

	return nil
}

func (i *Volumes) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.VolumeId,
			val.Size,
			val.State,
			val.Attachments,
			val.CreationTime,
			val.VolumeType,
			val.SnapshoId,
			val.DeleteOnTerm,
			val.AvailabilityZone,
		}
	}

	table.SetHeader([]string{"Name", "Volume Id", "Size", "State", "Attachment", "Creation Time", "Volume Type", "Snapshot Id", "Delete on Termination", "Availability Zone"})

	table.AppendBulk(rows)
	table.Render()
}
