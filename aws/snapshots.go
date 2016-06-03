package aws

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/terminal"
	"github.com/olekukonko/tablewriter"
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

func GetLatestSnapshotByTag(region, key, value string) (Snapshot, error) {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DescribeSnapshotsInput{
		OwnerIds: []*string{aws.String("self")},
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + key),
				Values: []*string{
					aws.String(value),
				},
			},
		},
	}

	result, err := svc.DescribeSnapshots(params)

	if err != nil {
		return Snapshot{}, err
	}

	count := len(result.Snapshots)

	switch count {
	case 0:
		return Snapshot{}, errors.New("No Snapshot found with tag [" + key + "] of [" + value + "] in [" + region + "], Aborting!")
	case 1:
		snapshot := new(Snapshot)
		snapshot.Marshall(result.Snapshots[0], region)
		return *snapshot, nil
	}

	snapList := make(Snapshots, len(result.Snapshots))
	for i, snapshot := range result.Snapshots {
		snapList[i].Marshall(snapshot, region)
	}

	sort.Sort(snapList)

	return snapList[0], nil
}

func GetLatestSnapshotByTagMulti(region, key string, value []string) (Snapshots, error) {
	var snapList Snapshots
	for _, v := range value {
		snapshot, err := GetLatestSnapshotByTag(region, key, v)
		if err != nil {
			return Snapshots{}, err
		}

		snapList = append(snapList, snapshot)
	}

	return snapList, nil
}

func GetSnapshots(search string) (*Snapshots, []error) {
	var wg sync.WaitGroup
	var errs []error

	snapList := new(Snapshots)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSnapshots(*region.RegionName, snapList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering snapshot list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return snapList, errs
}

func (s *Snapshot) Marshall(snapshot *ec2.Snapshot, region string) {
	s.Name = GetTagValue("Name", snapshot.Tags)
	s.Class = GetTagValue("Class", snapshot.Tags)
	s.Description = aws.StringValue(snapshot.Description)
	s.SnapshotId = aws.StringValue(snapshot.SnapshotId)
	s.VolumeId = aws.StringValue(snapshot.VolumeId)
	s.State = aws.StringValue(snapshot.State)
	s.StartTime = snapshot.StartTime.String()
	s.Progress = aws.StringValue(snapshot.Progress)
	s.VolumeSize = fmt.Sprint(aws.Int64Value(snapshot.VolumeSize))
	s.Region = region
}

func GetRegionSnapshots(region string, snapList *Snapshots, search string) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeSnapshots(&ec2.DescribeSnapshotsInput{OwnerIds: []*string{aws.String("self")}})

	if err != nil {
		return err
	}

	snap := make(Snapshots, len(result.Snapshots))
	for i, snapshot := range result.Snapshots {
		snap[i].Marshall(snapshot, region)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, in := range snap {
			rInst := reflect.ValueOf(in)

			for k := 0; k < rInst.NumField(); k++ {
				sVal := rInst.Field(k).String()

				if term.MatchString(sVal) {
					*snapList = append(*snapList, snap[i])
					continue Loop
				}
			}
		}
	} else {
		*snapList = append(*snapList, snap[:]...)
	}

	return nil
}

// Functions for sorting
func (s *Snapshot) Timestamp() time.Time {
	timestamp, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", s.StartTime)
	if err != nil {
		terminal.ErrorLine("Error parsing the timestamp for volume [" + s.SnapshotId + "]!")
	}

	return timestamp
}

func (v Snapshots) Len() int {
	return len(v)
}

func (v Snapshots) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Snapshots) Less(i, j int) bool {
	return v[i].Timestamp().After(v[j].Timestamp())
}

func (i *Snapshots) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

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

	table.SetHeader([]string{"Name", "Class", "Description", "Snapshot Id", "Volume Id", "State", "Start Time", "Progress", "Volume Size", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
