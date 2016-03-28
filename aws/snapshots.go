package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/cli"
)

type Snapshots []Snapshot

type Snapshot struct {
	Name        string
	Class       string
	Description string
	SnapshotId  string
	VolumeId    string
	State       string
	StartTime   string
	Progress    string
	VolumeSize  string
	Region      string
}

func GetSnapshots() (*Snapshots, error) {
	var wg sync.WaitGroup

	snapList := new(Snapshots)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSnapshots(region.RegionName, snapList)
			if err != nil {
				cli.ShowErrorMessage("Error gathering Snapshot list", err.Error())
			}
		}(region)
	}
	wg.Wait()

	return snapList, nil
}

func GetRegionSnapshots(region *string, snapList *Snapshots) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeSnapshots(&ec2.DescribeSnapshotsInput{})

	if err != nil {
		return err
	}

	snap := make(Snapshots, len(result.Snapshots))
	for i, snapshot := range result.Snapshots {
		snap[i] = Snapshot{
			Name:        GetTagValue("Name", snapshot.Tags),
			Class:       GetTagValue("Class", snapshot.Tags),
			Description: aws.StringValue(snapshot.Description),
			SnapshotId:  aws.StringValue(snapshot.SnapshotId),
			VolumeId:    aws.StringValue(snapshot.VolumeId),
			State:       aws.StringValue(snapshot.State),
			StartTime:   aws.TimeValue(snapshot.StartTime).String(),
			Progress:    aws.StringValue(snapshot.Progress),
			VolumeSize:  string(aws.Int64Value(snapshot.VolumeSize)),
			Region:      fmt.Sprintf(*region),
		}
	}
	*snapList = append(*snapList, snap[:]...)

	return nil
}

func (i *Snapshots) PrintTable() {
	collumns := []string{"Name", "Class", "Description", "Snapshot Id", "Volume Id", "State", "Start Time", "Progress", "Volume Size", "Region"}

	// "Name", "Class", "Description", "Snapshot Id", "Volume Id", "State", "Start Time", "Progress", "Volume Size", "Region"

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.Class,
			val.Description,
			val.SnapshotId,
			val.VolumeId,
			val.State,
			val.StartTime,
			val.Progress,
			val.VolumeSize,
			val.Region,
		}
	}

	printTable(collumns, rows)
}
