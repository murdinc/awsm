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
	"github.com/dustin/go-humanize"
	"github.com/murdinc/awsm/terminal"
	"github.com/olekukonko/tablewriter"
)

type Volumes []Volume

type Volume struct {
	Class            string
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

func GetVolumes(search string) (*Volumes, []error) {
	var wg sync.WaitGroup
	var errs []error

	volList := new(Volumes)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionVolumes(region.RegionName, volList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering volume list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return volList, errs
}

func GetRegionVolumes(region *string, volList *Volumes, search string) error {
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
			//CreationTime: aws.TimeValue(volume.CreateTime).String(),
			CreationTime: humanize.Time(aws.TimeValue(volume.CreateTime)),
			VolumeType:   aws.StringValue(volume.VolumeType),
			SnapshoId:    aws.StringValue(volume.SnapshotId),
			//DeleteOnTerm:     aws.StringValue(volume.), // TODO
			AvailabilityZone: aws.StringValue(volume.AvailabilityZone),
		}
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, v := range vol {
			rV := reflect.ValueOf(v)

			for k := 0; k < rV.NumField(); k++ {
				sVal := rV.Field(k).String()

				if term.MatchString(sVal) {
					*volList = append(*volList, vol[i])
					continue Loop
				}
			}
		}
	} else {
		*volList = append(*volList, vol[:]...)
	}

	return nil
}

func (i *Volumes) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Volumes Found!")
		return
	}

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
