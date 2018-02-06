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
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// Volumes represents a slice of EBS Volumes
type Volumes []Volume

// Volume represents a single EBS Volume
type Volume models.Volume

// GetVolumesByInstanceID returns a list of EBS Volumes given an instance Id
func GetVolumesByInstanceID(region, instanceId string) (Volumes, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.DescribeVolumesInput{

		Filters: []*ec2.Filter{
			{
				Name: aws.String("attachment.instance-id"),
				Values: []*string{
					aws.String(instanceId),
				},
			},
		},
	}

	result, err := svc.DescribeVolumes(params)
	if err != nil {
		return Volumes{}, err
	}

	instList := new(Instances)
	GetRegionInstances(region, instList, "", false)

	volList := make(Volumes, len(result.Volumes))
	for i, volume := range result.Volumes {
		volList[i].Marshal(volume, region, instList)
	}

	return volList, nil
}

// GetVolumeById returns a single EBS Volume given a region and volume id
func GetVolumeById(region, volumeId string) (Volume, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(volumeId)},
	}

	result, err := svc.DescribeVolumes(params)

	if err != nil {
		return Volume{}, err
	}

	count := len(result.Volumes)

	instList := new(Instances)
	GetRegionInstances(region, instList, "", false)

	switch count {
	case 0:
		return Volume{}, errors.New("No Volume found with id of [" + volumeId + "] in [" + region + "].")
	case 1:
		volume := new(Volume)
		volume.Marshal(result.Volumes[0], region, instList)
		return *volume, nil
	}

	volList := make(Volumes, len(result.Volumes))
	for i, volume := range result.Volumes {
		volList[i].Marshal(volume, region, instList)
	}

	sort.Sort(volList)
	volList.PrintTable()

	return Volume{}, errors.New("Please limit your search to return only one volume.")
}

// GetVolumeByTag returns a single EBS Volume given a region and Tag key/value
func GetVolumeByTag(region, key, value string) (Volume, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + key),
				Values: []*string{
					aws.String(value),
				},
			},
		},
	}

	result, err := svc.DescribeVolumes(params)

	if err != nil {
		return Volume{}, err
	}

	count := len(result.Volumes)

	instList := new(Instances)
	GetRegionInstances(region, instList, "", false)

	switch count {
	case 0:
		return Volume{}, errors.New("No Volume found with tag [" + key + "] of [" + value + "].")
	case 1:
		volume := new(Volume)
		volume.Marshal(result.Volumes[0], region, instList)
		return *volume, nil
	}

	volList := make(Volumes, len(result.Volumes))
	for i, volume := range result.Volumes {
		volList[i].Marshal(volume, region, instList)
	}

	sort.Sort(volList)
	volList.PrintTable()

	return Volume{}, errors.New("Please limit your search to return only one volume.")
}

// GetVolumeByInstanceIDandTag returns a list of EBS Volumes given an Instance Id and tag pair
func GetVolumeByInstanceIDSearch(region, instanceId, search string) (*Volumes, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	params := &ec2.DescribeVolumesInput{

		Filters: []*ec2.Filter{
			{
				Name: aws.String("attachment.instance-id"),
				Values: []*string{
					aws.String(instanceId),
				},
			},
		},
	}

	result, err := svc.DescribeVolumes(params)
	if err != nil {
		return &Volumes{}, err
	}

	volList := new(Volumes)

	instList := new(Instances)
	GetRegionInstances(region, instList, "", false)

	vol := make(Volumes, len(result.Volumes))
	for i, volume := range result.Volumes {
		vol[i].Marshal(volume, region, instList)
	}

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

	return volList, nil
}

// GetVolumes returns a slice of Volumes that match the provided search term and optional available flag
func GetVolumes(search string, available bool) (*Volumes, []error) {
	var wg sync.WaitGroup
	var errs []error

	volList := new(Volumes)
	regions := GetRegionListWithoutIgnored()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionVolumes(*region.RegionName, volList, search, available)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering volume list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return volList, errs
}

