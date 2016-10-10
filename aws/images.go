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
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type Images []Image

type Image models.Image

func (i *Images) GetImageName(id string) string {
	for _, img := range *i {
		if img.ImageId == id && img.Name != "" {
			return img.Name
		} else if img.ImageId == id {
			return id
		}
	}
	return id
}

func GetImagesByTag(region, key, value string) (Images, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DescribeImagesInput{
		Owners: []*string{aws.String("self")},
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + key),
				Values: []*string{
					aws.String(value),
				},
			},
		},
	}

	result, err := svc.DescribeImages(params)

	imgList := make(Images, len(result.Images))
	for i, image := range result.Images {
		imgList[i].Marshal(image, region)
	}

	if len(imgList) == 0 {
		return imgList, errors.New("No Images found with tag [" + key + "] of [" + value + "] in [" + region + "].")
	}

	return imgList, err

}

func GetLatestImageByTag(region, key, value string) (Image, error) {
	images, err := GetImagesByTag(region, key, value)
	sort.Sort(images)

	return images[0], err
}

func GetImages(search string, available bool) (*Images, []error) {
	var wg sync.WaitGroup
	var errs []error

	imgList := new(Images)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionImages(*region.RegionName, imgList, search, available)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering image list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return imgList, errs
}

func GetRegionImages(region string, imgList *Images, search string, available bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeImages(&ec2.DescribeImagesInput{Owners: []*string{aws.String("self")}})

	if err != nil {
		return err
	}

	img := make(Images, len(result.Images))
	for i, image := range result.Images {
		img[i].Marshal(image, region)
	}

	if search != "" {
		for i, in := range img {
			if in.Class == search {
				*imgList = append(*imgList, img[i])
			}
		}

		term := regexp.MustCompile(search)
	Loop:
		for i, in := range img {
			rInst := reflect.ValueOf(in)

			for k := 0; k < rInst.NumField(); k++ {
				sVal := rInst.Field(k).String()

				if term.MatchString(sVal) && ((available && img[i].State == "available") || !available) {
					*imgList = append(*imgList, img[i])
					continue Loop
				}
			}
		}

	} else {
		*imgList = append(*imgList, img[:]...)
	}

	return nil
}

func CopyImage(search, region string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Validate the destination region
	if !ValidRegion(region) {
		return errors.New("Region [" + region + "] is Invalid!")
	}

	// Get the source image
	images, _ := GetImages(search, true)
	imgCount := len(*images)
	if imgCount == 0 {
		return errors.New("No available images found for your search terms.")
	}
	if imgCount > 1 {
		images.PrintTable()
		return errors.New("Please limit your search to return only one image.")
	}
	image := (*images)[0]

	// Copy image to the destination region
	copyImageResp, err := copyImage(image, region, dryRun)

	if err != nil {
		return err
	}

	terminal.Information("Created Image [" + *copyImageResp.ImageId + "] named [" + image.Name + "] to [" + region + "]!")

	// Add Tags
	return SetEc2NameAndClassTags(copyImageResp.ImageId, image.Name, image.Class, region)
}

func copyImage(image Image, region string, dryRun bool) (*ec2.CopyImageOutput, error) {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Copy image to the destination region
	params := &ec2.CopyImageInput{
		Name:          aws.String(image.Name),
		SourceImageId: aws.String(image.ImageId),
		SourceRegion:  aws.String(image.Region),
		DryRun:        aws.Bool(dryRun),
		//ClientToken: aws.String("String"),
		//Description: aws.String("String"),
		//Encrypted:   aws.Bool(true),
		//KmsKeyId:    aws.String("String"),
	}
	copyImageResp, err := svc.CopyImage(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return copyImageResp, errors.New(awsErr.Message())
		}
	}

	return copyImageResp, err
}

func CreateImage(search, class, name string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	cfg, err := config.LoadImageClass(class)

	if err != nil {
		return err
	} else {
		terminal.Information("Found Image Class Configuration for [" + class + "]!")
	}

	// Locate the Instance
	instances, _ := GetInstances(search, true)
	instCount := len(*instances)
	if instCount == 0 {
		return errors.New("No running instances found for your search terms.")
	}
	if instCount > 1 {
		instances.PrintTable()
		return errors.New("Please limit your search to return only one instance.")
	}

	instance := (*instances)[0]
	region := instance.Region

	createImageResp, err := createImage(instance.InstanceId, name, region, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Created Image [" + *createImageResp.ImageId + "] named [" + name + "] in [" + region + "]!")

	// Add Tags
	err = SetEc2NameAndClassTags(createImageResp.ImageId, name, class, region)

	if err != nil {
		return err
	}

	sourceImage := Image{Name: name, Class: class, ImageId: *createImageResp.ImageId, Region: region}

	// Check for Propagate flag
	if cfg.Propagate && cfg.PropagateRegions != nil {

		var wg sync.WaitGroup
		var errs []error

		terminal.Information("Propagate flag is set, waiting for initial snapshot to complete...")

		// Wait for the image to complete.
		err = waitForImage(*createImageResp.ImageId, region, dryRun)
		if err != nil {
			return err
		}

		// Copy to other regions
		for _, propRegion := range cfg.PropagateRegions {
			wg.Add(1)

			go func(region string) {
				defer wg.Done()

				// Copy image to the destination region
				copyImageResp, err := copyImage(sourceImage, propRegion, dryRun)

				if err != nil {
					terminal.ShowErrorMessage(fmt.Sprintf("Error propagating image [%s] to region [%s]", sourceImage.ImageId, propRegion), err.Error())
					errs = append(errs, err)
				}

				// Add Tags
				err = SetEc2NameAndClassTags(copyImageResp.ImageId, name, class, propRegion)

				terminal.Information(fmt.Sprintf("Copied image [%s] to region [%s].", sourceImage.ImageId, propRegion))

			}(propRegion)
		}

		wg.Wait()

		if errs != nil {
			return errors.New("Error propagating snapshot to other regions!")
		}
	}

	// Rotate out older images
	if cfg.Retain > 1 {
		err := RotateImages(class, cfg, dryRun)
		if err != nil {
			terminal.ShowErrorMessage(fmt.Sprintf("Error rotating [%s] images!", sourceImage.Class), err.Error())
			return err
		}
	}

	return nil
}

func RotateImages(class string, cfg config.ImageClass, dryRun bool) error {
	var wg sync.WaitGroup
	var errs []error

	launchConfigs, err := GetLaunchConfigurations("")
	if err != nil {
		return errors.New("Error while retrieving the list of assets to exclude from rotation!")
	}
	excludedImages := launchConfigs.LockedImageIds()

	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()

			// Get the images of this class in this region
			images, err := GetImagesByTag(*region.RegionName, "Class", class)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering image list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}

			// Exclude the images being used in Launch Configurations
			for i, image := range images {
				if excludedImages[image.ImageId] {
					terminal.Information("Image [" + image.Name + " (" + image.ImageId + ") ] is being used in a launch configuration, skipping!")
					images = append(images[:i], images[i+1:]...)
				}
			}

			// Delete the oldest ones if we have more than the retention number
			if len(images) > cfg.Retain {
				sort.Sort(images) // important!
				di := images[cfg.Retain:]
				deleteImages(&di, dryRun)
			}

		}(region)
	}
	wg.Wait()

	if errs != nil {
		return errors.New("Error rotating images for [" + class + "]!")
	}

	return nil
}

