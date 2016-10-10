package aws

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/dustin/go-humanize"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type Snapshots []Snapshot

type Snapshot models.Snapshot

func GetSnapshotsByTag(region, key, value string) (Snapshots, error) {
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

	snapList := make(Snapshots, len(result.Snapshots))
	for i, snapshot := range result.Snapshots {
		snapList[i].Marshal(snapshot, region)
	}

	if len(snapList) == 0 {
		return snapList, errors.New("No Snapshot found with tag [" + key + "] of [" + value + "] in [" + region + "].")
	}

	return snapList, err
}

func GetLatestSnapshotByTag(region, key, value string) (Snapshot, error) {
	snapshots, err := GetSnapshotsByTag(region, key, value)
	sort.Sort(snapshots)

	return snapshots[0], err
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

func GetSnapshots(search string, completed bool) (*Snapshots, []error) {
	var wg sync.WaitGroup
	var errs []error

	snapList := new(Snapshots)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionSnapshots(*region.RegionName, snapList, search, completed)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering snapshot list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return snapList, errs
}

func (s *Snapshot) Marshal(snapshot *ec2.Snapshot, region string) {
	s.Name = GetTagValue("Name", snapshot.Tags)
	s.Class = GetTagValue("Class", snapshot.Tags)
	s.Description = aws.StringValue(snapshot.Description)
	s.SnapshotId = aws.StringValue(snapshot.SnapshotId)
	s.VolumeId = aws.StringValue(snapshot.VolumeId)
	s.State = aws.StringValue(snapshot.State)
	s.StartTime = *snapshot.StartTime                                 // robots
	s.CreatedHuman = humanize.Time(aws.TimeValue(snapshot.StartTime)) // humans
	s.Progress = aws.StringValue(snapshot.Progress)
	s.VolumeSize = fmt.Sprint(aws.Int64Value(snapshot.VolumeSize))
	s.Region = region

	switch s.State {

	case "error":
		s.Progress = "failed!"

	case "completed":
		s.Progress = "ready"
	}
}

func GetRegionSnapshots(region string, snapList *Snapshots, search string, completed bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeSnapshots(&ec2.DescribeSnapshotsInput{OwnerIds: []*string{aws.String("self")}})

	if err != nil {
		return err
	}

	snap := make(Snapshots, len(result.Snapshots))
	for i, snapshot := range result.Snapshots {
		snap[i].Marshal(snapshot, region)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, in := range snap {
			rInst := reflect.ValueOf(in)

			for k := 0; k < rInst.NumField(); k++ {
				sVal := rInst.Field(k).String()

				if term.MatchString(sVal) && ((completed && snap[i].State == "completed") || !completed) {
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

func CopySnapshot(search, region string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Validate the destination region
	if !ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	// Get the source snapshot
	snapshots, _ := GetSnapshots(search, true)
	snapCount := len(*snapshots)
	if snapCount == 0 {
		return errors.New("No available snapshots found for your search terms.")
	}
	if snapCount > 1 {
		snapshots.PrintTable()
		return errors.New("Please limit your search to return only one snapshot.")
	}

	snapshot := (*snapshots)[0]

	copySnapResp, err := copySnapshot(snapshot, region, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Created Snapshot [" + *copySnapResp.SnapshotId + "] named [" + snapshot.Name + "] to [" + region + "]!")

	// Add Tags
	err = SetEc2NameAndClassTags(copySnapResp.SnapshotId, snapshot.Name, snapshot.Class, region)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil
}

func copySnapshot(snapshot Snapshot, region string, dryRun bool) (*ec2.CopySnapshotOutput, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Copy snapshot to the destination region
	params := &ec2.CopySnapshotInput{
		SourceRegion:      aws.String(snapshot.Region),
		SourceSnapshotId:  aws.String(snapshot.SnapshotId),
		Description:       aws.String(snapshot.Description),
		DestinationRegion: aws.String(region),
		DryRun:            aws.Bool(dryRun),
		//Encrypted:       aws.Bool(true),
		//KmsKeyId:        aws.String("String"),
		//PresignedUrl:    aws.String("String"),
	}
	copySnapResp, err := svc.CopySnapshot(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return copySnapResp, errors.New(awsErr.Message())
		}
	}

	return copySnapResp, err
}

func CreateSnapshot(search, class, name string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Locate the Volume
	volumes, _ := GetVolumes(search, false)
	if len(*volumes) == 0 {
		return errors.New("No volumes found for your search terms.")
	}
	if len(*volumes) > 1 {
		volumes.PrintTable()
		return errors.New("Please limit your search to return only one volume.")
	}

	volume := (*volumes)[0]
	region := volume.Region

	// Class Config
	cfg, err := config.LoadSnapshotClass(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Snapshot Class Configuration for [" + class + "]!")
	}

	// Create the snapshot
	createSnapshotResp, err := createSnapshot(volume.VolumeId, region, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Created Snapshot [" + *createSnapshotResp.SnapshotId + "] named [" + name + "] in [" + region + "]!")

	// Add Tags
	err = SetEc2NameAndClassTags(createSnapshotResp.SnapshotId, name, class, region)

	if err != nil {
		return err
	}

	sourceSnapshot := Snapshot{Name: name, Class: class, SnapshotId: *createSnapshotResp.SnapshotId}

	// Check for Propagate flag
	if cfg.Propagate && cfg.PropagateRegions != nil {

		var wg sync.WaitGroup
		var errs []error

		terminal.Information("Propagate flag is set, waiting for initial snapshot to complete.")

		// Wait for the snapshot to complete.
		err = waitForSnapshot(*createSnapshotResp.SnapshotId, region, dryRun)
		if err != nil {
			return err
		}

		// Copy to other regions
		for _, propRegion := range cfg.PropagateRegions {
			wg.Add(1)

			go func(region string) {
				defer wg.Done()

				// Copy snapshot to the destination region
				copySnapResp, err := copySnapshot(sourceSnapshot, propRegion, dryRun)

				if err != nil {
					terminal.ShowErrorMessage(fmt.Sprintf("Error propagating snapshot [%s] to region [%s]", sourceSnapshot.SnapshotId, propRegion), err.Error())
					errs = append(errs, err)
				}

				// Add Tags
				err = SetEc2NameAndClassTags(copySnapResp.SnapshotId, name, class, propRegion)
				terminal.Information(fmt.Sprintf("Copied snapshot [%s] to region [%s].", sourceSnapshot.SnapshotId, propRegion))

			}(propRegion)
		}

		wg.Wait()

		if errs != nil {
			return errors.New("Error propagating snapshot to other regions!")
		}
	}

	// Rotate out older snapshots
	if cfg.Retain > 1 {
		err := RotateSnapshots(class, cfg, dryRun)
		if err != nil {
			terminal.ShowErrorMessage(fmt.Sprintf("Error rotating [%s] snapshots!", sourceSnapshot.Class), err.Error())
			return err
		}
	}

	return nil
}

func RotateSnapshots(class string, cfg config.SnapshotClass, dryRun bool) error {
	var wg sync.WaitGroup
	var errs []error

	launchConfigs, err := GetLaunchConfigurations("")
	if err != nil {
		return errors.New("Error while retrieving the list of assets to exclude from rotation!")
	}
	excludedSnaps := launchConfigs.LockedSnapshotIds()

	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()

			// Get all the snapshots of this class in this region
			snapshots, err := GetSnapshotsByTag(*region.RegionName, "Class", class)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering snapshot list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}

			// Exclude the snapshots being used in Launch Configurations
			for i, snap := range snapshots {
				if excludedSnaps[snap.SnapshotId] {
					terminal.Information("Snapshot [" + snap.Name + " (" + snap.SnapshotId + ") ] is being used in a launch configuration, skipping!")
					snapshots = append(snapshots[:i], snapshots[i+1:]...)
				}
			}

			// Delete the oldest ones if we have more than the retention number
			if len(snapshots) > cfg.Retain {
				sort.Sort(snapshots) // important!
				ds := snapshots[cfg.Retain:]
				deleteSnapshots(&ds, dryRun)
			}

		}(region)
	}
	wg.Wait()

	if errs != nil {
		return errors.New("Error rotating snapshots for [" + class + "]!")
	}

	return nil
}

func waitForSnapshot(snapshotId, region string, dryRun bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Wait for the snapshot to complete.
	waitParams := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{aws.String(snapshotId)},
		DryRun:      aws.Bool(dryRun),
	}

	err := svc.WaitUntilSnapshotCompleted(waitParams)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
	}
	return err
}

func createSnapshot(volumeId, region string, dryRun bool) (*ec2.Snapshot, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Create the Snapshot
	snapshotParams := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(volumeId),
		DryRun:   aws.Bool(dryRun),
	}

	createSnapshotResp, err := svc.CreateSnapshot(snapshotParams)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return createSnapshotResp, errors.New(awsErr.Message())
		}
	}

	return createSnapshotResp, err
}

