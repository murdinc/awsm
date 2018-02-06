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
	humanize "github.com/dustin/go-humanize"
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
func GetSnapshotsByTag(region, key, value string, completed bool) (Snapshots, error) {
	snapList := new(Snapshots)

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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

	snap := make(Snapshots, len(result.Snapshots))
	for i, snapshot := range result.Snapshots {
		snap[i].Marshal(snapshot, region)
	}

	if len(snap) == 0 {
		return *snapList, nil
	}

	if completed {
		for i, _ := range snap {
			if snap[i].State == "completed" {
				*snapList = append(*snapList, snap[i])
			}
		}
	} else {
		*snapList = append(*snapList, snap[:]...)
	}

	return *snapList, err
}

// GetLatestSnapshotByTag returns the newest EBS Snapshot that matches the provided region and Tag key/value
func GetLatestSnapshotByTag(region, key, value string) (Snapshot, error) {
	snapshots, err := GetSnapshotsByTag(region, key, value, true)
	if err != nil {
		return Snapshot{}, err
	}

	if len(snapshots) == 0 {
		return Snapshot{}, errors.New("No snapshots found in " + region + " with " + key + " of " + value + "!")
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
	regions := GetRegionListWithoutIgnored()

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
	s.StartTime = *snapshot.StartTime
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

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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
		if completed {
			for i, _ := range snap {
				if snap[i].State == "completed" {
					*snapList = append(*snapList, snap[i])
				}
			}
		} else {
			*snapList = append(*snapList, snap[:]...)
		}
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

	_, err := copySnapshot(snapshot, region, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Copied Snapshot [" + snapshot.SnapshotID + "] named [" + snapshot.Name + "] to [" + region + "]!")

	return nil
}

// private function without terminal prompts
func copySnapshot(snapshot Snapshot, region string, dryRun bool) (string, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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
	newSnapshotId := aws.StringValue(copySnapResp.SnapshotId)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return newSnapshotId, errors.New(awsErr.Message())
		}
	}

	// Add Tags
	err = SetEc2NameAndClassTags(&newSnapshotId, snapshot.Name, snapshot.Class, region)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return newSnapshotId, errors.New(awsErr.Message())
		}
		return newSnapshotId, err
	}

	terminal.Delta("Created [" + newSnapshotId + "] from [" + snapshot.SnapshotID + "] named [" + snapshot.Name + "] copied from region [" + snapshot.Region + "] to region [" + region + "]!")

	return newSnapshotId, err
}

