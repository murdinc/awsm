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
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/dustin/go-humanize"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
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
	CreationTime     time.Time
	CreatedHuman     string
	VolumeType       string
	SnapshoId        string
	DeleteOnTerm     string
	AvailabilityZone string
	Region           string
}

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

	switch count {
	case 0:
		return Volume{}, errors.New("No Volume found with tag [" + key + "] of [" + value + "].")
	case 1:
		volume := new(Volume)
		volume.Marshall(result.Volumes[0], region)
		return *volume, nil
	}

	volList := make(Volumes, len(result.Volumes))
	for i, volume := range result.Volumes {
		volList[i].Marshall(volume, region)
	}

	sort.Sort(volList)

	volList.PrintTable()

	return Volume{}, errors.New("Please limit your search to return only one volume.")
}

func GetVolumes(search string, available bool) (*Volumes, []error) {
	var wg sync.WaitGroup
	var errs []error

	volList := new(Volumes)
	regions := GetRegionList()

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
		VolumeId:   aws.String(vol.VolumeId),
		InstanceId: aws.String(inst.InstanceId),
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
	var volCfg config.VolumeClassConfig
	err = volCfg.LoadConfig(vol.Class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Volume Class Configuration for [" + vol.Class + "]!")
	}

	// Attach it
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.AttachVolumeInput{
		Device:     aws.String(volCfg.DeviceName),
		InstanceId: aws.String(inst.InstanceId),
		VolumeId:   aws.String(vol.VolumeId),
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

func CreateVolume(class, name, az string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	var volCfg config.VolumeClassConfig
	err := volCfg.LoadConfig(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Volume Class Configuration for [" + class + "]!")
	}

	// Verify the az input
	azs, errs := GetAZs()
	if errs != nil {
		return errors.New("Error Verifying Availability Zone input")
	}
	if !azs.ValidAZ(az) {
		return cli.NewExitError("Availability Zone ["+az+"] is Invalid!", 1)
	} else {
		terminal.Information("Found Availability Zone [" + az + "]!")
	}

	region := azs.GetRegion(az)

	// Get the latest snapshot
	latestSnapshot, err := GetLatestSnapshotByTag(region, "Class", volCfg.Snapshot)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Snapshot [" + latestSnapshot.SnapshotId + "] with class [" + latestSnapshot.Class + "] created [" + latestSnapshot.CreatedHuman + "]!")
	}

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(az),
		DryRun:           aws.Bool(dryRun),
		Size:             aws.Int64(int64(volCfg.VolumeSize)),
		SnapshotId:       aws.String(latestSnapshot.SnapshotId),
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
	volumeTagsParams := &ec2.CreateTagsInput{
		Resources: []*string{
			createVolumeResp.VolumeId,
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
			{
				Key:   aws.String("Class"),
				Value: aws.String(class),
			},
		},
		DryRun: aws.Bool(dryRun),
	}
	_, err = svc.CreateTags(volumeTagsParams)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil

}

func (v *Volume) Marshall(volume *ec2.Volume, region string) {
	v.Name = GetTagValue("Name", volume.Tags)
	v.Class = GetTagValue("Class", volume.Tags)
	v.VolumeId = aws.StringValue(volume.VolumeId)
	v.Size = fmt.Sprint(aws.Int64Value(volume.Size))
	v.State = aws.StringValue(volume.State)
	v.Iops = fmt.Sprint(aws.Int64Value(volume.Iops))
	//v.Attachments =  aws.StringValue(volume.Attachments) // TODO
	v.CreationTime = *volume.CreateTime
	v.CreatedHuman = humanize.Time(aws.TimeValue(volume.CreateTime)) // humans
	v.VolumeType = aws.StringValue(volume.VolumeType)
	v.SnapshoId = aws.StringValue(volume.SnapshotId)
	//v.DeleteOnTerm =     aws.StringValue(volume.) // TODO
	v.AvailabilityZone = aws.StringValue(volume.AvailabilityZone)
	v.Region = region
}

func GetRegionVolumes(region string, volList *Volumes, search string, available bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeVolumes(&ec2.DescribeVolumesInput{})

	if err != nil {
		return err
	}

	vol := make(Volumes, len(result.Volumes))
	for i, volume := range result.Volumes {
		vol[i].Marshall(volume, region)
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

func (v Volumes) Len() int {
	return len(v)
}

func (v Volumes) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Volumes) Less(i, j int) bool {
	return v[i].CreationTime.After(v[j].CreationTime)
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
			val.Class,
			val.VolumeId,
			val.Size,
			val.State,
			val.Attachments,
			val.CreatedHuman,
			val.VolumeType,
			val.SnapshoId,
			val.DeleteOnTerm,
			val.AvailabilityZone,
		}
	}

	table.SetHeader([]string{"Name", "Class", "Volume Id", "Size", "State", "Attachment", "Created", "Volume Type", "Snapshot Id", "Delete on Termination", "Availability Zone"})

	table.AppendBulk(rows)
	table.Render()
}