// GetRegionVolumes returns a slice of region Volumes into the provided Volumes slice that matches the provided region and search, and optional available flag
func GetRegionVolumes(region string, volList *Volumes, search string, available bool) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	result, err := svc.DescribeVolumes(&ec2.DescribeVolumesInput{})
	if err != nil {
		return err
	}

	instList := new(Instances)
	GetRegionInstances(region, instList, "", false)

	vol := make(Volumes, len(result.Volumes))
	for i, volume := range result.Volumes {
		vol[i].Marshal(volume, region, instList)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, v := range vol {
			rV := reflect.ValueOf(v)

			for k := 0; k < rV.NumField(); k++ {
				sVal := rV.Field(k).String()

				if term.MatchString(sVal) && ((available && vol[i].State == "available") || !available) {
					*volList = append(*volList, vol[i])
					continue Loop
				}
			}
		}
	} else {
		if available {
			for i, _ := range vol {
				if vol[i].State == "available" {
					*volList = append(*volList, vol[i])
				}
			}
		} else {
			*volList = append(*volList, vol[:]...)
		}
	}

	return nil
}

// RefreshVolume refreshes an EBS Volume on an Instance
func RefreshVolume(volumeSearch, instanceSearch string, force, dryRun bool) error {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the instance
	instances, _ := GetInstances(instanceSearch, true)
	instCount := len(*instances)
	if instCount == 0 {
		return errors.New("No instances found for search term.")
	} else if instCount > 1 {
		instances.PrintTable()
		return errors.New("Please limit your search terms to return only one instance.")
	}

	instance := (*instances)[0]
	region := instance.Region

	terminal.Information("Found Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + instance.Region + "]!")

	volList, err := GetVolumeByInstanceIDSearch(region, instance.InstanceID, volumeSearch)
	if err != nil {
		return err
	}

	volCount := len(*volList)
	if volCount == 0 {
		return errors.New("No volumes found in the same region as instance with your search term.")
	} else if volCount > 1 {
		volList.PrintTable()
		return errors.New("Please limit your search terms to return only one volume.")
	}

	volume := (*volList)[0]

	terminal.Information("Found Volume [" + volume.VolumeID + "] named [" + volume.Name + "] in [" + volume.Region + "]!")

	if volume.Class == "" {
		return errors.New("Volume [" + volume.VolumeID + "] does not have a Class associated with it. Aborting!")
	}

	volList.PrintTable()

	// Confirm
	if !terminal.PromptBool("Are you sure you want to refresh this Volume?") {
		return errors.New("Aborting!")
	}

	// Class Config
	volCfg, err := config.LoadVolumeClass(volume.Class)
	if err != nil {
		return err
	}

	// Check if we are able to send SSM commands to this instance, if needed
	runCommands := false
	var ssmInstance SSMInstance
	if volCfg.AttachCommand != "" || volCfg.DetachCommand != "" {
		terminal.Information("Volume Class [" + volume.Class + "] has SSM Commands configured, checking if we are able to send them...")

		ssmInstance, err = GetSSMInstanceById(instance.Region, instance.InstanceID)
		if err != nil || ssmInstance.InstanceID == "" {
			terminal.ErrorLine(err.Error() + " No attach/detach SSM Commands will be run on this instance!")

			// Confirm continue if we can't run them.
			if !terminal.PromptBool("Do you want to continue without running any attach/detach scripts?") {
				return errors.New("Aborting!")
			}
		} else {
			runCommands = true
			terminal.Information("Found SSM Instance [" + ssmInstance.InstanceID + "] named [" + ssmInstance.ComputerName + "] and a ping time of [" + humanize.Time(ssmInstance.LastPingDateTime) + "]!")
		}
	}

	// Refresh it
	err = refreshVolumes(volList, instance, ssmInstance, runCommands, force, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")
	return nil
}

// Private function without the confirmation terminal prompts
func refreshVolumes(volList *Volumes, instance Instance, ssmInstance SSMInstance, runCommands, force, dryRun bool) (err error) {
	for _, volume := range *volList {

		// Class Config
		volCfg, err := config.LoadVolumeClass(volume.Class)
		if err != nil {
			return err
		}

		// Get the latest snapshot
		latestSnapshot, err := GetLatestSnapshotByTag(volume.Region, "Class", volCfg.Snapshot)
		if err != nil {
			return err
		}

		// Notify according to snapshot age
		if latestSnapshot.SnapshotID == volume.SnapshoID {
			terminal.Information("Volume [" + volume.VolumeID + "] named [" + volume.Name + "] is already using the latest snapshot [" + latestSnapshot.SnapshotID + "] named [" + latestSnapshot.Name + "].")
		} else {
			terminal.Notice("Found newer snapshot [" + latestSnapshot.SnapshotID + "] with class of [" + latestSnapshot.Class + "] named [" + latestSnapshot.Name + "] created [" + humanize.Time(latestSnapshot.StartTime) + "]")
		}

		// Create the new Volume
		newVolume, err := createVolume(volume.Name, volume.Class, volume.AvailabilityZone, volCfg, latestSnapshot, dryRun)
		if err != nil {
			return err
		}

		// Detach the old Volume
		err = detachVolume(volume, volCfg, ssmInstance, runCommands, force, dryRun)
		if err != nil {
			return err
		}

		// Attach the new Volume
		err = attachVolume(newVolume, volCfg, instance, ssmInstance, runCommands, dryRun)
		if err != nil {
			return err
		}

		// Delete the old Volume
		err = deleteVolumes(volList, dryRun)
		if err != nil {
			return err
		}

		terminal.Delta("Refreshed Volume [" + volume.VolumeID + "] named [" + volume.Name + "] in [" + volume.Region + "]!")
	}

	return nil
}

// DetachVolume detaches an EBS Volume from an Instance
func DetachVolume(volumeSearch, instanceSearch string, force, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the instance
	instances, _ := GetInstances(instanceSearch, true)
	instCount := len(*instances)
	if instCount == 0 {
		return errors.New("No instances found for search term.")
	} else if instCount > 1 {
		instances.PrintTable()
		return errors.New("Please limit your search terms to return only one instance.")
	}

	instance := (*instances)[0]
	region := instance.Region

	terminal.Information("Found Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + region + "]!")

	volList, err := GetVolumeByInstanceIDSearch(region, instance.InstanceID, volumeSearch)
	if err != nil {
		return err
	}

	volCount := len(*volList)
	if volCount == 0 {
		return errors.New("No volumes found in the same region as instance with your search term.")
	} else if volCount > 1 {
		volList.PrintTable()
		return errors.New("Please limit your search terms to return only one volume.")
	}

	volume := (*volList)[0]
	var volCfg config.VolumeClass
	if volume.Class == "" {
		return errors.New("Volume [" + volume.VolumeID + "] does not have a Class associated with it.")
	} else {
		// Class Config
		volCfg, err = config.LoadVolumeClass(volume.Class)
		if err != nil {
			return err
		}
	}

	volList.PrintTable()

	// Confirm
	if !terminal.PromptBool("Are you sure you want to detatch this Volume?") {
		return errors.New("Aborting!")
	}

	// Check if we are able to send SSM commands to this instance, if needed
	runCmd := false
	var ssmInstance SSMInstance
	if volCfg.AttachCommand != "" || volCfg.DetachCommand != "" {
		terminal.Information("Volume Class [" + volume.Class + "] has SSM Commands configured, checking if we are able to send them...")

		ssmInstance, err = GetSSMInstanceById(instance.Region, instance.InstanceID)
		if err != nil || ssmInstance.InstanceID == "" {
			terminal.ErrorLine(err.Error() + " No attach/detach SSM Commands will be run on this instance!")

			// Confirm continue if we can't run them.
			if !terminal.PromptBool("Do you want to continue without running any attach/detach scripts?") {
				return errors.New("Aborting!")
			}
		} else {
			runCmd = true
			terminal.Information("Found SSM Instance [" + ssmInstance.InstanceID + "] named [" + ssmInstance.ComputerName + "] and a ping time of [" + humanize.Time(ssmInstance.LastPingDateTime) + "]!")
		}
	}

	// Detach it
	err = detachVolume(volume, volCfg, ssmInstance, runCmd, force, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")
	return nil

}

// Private function without the confirmation terminal prompts
func detachVolume(volume Volume, volCfg config.VolumeClass, ssmInstance SSMInstance, runCmd, force, dryRun bool) (err error) {

	// Run the Detach Command on the Instance
	if runCmd && volCfg.DetachCommand != "" {
		terminal.Delta("Running Detach Command...")
		invocations, err := runCommand(&SSMInstances{ssmInstance}, volCfg.DetachCommand, dryRun)
		if err != nil {
			return err
		}
		invocations.PrintOutput()
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(volume.Region)}))
	svc := ec2.New(sess)

	params := &ec2.DetachVolumeInput{
		VolumeId: aws.String(volume.VolumeID),
		DryRun:   aws.Bool(dryRun),
		Force:    aws.Bool(force),
		// InstanceId: aws.String(volume.InstanceID),
		// Device:   aws.String("String"),
	}

	terminal.Delta("Detaching Volume [" + volume.VolumeID + "] named [" + volume.Name + "] from instance [" + volume.InstanceID + "] in [" + volume.Region + "]...")

	_, err = svc.DetachVolume(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Notice("Waiting for EBS Volume to detach...")

	// Wait for it
	waitForVolume(volume.VolumeID, volume.Region, dryRun)

	terminal.Information("Detached Volume [" + volume.VolumeID + "] named [" + volume.Name + "] in [" + volume.Region + "]!")

	return nil
}

// AttachVolume attaches an EBS Volume to an Instance
func AttachVolume(volumeSearch, instanceSearch string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Get the instance
	instances, _ := GetInstances(instanceSearch, true)
	instCount := len(*instances)
	if instCount == 0 {
		return errors.New("No instances found for search term.")
	} else if instCount > 1 {
		instances.PrintTable()
		return errors.New("Please limit your search terms to return only one instance.")
	}

	instance := (*instances)[0]
	region := instance.Region

	terminal.Information("Found Instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + region + "]!")

	// Look for the volume in the same region as the instance
	volList := new(Volumes)
	err := GetRegionVolumes(region, volList, volumeSearch, true)
	if err != nil {
		return err
	}

	volCount := len(*volList)
	if volCount == 0 {
		return errors.New("No currently unattached volumes found in the same region as instance with your search term.")
	} else if volCount > 1 {
		volList.PrintTable()
		return errors.New("Please limit your search terms to return only one volume.")
	}

	volume := (*volList)[0]
	var volCfg config.VolumeClass
	if volume.Class == "" {
		return errors.New("Volume [" + volume.VolumeID + "] does not have a Class associated with it.")
	} else {
		// Class Config
		volCfg, err = config.LoadVolumeClass(volume.Class)
		if err != nil {
			return err
		}
	}

	volList.PrintTable()

	// Confirm
	if !terminal.PromptBool("Are you sure you want to attach this Volume?") {
		return errors.New("Aborting!")
	}

	// Check if we are able to send SSM commands to this instance, if needed
	runCmd := false
	var ssmInstance SSMInstance
	if volCfg.AttachCommand != "" {
		terminal.Information("Volume Class [" + volume.Class + "] has an SSM Attach Command configured, checking if we are able to send it...")

		ssmInstance, err = GetSSMInstanceById(instance.Region, instance.InstanceID)
		if err != nil || ssmInstance.InstanceID == "" {
			terminal.ErrorLine(err.Error() + " No SSM Attach Command will be run on this instance!")

			// Confirm continue if we can't run it.
			if !terminal.PromptBool("Do you want to continue without running any attach commands?") {
				return errors.New("Aborting!")
			}
		} else {
			runCmd = true
			terminal.Information("Found SSM Instance [" + ssmInstance.InstanceID + "] named [" + ssmInstance.ComputerName + "] and a ping time of [" + humanize.Time(ssmInstance.LastPingDateTime) + "]!")
		}
	}

	// Attach it
	err = attachVolume(volume, volCfg, instance, ssmInstance, runCmd, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")
	return nil
}

// Private function without the confirmation terminal prompts
func attachVolume(volume Volume, volCfg config.VolumeClass, instance Instance, ssmInstance SSMInstance, runCmd, dryRun bool) (err error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(volume.Region)}))
	svc := ec2.New(sess)

	params := &ec2.AttachVolumeInput{
		Device:     aws.String(volCfg.DeviceName),
		InstanceId: aws.String(instance.InstanceID),
		VolumeId:   aws.String(volume.VolumeID),
		DryRun:     aws.Bool(dryRun),
	}

	terminal.Delta("Attaching Volume [" + volume.VolumeID + "] named [" + volume.Name + "] to [" + instance.InstanceID + "] in [" + volume.Region + "]...")

	_, err = svc.AttachVolume(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Attached Volume [" + volume.VolumeID + "] named [" + volume.Name + "] to instance [" + instance.InstanceID + "] named [" + instance.Name + "] in [" + volume.Region + "]!")

	// Run the Attach Command on the Instance
	if runCmd && volCfg.AttachCommand != "" {
		terminal.Delta("Running Attach Command...")
		invocations, err := runCommand(&SSMInstances{ssmInstance}, volCfg.AttachCommand, dryRun)
		if err != nil {
			return err
		}
		invocations.PrintOutput()
	}

	return nil
}