// CreateSnapshot creates a new EBS Snapshot
func CreateSnapshot(class, search string, waitFlag, forceYes, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// --wait flag
	if waitFlag {
		terminal.Information("--wait flag is set!")
	}

	// --force-yes flag
	if forceYes {
		terminal.Information("--force-yes flag is set!")
	}

	// Class Config
	snapCfg, err := config.LoadSnapshotClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found Snapshot Class Configuration for [" + class + "]!")

	sourceVolume := snapCfg.Volume
	if search != "" {
		sourceVolume = search
		terminal.Information("Volume search argument passed, looking for Volume matching [" + sourceVolume + "]...")
	}
	if sourceVolume == "" {
		return errors.New("No volume specified in command arguments or Snapshot class config. Please provide the volume search argument or set one in the config.")
	}

	// Locate the Volume
	volumes, _ := GetVolumes(sourceVolume, false)
	if len(*volumes) == 0 {
		return errors.New("No volumes found matching [" + sourceVolume + "], Aborting!")
	}
	if len(*volumes) > 1 {
		volumes.PrintTable()
		return errors.New("Found more than one volume matching [" + sourceVolume + "], Aborting!")
	}

	volume := (*volumes)[0]
	region := volume.Region

	terminal.Information("Found Volume [" + volume.VolumeID + "] named [" + volume.Name + "] in region [" + region + "]!")

	// Ask to save the new volume id if it doesn't match the snapCfg.Volume
	if !dryRun && snapCfg.Volume != volume.VolumeID {
		volTable := &Volumes{volume}
		volTable.PrintTable()

		// Save into config prompt
		if forceYes || terminal.PromptBool("Do you want to set volume ["+volume.VolumeID+"] named ["+volume.Name+"] as the new default for the "+class+" snapshot class?") {
			snapCfg.SetVolume(class, volume.VolumeID)
		}
	}

	// Confirm
	if !forceYes && !terminal.PromptBool("Are you sure you want to create this Snapshot?") {
		return errors.New("Aborting!")
	}

	// Check if we are able to send SSM commands to this instance, if needed
	runCmds := false
	var ssmInstance SSMInstance

	if volume.InstanceID != "" {
		if snapCfg.PreSnapshotCommand != "" || snapCfg.PostSnapshotCommand != "" {
			terminal.Information("Snapshot Class [" + class + "] has SSM Commands configured, checking if we are able to send them...")

			ssmInstance, err = GetSSMInstanceById(volume.Region, volume.InstanceID)
			if err != nil || ssmInstance.InstanceID == "" {
				terminal.ErrorLine(err.Error() + " No SSM pre/post SnapshotCommands will be run on this instance!")

				// Confirm continue if we can't run them.
				if !terminal.PromptBool("Do you want to continue without running any pre/post Snapshot scripts?") {
					return errors.New("Aborting!")
				}
			} else {
				runCmds = true
				terminal.Information("Found SSM Instance [" + ssmInstance.InstanceID + "] named [" + ssmInstance.ComputerName + "] and a ping time of [" + humanize.Time(ssmInstance.LastPingDateTime) + "]!")
			}
		}
	}

	// Increment the version
	terminal.Information(fmt.Sprintf("Previous version of snapshot is [%d]", snapCfg.Version))
	if !dryRun {
		snapCfg.Increment(class)
	} else {
		snapCfg.Version++
	}
	terminal.Delta(fmt.Sprintf("New version of snapshot is [%d]", snapCfg.Version))

	name := fmt.Sprintf("%s-v%d", class, snapCfg.Version)

	// Create the snapshot
	newSnapshotId, err := createSnapshot(volume, snapCfg, ssmInstance, runCmds, dryRun)
	if err != nil {
		return err
	}

	// Add Tags
	err = SetEc2NameAndClassTags(&newSnapshotId, name, class, region)
	if err != nil {
		return err
	}

	terminal.Delta("Created Snapshot [" + newSnapshotId + "] named [" + name + "] in [" + region + "]!")

	sourceSnapshot := Snapshot{Name: name, Class: class, SnapshotID: newSnapshotId, Region: region, Description: snapCfg.Description}

	// Check for Propagate flag
	if snapCfg.Propagate && snapCfg.PropagateRegions != nil {

		var wg sync.WaitGroup
		var errs []error

		terminal.Notice("Propagate flag is set, waiting for initial snapshot to complete...")

		// Wait for the snapshot to complete.
		err = waitForSnapshot(newSnapshotId, region, dryRun)
		if err != nil {
			return err
		}

		// Copy to other regions
		for _, propRegion := range snapCfg.PropagateRegions {

			if propRegion != region {

				wg.Add(1)
				go func(propRegion string) {
					defer wg.Done()

					// Copy snapshot to the destination region
					newSnapshotId, err := copySnapshot(sourceSnapshot, propRegion, dryRun)

					if err != nil {
						terminal.ShowErrorMessage(fmt.Sprintf("Error propagating snapshot [%s] to region [%s]", sourceSnapshot.SnapshotID, propRegion), err.Error())
						errs = append(errs, err)
					} else {
						// Add Tags
						SetEc2NameAndClassTags(&newSnapshotId, name, class, propRegion)
						terminal.Information(fmt.Sprintf("Copied snapshot [%s] to region [%s].", sourceSnapshot.SnapshotID, propRegion))

					}

					if waitFlag {
						// Wait for the snapshot to complete.
						terminal.Notice(fmt.Sprintf("Waiting for snapshot [%s] to complete...", newSnapshotId))
						err = waitForSnapshot(newSnapshotId, propRegion, dryRun)
						if err != nil {
							errs = append(errs, err)
						}
						terminal.Delta(fmt.Sprintf("Snapshot [%s] in [%s] has completed!", newSnapshotId, propRegion))
					}

				}(propRegion)
			}
		}

		wg.Wait()

		if errs != nil {
			return errors.New("Error propagating snapshot to other regions!")
		}

	} else if waitFlag {
		// Wait for the snapshot to complete here otherwise, maybe
		terminal.Notice(fmt.Sprintf("Waiting for snapshot [%s] to complete...", newSnapshotId))
		err = waitForSnapshot(newSnapshotId, region, dryRun)
		if err != nil {
			return err
		}
		terminal.Delta(fmt.Sprintf("Snapshot [%s] completed!", newSnapshotId))
	}

	// Rotate out older snapshots
	if snapCfg.Rotate && snapCfg.Retain > 1 {
		terminal.Notice("Rotate flag is set, looking for snapshots to rotate...")
		err := rotateSnapshots(class, snapCfg, dryRun)
		if err != nil {
			terminal.ShowErrorMessage(fmt.Sprintf("Error rotating [%s] snapshots!", sourceSnapshot.Class), err.Error())
			return err
		}
	}

	terminal.Information("Done!")

	return nil
}

