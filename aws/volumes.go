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
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// Volumes represents a slice of EBS Volumes
type Volumes []Volume

// Volume represents a single EBS Volume
type Volume models.Volume

// GetVolumeByTag returns a single EBS Volume given a region and Tag key/value
func GetVolumeByTag(region, key, value string) (Volume, error) {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

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

// GetVolumes returns a slice of Volumes that match the provided search term and optional available flag
func GetVolumes(search string, available bool) (*Volumes, []error) {
	var wg sync.WaitGroup
	var errs []error

	volList := new(Volumes)
	regions := regions.GetRegionList()

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
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
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
		*volList = append(*volList, vol[:]...)
	}

	return nil
}

// DetachVolume detaches an EBS Volume from an Instance
func DetachVolume(volume, instance string, force, dryRun bool) error {
	// Get the instance
	instances, _ := GetInstances(instance, true)

	instCount := len(*instances)
	if instCount == 0 {
		return errors.New("No instances found for search term.")
	} else if instCount > 1 {
		return errors.New("Please limit your search terms to return only one instance.")
	}

	inst := (*instances)[0]
	region := inst.Region

	// Look for the volume in the same region as the instance
	volList := new(Volumes)
	err := GetRegionVolumes(region, volList, volume, true)
	if err != nil {
		return err
	}

	volCount := len(*volList)
	if volCount == 0 {
		return errors.New("No volumes found in the same region as instance with your search term.")
	} else if volCount > 1 {
		return errors.New("Please limit your search terms to return only one volume.")
	}

	vol := (*volList)[0]

	// Detach it
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DetachVolumeInput{
		VolumeId:   aws.String(vol.VolumeID),
		InstanceId: aws.String(inst.InstanceID),
		DryRun:     aws.Bool(dryRun),
		Force:      aws.Bool(force),
		//Device:   aws.String("String"),
	}

	_, err = svc.DetachVolume(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil
}

// AttachVolume attaches an EBS Volume to an Instance
func AttachVolume(volume, instance string, dryRun bool) error {

	// Get the instance
	instances, _ := GetInstances(instance, true)

	instCount := len(*instances)
	if instCount == 0 {
		return errors.New("No instances found for search term.")
	} else if instCount > 1 {
		return errors.New("Please limit your search terms to return only one instance.")
	}

	inst := (*instances)[0]
	region := inst.Region

	// Look for the volume in the same region as the instance
	volList := new(Volumes)
	err := GetRegionVolumes(region, volList, volume, true)
	if err != nil {
		return err
	}

	volCount := len(*volList)
	if volCount == 0 {
		return errors.New("No volumes found in the same region as instance with your search term.")
	} else if volCount > 1 {
		return errors.New("Please limit your search terms to return only one volume.")
	}

	vol := (*volList)[0]

	// Class Config
	volCfg, err := config.LoadVolumeClass(vol.Class)
	if err != nil {
		return err
	}

	terminal.Information("Found Volume Class Configuration for [" + vol.Class + "]!")

	// Attach it
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.AttachVolumeInput{
		Device:     aws.String(volCfg.DeviceName),
		InstanceId: aws.String(inst.InstanceID),
		VolumeId:   aws.String(vol.VolumeID),
		DryRun:     aws.Bool(dryRun),
	}

	_, err = svc.AttachVolume(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
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

	terminal.Information("Found Snapshot [" + latestSnapshot.SnapshotID + "] with class [" + latestSnapshot.Class + "] created [" + latestSnapshot.CreatedHuman + "]!")

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

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
	createVolumeResp, err := svc.CreateVolume(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Created Volume [" + *createVolumeResp.VolumeId + "] named [" + name + "] in [" + region + "]!")

	// Add Tags
	err = SetEc2NameAndClassTags(createVolumeResp.VolumeId, name, class, region)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil

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
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(volume.Region)}))

		params := &ec2.DeleteVolumeInput{
			VolumeId: aws.String(volume.VolumeID),
			DryRun:   aws.Bool(dryRun),
		}

		_, err := svc.DeleteVolume(params)
		if err != nil {
			return err
		}

		terminal.Information("Deleted Volume [" + volume.Name + "] in [" + volume.Region + "]!")
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
	v.Iops = fmt.Sprint(aws.Int64Value(volume.Iops))
	v.CreationTime = *volume.CreateTime                              // robots
	v.CreatedHuman = humanize.Time(aws.TimeValue(volume.CreateTime)) // humans
	v.VolumeType = aws.StringValue(volume.VolumeType)
	v.SnapshoID = aws.StringValue(volume.SnapshotId)
	v.AvailabilityZone = aws.StringValue(volume.AvailabilityZone)
	v.Region = region

	if v.State == "in-use" {
		v.InstanceID = aws.StringValue(volume.Attachments[0].InstanceId)
		instance := instList.GetInstanceName(v.InstanceID)
		v.Attachment = instance
		v.DeleteOnTerm = aws.BoolValue(volume.Attachments[0].DeleteOnTermination)

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