func waitForImage(imageId, region string, dryRun bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Wait for the snapshot to complete.
	waitParams := &ec2.DescribeImagesInput{
		ImageIds: []*string{aws.String(imageId)},
		Owners:   []*string{aws.String("self")},
		DryRun:   aws.Bool(dryRun),
	}

	err := svc.WaitUntilImageAvailable(waitParams)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
	}
	return err
}

func createImage(instanceId, name, region string, dryRun bool) (*ec2.CreateImageOutput, error) {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Create the Image
	params := &ec2.CreateImageInput{
		InstanceId: aws.String(instanceId),
		Name:       aws.String(name),
		/*
			BlockDeviceMappings: []*ec2.BlockDeviceMapping{
				{ // Required
					DeviceName: aws.String("String"),
					Ebs: &ec2.EbsBlockDevice{
						DeleteOnTermination: aws.Bool(true),
						Encrypted:           aws.Bool(true),
						Iops:                aws.Int64(1),
						SnapshotId:          aws.String("String"),
						VolumeSize:          aws.Int64(1),
						VolumeType:          aws.String("VolumeType"),
					},
					NoDevice:    aws.String("String"),
					VirtualName: aws.String("String"),
				},
			},
		*/
		//Description: aws.String("String"),
		DryRun: aws.Bool(dryRun),
		//NoReboot:    aws.Bool(true),
	}
	createImageResp, err := svc.CreateImage(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return createImageResp, errors.New(awsErr.Message())
		}
	}

	return createImageResp, err
}

func (i *Image) Marshal(image *ec2.Image, region string) {
	var snapshotId, volSize string
	root := aws.StringValue(image.RootDeviceType)

	if root == "ebs" {
		for _, mapping := range image.BlockDeviceMappings {

			if *mapping.DeviceName == *image.RootDeviceName {
				snapshotId = aws.StringValue(mapping.Ebs.SnapshotId)
				volSize = fmt.Sprintf("%d GB", *mapping.Ebs.VolumeSize)
			}
		}
	}

	i.Name = GetTagValue("Name", image.Tags)
	i.Class = GetTagValue("Class", image.Tags)
	i.CreationDate, _ = time.Parse("2006-01-02T15:04:05.000Z", aws.StringValue(image.CreationDate)) // robots
	i.CreatedHuman = humanize.Time(i.CreationDate)                                                  // humans
	i.ImageId = aws.StringValue(image.ImageId)
	i.State = aws.StringValue(image.State)
	i.Root = root
	i.SnapshotId = snapshotId
	i.VolumeSize = volSize
	i.Region = region
}

// Public function with confirmation terminal prompt
func DeleteImages(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	imgList := new(Images)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionImages(region, imgList, search, false)
	} else {
		imgList, _ = GetImages(search, false)
	}

	if err != nil {
		return errors.New("Error gathering Image list")
	}

	if len(*imgList) > 0 {
		// Print the table
		imgList.PrintTable()
	} else {
		return errors.New("No available Images found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Volumes?") {
		return errors.New("Aborting!")
	}

	// Delete 'Em
	err = deleteImages(imgList, dryRun)
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
func deleteImages(imgList *Images, dryRun bool) (err error) {
	for _, image := range *imgList {
		svc := ec2.New(session.New(&aws.Config{Region: aws.String(image.Region)}))

		params := &ec2.DeregisterImageInput{
			ImageId: aws.String(image.ImageId),
			DryRun:  aws.Bool(dryRun),
		}

		_, err := svc.DeregisterImage(params)
		if err != nil {
			return err
		}

		terminal.Information("Deleted Image [" + image.Name + "] in [" + image.Region + "]!")
	}

	return nil
}

// Functions for sorting
func (s Images) Len() int {
	return len(s)
}

func (s Images) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Images) Less(i, j int) bool {
	return s[i].CreationDate.After(s[j].CreationDate)
}

func (i *Images) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Images Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, image := range *i {
		models.ExtractAwsmTable(index, image, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