// private function without terminal prompts
func createSnapshot(volume Volume, snapCfg config.SnapshotClass, ssmInstance SSMInstance, runCmds, dryRun bool) (string, error) {

	// Run the Pre-Snapshot Command on the Instance
	if runCmds && snapCfg.PreSnapshotCommand != "" {
		terminal.Delta("Running Pre-Snapshot Command...")
		invocations, err := runCommand(&SSMInstances{ssmInstance}, snapCfg.PreSnapshotCommand, dryRun)
		if err != nil {
			return "", err
		}
		invocations.PrintOutput()
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(volume.Region)}))
	svc := ec2.New(sess)

	// Create the Snapshot
	snapshotParams := &ec2.CreateSnapshotInput{
		VolumeId:    aws.String(volume.VolumeID),
		DryRun:      aws.Bool(dryRun),
		Description: aws.String(snapCfg.Description),
	}

	createSnapshotResp, err := svc.CreateSnapshot(snapshotParams)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return "", errors.New(awsErr.Message())
		}
	}

	// Run the Post-Snapshot Command on the Instance
	if runCmds && snapCfg.PostSnapshotCommand != "" {
		terminal.Delta("Running Post-Snapshot Command...")
		invocations, err := runCommand(&SSMInstances{ssmInstance}, snapCfg.PostSnapshotCommand, dryRun)
		if err != nil {
			return "", err
		}
		invocations.PrintOutput()
	}

	return aws.StringValue(createSnapshotResp.SnapshotId), err
}

// rotateSnapshots rotates out older Snapshots
func rotateSnapshots(class string, cfg config.SnapshotClass, dryRun bool) error {
	// Bail early
	if cfg.Retain <= 0 {
		return nil
	}

	var wg sync.WaitGroup
	var errs []error

	launchConfigs, err := GetLaunchConfigurations("")
	if err != nil {
		return errors.New("Error while retrieving the list of assets to exclude from rotation!")
	}
	lockedSnapshots := launchConfigs.LockedSnapshotIds()

	regions := GetRegionListWithoutIgnored()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()

			// Get all the snapshots of this class in this region
			snapshots, err := GetSnapshotsByTag(*region.RegionName, "Class", class, true)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering snapshot list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}

			var unlockedSnapshots Snapshots

			// Exclude the snapshots being used in Launch Configurations
			for _, snap := range snapshots {
				if lockedSnapshots[snap.SnapshotID] {
					terminal.Notice("Snapshot [" + snap.SnapshotID + "] in [" + *region.RegionName + "] named [" + snap.Name + "] is being used in a launch configuration, skipping!")
				} else {
					unlockedSnapshots = append(unlockedSnapshots, snap)
				}
			}

			// Delete the oldest ones if we have more than the retention number
			if len(unlockedSnapshots) > cfg.Retain {
				sort.Sort(unlockedSnapshots) // important!
				ds := unlockedSnapshots[cfg.Retain:]
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

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

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
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(snapshot.Region)}))
		svc := ec2.New(sess)

		params := &ec2.DeleteSnapshotInput{
			SnapshotId: aws.String(snapshot.SnapshotID),
			DryRun:     aws.Bool(dryRun),
		}

		_, err := svc.DeleteSnapshot(params)
		if err != nil {
			return err
		}

		terminal.Delta("Deleted Snapshot [" + snapshot.SnapshotID + "] named [" + snapshot.Name + "] in [" + snapshot.Region + "]!")
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
