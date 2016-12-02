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
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// Snapshots represents a slice of EBS Snapshots
type Snapshots []Snapshot

// Snapshot represents a single EBS Snapshot
type Snapshot models.Snapshot

// GetSnapshotsByTag returns a slice of EBS Snapshots that match the provided region and Tag key/value
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
		return snapList, nil //errors.New("No Snapshot found with tag [" + key + "] of [" + value + "] in [" + region + "].")
	}

	return snapList, err
}

// GetLatestSnapshotByTag returns the newest EBS Snapshot that matches the provided region and Tag key/value
func GetLatestSnapshotByTag(region, key, value string) (Snapshot, error) {
	snapshots, err := GetSnapshotsByTag(region, key, value)
	if err != nil {
		return Snapshot{}, err
	}

	sort.Sort(snapshots)

	return snapshots[0], err
}

// GetLatestSnapshotByTagMulti returns a slice of the newest EBS Snapshot that matches the provided region and Tag key/value. Multiple Tag values can be provided
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

// GetSnapshots returns a slice of EBS Snapshots that match the provided search term and optional completed flag
func GetSnapshots(search string, completed bool) (*Snapshots, []error) {
	var wg sync.WaitGroup
	var errs []error

	snapList := new(Snapshots)
	regions := regions.GetRegionList()

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

// Marshal parses the response from the aws sdk into an awsm Snapshot
func (s *Snapshot) Marshal(snapshot *ec2.Snapshot, region string) {
	s.Name = GetTagValue("Name", snapshot.Tags)
	s.Class = GetTagValue("Class", snapshot.Tags)
	s.Description = aws.StringValue(snapshot.Description)
	s.SnapshotID = aws.StringValue(snapshot.SnapshotId)
	s.VolumeID = aws.StringValue(snapshot.VolumeId)
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

// GetRegionSnapshots returns a list of a regions Snapshots into the provided Snapshots slice that match the provided search term and optional completed flag
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

// CopySnapshot copies a Snapshot to another region
func CopySnapshot(search, region string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Validate the destination region
	if !regions.ValidRegion(region) {
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

	terminal.Delta("Created Snapshot [" + *copySnapResp.SnapshotId + "] named [" + snapshot.Name + "] to [" + region + "]!")

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

// private function without terminal prompts
func copySnapshot(snapshot Snapshot, region string, dryRun bool) (*ec2.CopySnapshotOutput, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Copy snapshot to the destination region
	params := &ec2.CopySnapshotInput{
		SourceSnapshotId: aws.String(snapshot.SnapshotID),
		Description:      aws.String(snapshot.Description),
		SourceRegion:     aws.String(snapshot.Region),
		DryRun:           aws.Bool(dryRun),
		//Encrypted:         aws.Bool(true),
		//KmsKeyId:          aws.String("String"),
		//PresignedUrl:      aws.String("String"),
		//DestinationRegion: aws.String(region), // only needed when using presigned url, bombs otherwise

	}

	copySnapResp, err := svc.CopySnapshot(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return copySnapResp, errors.New(awsErr.Message())
		}
	}

	return copySnapResp, err
}

// CreateSnapshot creates a new EBS Snapshot
func CreateSnapshot(class, name string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	cfg, err := config.LoadSnapshotClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found Snapshot Class Configuration for [" + class + "]!")

	if cfg.Volume == "" {
		return errors.New("Snapshot Class [" + class + "] does not have a source Volume specified yet, Aborting!")
	}

	// Locate the Volume
	volumes, _ := GetVolumes(cfg.Volume, false)
	if len(*volumes) == 0 {
		return errors.New("No volumes found matching [" + cfg.Volume + "], Aborting!")
	}
	if len(*volumes) > 1 {
		volumes.PrintTable()
		return errors.New("Found more than one volume matching [" + cfg.Volume + "], Aborting!")
	}

	volume := (*volumes)[0]
	region := volume.Region

	// Create the snapshot
	createSnapshotResp, err := createSnapshot(volume.VolumeID, cfg.Description, region, dryRun)
	if err != nil {
		return err
	}

	terminal.Delta("Created Snapshot [" + *createSnapshotResp.SnapshotId + "] named [" + name + "] in [" + region + "]!")

	// Add Tags
	err = SetEc2NameAndClassTags(createSnapshotResp.SnapshotId, name, class, region)

	if err != nil {
		return err
	}

	sourceSnapshot := Snapshot{Name: name, Class: class, SnapshotID: *createSnapshotResp.SnapshotId, Region: region, Description: cfg.Description}

	// Check for Propagate flag
	if cfg.Propagate && cfg.PropagateRegions != nil {

		var wg sync.WaitGroup
		var errs []error

		terminal.Notice("Propagate flag is set, waiting for initial snapshot to complete...")

		// Wait for the snapshot to complete.
		err = waitForSnapshot(*createSnapshotResp.SnapshotId, region, dryRun)
		if err != nil {
			return err
		}

		// Copy to other regions
		for _, propRegion := range cfg.PropagateRegions {

			if propRegion != region {

				wg.Add(1)
				go func(propRegion string) {
					defer wg.Done()

					// Copy snapshot to the destination region
					copySnapResp, err := copySnapshot(sourceSnapshot, propRegion, dryRun)

					if err != nil {
						terminal.ShowErrorMessage(fmt.Sprintf("Error propagating snapshot [%s] to region [%s]", sourceSnapshot.SnapshotID, propRegion), err.Error())
						errs = append(errs, err)
					} else {
						// Add Tags
						SetEc2NameAndClassTags(copySnapResp.SnapshotId, name, class, propRegion)
						terminal.Delta(fmt.Sprintf("Copied snapshot [%s] to region [%s].", sourceSnapshot.SnapshotID, propRegion))
					}

				}(propRegion)
			}
		}

		wg.Wait()

		if errs != nil {
			return errors.New("Error propagating snapshot to other regions!")
		}
	}

	// Rotate out older snapshots
	if cfg.Rotate && cfg.Retain > 1 {
		terminal.Notice("Rotate flag is set, looking for snapshots to rotate...")
		err := rotateSnapshots(class, cfg, dryRun)
		if err != nil {
			terminal.ShowErrorMessage(fmt.Sprintf("Error rotating [%s] snapshots!", sourceSnapshot.Class), err.Error())
			return err
		}
	}

	terminal.Information("Done!")

	return nil
}

// rotateSnapshots rotates out older Snapshots
func rotateSnapshots(class string, cfg config.SnapshotClass, dryRun bool) error {
	var wg sync.WaitGroup
	var errs []error

	launchConfigs, err := GetLaunchConfigurations("")
	if err != nil {
		return errors.New("Error while retrieving the list of assets to exclude from rotation!")
	}
	excludedSnaps := launchConfigs.LockedSnapshotIds()

	regions := regions.GetRegionList()

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
				if excludedSnaps[snap.SnapshotID] {
					terminal.Notice("Snapshot [" + snap.Name + " (" + snap.SnapshotID + ") ] is being used in a launch configuration, skipping!")
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

// waitForSnapshot waits for a snapshot to complete
func waitForSnapshot(snapshotID, region string, dryRun bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Wait for the snapshot to complete.
	waitParams := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{aws.String(snapshotID)},
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

// private function without terminal prompts
func createSnapshot(volumeID, description, region string, dryRun bool) (*ec2.Snapshot, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Create the Snapshot
	snapshotParams := &ec2.CreateSnapshotInput{
		VolumeId:    aws.String(volumeID),
		DryRun:      aws.Bool(dryRun),
		Description: aws.String(description),
	}

	createSnapshotResp, err := svc.CreateSnapshot(snapshotParams)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return createSnapshotResp, errors.New(awsErr.Message())
		}
	}

	return createSnapshotResp, err
}

// DeleteSnapshots deletes one or more EBS Snapshots based on the given search term an optional region input.
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

// private function without the confirmation terminal prompts
func deleteSnapshots(snapList *Snapshots, dryRun bool) (err error) {
	for _, snapshot := range *snapList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(snapshot.Region)}))

		params := &ec2.DeleteSnapshotInput{
			SnapshotId: aws.String(snapshot.SnapshotID),
			DryRun:     aws.Bool(dryRun),
		}

		_, err := svc.DeleteSnapshot(params)
		if err != nil {
			return err
		}

		terminal.Delta("Deleted Snapshot [" + snapshot.Name + "] in [" + snapshot.Region + "]!")
	}

	return nil
}

// Len returns the number of EBS Snapshots
func (s Snapshots) Len() int {
	return len(s)
}

// Swap swaps the Snapshots at index i and j
func (s Snapshots) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns true of the Snapshot at index i is newer than the Snapshot at index j
func (s Snapshots) Less(i, j int) bool {
	return s[i].StartTime.After(s[j].StartTime)
}

// PrintTable Prints an ascii table of the list of EBS Snapshots
func (s *Snapshots) PrintTable() {
	if len(*s) == 0 {
		terminal.ShowErrorMessage("Warning", "No Snapshots Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*s))

	for index, snapshot := range *s {
		models.ExtractAwsmTable(index, snapshot, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