// Public function with confirmation terminal prompt
func DeleteSnapshots(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	snapList := new(Snapshots)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionSnapshots(region, snapList, search, false)
	} else {
		snapList, _ = GetSnapshots(search, false)
	}

	if err != nil {
		return errors.New("Error gathering Snapshots list")
	}

	if len(*snapList) > 0 {
		// Print the table
		snapList.PrintTable()
	} else {
		return errors.New("No available Snapshots found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Snapshots?") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = deleteSnapshots(snapList, dryRun)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func deleteSnapshots(snapList *Snapshots, dryRun bool) (err error) {
	for _, snapshot := range *snapList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(snapshot.Region)}))

		params := &ec2.DeleteSnapshotInput{
			SnapshotId: aws.String(snapshot.SnapshotId),
			DryRun:     aws.Bool(dryRun),
		}

		_, err := svc.DeleteSnapshot(params)
		if err != nil {
			return err
		}

		terminal.Information("Deleted Snapshot [" + snapshot.Name + "] in [" + snapshot.Region + "]!")
	}

	return nil
}

func (v Snapshots) Len() int {
	return len(v)
}

func (v Snapshots) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Snapshots) Less(i, j int) bool {
	return v[i].StartTime.After(v[j].StartTime)
}

func (i *Snapshots) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Snapshots Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, snapshot := range *i {
		models.ExtractAwsmTable(index, snapshot, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