// CreateVolume creates a new EBS Volume
func CreateVolume(class, name, az string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	volCfg, err := config.LoadVolumeClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found Volume Class Configuration for [" + class + "]!")

	// Verify the az input
	azs, errs := regions.GetAZs()
	if errs != nil {
		return errors.New("Error Verifying Availability Zone input")
	}
	if !azs.ValidAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	}

	terminal.Information("Found Availability Zone [" + az + "]!")

	region := azs.GetRegion(az)

	// Get the latest snapshot
	latestSnapshot, err := GetLatestSnapshotByTag(region, "Class", volCfg.Snapshot)
	if err != nil {
		return err
	}

	terminal.Information("Found Snapshot [" + latestSnapshot.SnapshotID + "] named [" + latestSnapshot.Name + "] with a class of [" + latestSnapshot.Class + "] created [" + humanize.Time(latestSnapshot.StartTime) + "]!")

	// Confirm
	if !terminal.PromptBool("Are you sure you want to create this Volume?") {
		return errors.New("Aborting!")
	}

	// Create it
	_, err = createVolume(name, class, az, volCfg, latestSnapshot, dryRun)
	if err != nil {
		return err
	}

	return nil

}

// Private function without the confirmation terminal prompts
func createVolume(name, class, az string, volCfg config.VolumeClass, latestSnapshot Snapshot, dryRun bool) (Volume, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(latestSnapshot.Region)}))
	svc := ec2.New(sess)

	params := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(az),
		DryRun:           aws.Bool(dryRun),
		Size:             aws.Int64(int64(volCfg.VolumeSize)),
		SnapshotId:       aws.String(latestSnapshot.SnapshotID),
		VolumeType:       aws.String(volCfg.VolumeType),
		//Encrypted:      aws.Bool(true),
		//Iops:           aws.Int64(1),
		//KmsKeyId:       aws.String("String"),
	}

	if volCfg.VolumeType == "io1" {
		params.SetIops(int64(volCfg.Iops))
	}

	createVolumeResp, err := svc.CreateVolume(params)
	newVolumeId := aws.StringValue(createVolumeResp.VolumeId)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return Volume{}, errors.New(awsErr.Message())
		}
		return Volume{}, err
	}

	terminal.Delta("Created Volume [" + newVolumeId + "] named [" + name + "] in [" + latestSnapshot.Region + "]!")

	terminal.Notice("Waiting to tag EBS Volume...")

	// Wait for it
	err = waitForVolume(newVolumeId, latestSnapshot.Region, dryRun)
	if err != nil {
		return Volume{}, err
	}

	terminal.Delta("Adding EBS Tags...")

	// Add Tags
	err = SetEc2NameAndClassTags(&newVolumeId, name, class, latestSnapshot.Region)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return Volume{}, errors.New(awsErr.Message())
		}
		return Volume{}, err
	}

	// Return the volume
	return GetVolumeById(latestSnapshot.Region, newVolumeId)
}

