package aws

import (
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

func GetLatestSnapshotMulti(classes []string, region string) (Snapshots, error) {
	snapList := new(Snapshots)
	matches := make(Snapshots, len(classes))

	err := GetRegionSnapshots(aws.String(region), snapList, "", false)
	if err != nil {
		return Snapshots{}, err
	}

	sort.Sort(snapList)

Loop:
	for i, c := range classes {
		for _, snap := range *snapList {
			if snap.Class == c {
				matches[i] = snap
				continue Loop
			}
		}
	}

	return matches, nil
}

func GetLatestSnapshot(class string, region string) (Snapshot, error) {
	snapList := new(Snapshots)

	err := GetRegionSnapshots(aws.String(region), snapList, class, true)
	if err != nil {
		return Snapshot{}, err
	}

	sort.Sort(snapList)

	return (*snapList)[0], nil
}

func GetSnapshots() (*Snapshots, []error) {
	var wg sync.WaitGroup
	var errs []error

	snapList := new(Snapshots)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSnapshots(region.RegionName, snapList, "", false)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering snapshot list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return snapList, errs
}

func GetRegionSnapshots(region *string, snapList *Snapshots, search string, searchClass bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeSnapshots(&ec2.DescribeSnapshotsInput{OwnerIds: []*string{aws.String("self")}})

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
			StartTime:   snapshot.StartTime.String(),
			Progress:    aws.StringValue(snapshot.Progress),
			VolumeSize:  fmt.Sprint(aws.Int64Value(snapshot.VolumeSize)),
			Region:      fmt.Sprintf(*region),
		}
	}

	if search != "" {
		if searchClass { // Specific class search
			for i, in := range snap {
				if in.Class == search {
					*snapList = append(*snapList, snap[i])
				}
			}
		} else { // General search
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