// waitForVolume waits for a Volume to complete being created
func waitForVolume(volumeID, region string, dryRun bool) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ec2.New(sess)

	// Wait for the snapshot to complete.
	waitParams := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(volumeID)},
		DryRun:    aws.Bool(dryRun),
	}

	err := svc.WaitUntilVolumeAvailable(waitParams)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
	}
	return err
}

// DeleteVolumes deletes one or more EBS Volumes given the search term and optional region input.
func DeleteVolumes(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	volList := new(Volumes)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionVolumes(region, volList, search, true)
	} else {
		volList, _ = GetVolumes(search, true)
	}

	if err != nil {
		return errors.New("Error gathering Volume list")
	}

	if len(*volList) > 0 {
		// Print the table
		volList.PrintTable()
	} else {
		return errors.New("No available Volumes found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Volumes?") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = deleteVolumes(volList, dryRun)
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
func deleteVolumes(volList *Volumes, dryRun bool) (err error) {
	for _, volume := range *volList {
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(volume.Region)}))
		svc := ec2.New(sess)

		params := &ec2.DeleteVolumeInput{
			VolumeId: aws.String(volume.VolumeID),
			DryRun:   aws.Bool(dryRun),
		}

		_, err := svc.DeleteVolume(params)
		if err != nil {
			return err
		}

		terminal.Delta("Deleted Volume [" + volume.Name + "] in [" + volume.Region + "]!")
	}

	return nil
}

// Marshal parses the response from the aws sdk into an awsm Volume
func (v *Volume) Marshal(volume *ec2.Volume, region string, instList *Instances) {
	v.Name = GetTagValue("Name", volume.Tags)
	v.Class = GetTagValue("Class", volume.Tags)
	v.VolumeID = aws.StringValue(volume.VolumeId)
	v.Size = int(aws.Int64Value(volume.Size))
	v.SizeHuman = fmt.Sprintf("%d GB", v.Size)
	v.State = aws.StringValue(volume.State)
	v.Encrypted = aws.BoolValue(volume.Encrypted)
	v.Iops = fmt.Sprint(aws.Int64Value(volume.Iops))
	v.CreationTime = *volume.CreateTime
	v.VolumeType = aws.StringValue(volume.VolumeType)
	v.SnapshoID = aws.StringValue(volume.SnapshotId)
	v.AvailabilityZone = aws.StringValue(volume.AvailabilityZone)
	v.Region = region

	if v.State == "in-use" {
		v.InstanceID = aws.StringValue(volume.Attachments[0].InstanceId)
		instance := instList.GetInstanceName(v.InstanceID)
		v.Attachment = instance
		v.DeleteOnTerm = aws.BoolValue(volume.Attachments[0].DeleteOnTermination)
		v.Device = aws.StringValue(volume.Attachments[0].Device)
	}
}

// Len returns the number of EBS Volumes
func (v Volumes) Len() int {
	return len(v)
}

// Swap swaps the Volumes at index i and j
func (v Volumes) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// Less returns true if the Volume at index i was created after the volume at index j
func (v Volumes) Less(i, j int) bool {
	return v[i].CreationTime.After(v[j].CreationTime)
}

// PrintTable Prints an ascii table of the list of EBS Volumes
func (v *Volumes) PrintTable() {
	if len(*v) == 0 {
		terminal.ShowErrorMessage("Warning", "No Volumes Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*v))

	for index, vol := range *v {
		models.ExtractAwsmTable(index, vol, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
