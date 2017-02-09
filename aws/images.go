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
	"github.com/murdinc/awsm/aws/regions"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// Images represents a slice of Amazon Machine Images
type Images []Image

// Image represents a single Amazon Machine Image
type Image models.Image

// GetImageName returns the name of an AMI given the provided AMI ID
func (i *Images) GetImageName(id string) string {
	for _, img := range *i {
		if img.ImageID == id && img.Name != "" {
			return img.Name
		} else if img.ImageID == id {
			return id
		}
	}
	return id
}

// GetImagesByTag returns a slice of Amazon Machine Images given the provided region, and tag key/values
func GetImagesByTag(region, key, value string, available bool) (Images, error) {

	imgList := new(Images)

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

	img := make(Images, len(result.Images))
	for i, image := range result.Images {
		img[i].Marshal(image, region)
	}

	if available {
		for i, _ := range img {
			if img[i].State == "available" {
				*imgList = append(*imgList, img[i])
			}
		}
	} else {
		*imgList = append(*imgList, img[:]...)
	}

	return *imgList, err
}

// GetImageById returns an Amazon Machine Image via its ID
func GetImageById(region, id string) (Image, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	params := &ec2.DescribeImagesInput{
		ImageIds: []*string{
			aws.String(id),
		},
	}

	result, err := svc.DescribeImages(params)

	if len(result.Images) == 0 {
		return Image{}, errors.New("No Images found with id of [" + id + "] in [" + region + "].")
	}

	imgList := make(Images, len(result.Images))
	for i, image := range result.Images {
		imgList[i].Marshal(image, region)
	}

	return imgList[0], err

}

// GetLatestImageByTag returns the newest Amazon Machine Image in the provided region that matches the key/value tag provided
func GetLatestImageByTag(region, key, value string) (Image, error) {
	images, err := GetImagesByTag(region, key, value, true)
	if err != nil {
		return Image{}, err
	}

	sort.Sort(images)

	return images[0], err
}

// GetImages returns a slice of Images based on the provided search term and optional available flag
func GetImages(search string, available bool) (*Images, []error) {
	var wg sync.WaitGroup
	var errs []error

	imgList := new(Images)
	regions := regions.GetRegionList()

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

// GetRegionImages returns a slice of AMI's into the passed Image slice based on the provided region and search term, and optional available flag
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
		if available {
			for i, _ := range img {
				if img[i].State == "available" {
					*imgList = append(*imgList, img[i])
				}
			}
		} else {
			*imgList = append(*imgList, img[:]...)
		}
	}

	return nil
}

// CopyImage copies an existing AMI to another region
func CopyImage(search, region string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Validate the destination region
	if !regions.ValidRegion(region) {
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

	terminal.Delta("Created Image [" + *copyImageResp.ImageId + "] named [" + image.Name + "] to [" + region + "]!")

	// Add Tags
	return SetEc2NameAndClassTags(copyImageResp.ImageId, image.Name, image.Class, region)
}

// private function without prompts
func copyImage(image Image, region string, dryRun bool) (*ec2.CopyImageOutput, error) {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Copy image to the destination region
	params := &ec2.CopyImageInput{
		Name:          aws.String(image.Name),
		SourceImageId: aws.String(image.ImageID),
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

// CreateImage creates a new Amazon Machine Image from an instance matching the provided search term. It assigns the Image the class and name that was provided
func CreateImage(class, search string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	cfg, err := config.LoadImageClass(class)
	if err != nil {
		return err
	}

	terminal.Information("Found Image Class Configuration for [" + class + "]!")

	sourceInstance := cfg.Instance
	if search != "" {
		sourceInstance = search
	}
	if sourceInstance == "" {
		return errors.New("No instance specified in command arguments or Image class config. Please provide the instance search argument or set one in the config.")
	}

	// Locate the Instance
	instances, _ := GetInstances(sourceInstance, true)
	instCount := len(*instances)
	if instCount == 0 {
		return errors.New("No running instances found matching [" + sourceInstance + "], Aborting!")
	}
	if instCount > 1 {
		instances.PrintTable()
		return errors.New("Found more than one instances found matching: " + sourceInstance)
	}

	instance := (*instances)[0]
	region := instance.Region

	// Save the new instance id if we were searching for one
	if search != "" && !dryRun {
		instTable := &Instances{instance}
		instTable.PrintTable()

		// Confirm
		if !terminal.PromptBool("Are you sure you want to create an image from this instance and set it as the default for the " + class + " class?") {
			return errors.New("Aborting!")
		}

		cfg.SetInstance(class, instance.InstanceID)
	}

	// Increment the version
	terminal.Information(fmt.Sprintf("Previous version of image is [%d]", cfg.Version))
	if !dryRun {
		cfg.Increment(class)
	} else {
		cfg.Version++
	}
	terminal.Delta(fmt.Sprintf("New version of image is [%d]", cfg.Version))

	name := fmt.Sprintf("%s-v%d", class, cfg.Version)

	createImageResp, err := createImage(instance.InstanceID, name, region, dryRun)
	if err != nil {
		return err
	}

	terminal.Delta("Created Image [" + *createImageResp.ImageId + "] named [" + name + "] in [" + region + "]!")

	// Add Tags
	err = SetEc2NameAndClassTags(createImageResp.ImageId, name, class, region)

	if err != nil {
		return err
	}

	sourceImage := Image{Name: name, Class: class, ImageID: *createImageResp.ImageId, Region: region}

	// Check for Propagate flag
	if cfg.Propagate && cfg.PropagateRegions != nil {

		var wg sync.WaitGroup
		var errs []error

		terminal.Notice("Propagate flag is set, waiting for initial image to complete...")

		// Wait for the image to complete.
		err = waitForImage(*createImageResp.ImageId, region, dryRun)
		if err != nil {
			return err
		}

		// Copy to other regions
		for _, propRegion := range cfg.PropagateRegions {

			if propRegion != region {

				wg.Add(1)
				go func(propRegion string) {
					defer wg.Done()

					// Copy image to the destination region
					copyImageResp, err := copyImage(sourceImage, propRegion, dryRun)

					if err != nil {
						terminal.ShowErrorMessage(fmt.Sprintf("Error propagating image [%s] to region [%s]", sourceImage.ImageID, propRegion), err.Error())
						errs = append(errs, err)
					} else {
						// Add Tags
						err = SetEc2NameAndClassTags(copyImageResp.ImageId, name, class, propRegion)
						terminal.Delta(fmt.Sprintf("Copied image [%s] to region [%s].", sourceImage.ImageID, propRegion))
					}

				}(propRegion)
			}
		}

		wg.Wait()

		if errs != nil {
			return errors.New("Error propagating snapshot to other regions!")
		}
	}

	// Rotate out older images
	if cfg.Rotate && cfg.Retain > 1 {
		terminal.Notice("Rotate flag is set, looking for images to rotate...")
		err := rotateImages(class, cfg, dryRun)
		if err != nil {
			terminal.ShowErrorMessage(fmt.Sprintf("Error rotating [%s] images!", sourceImage.Class), err.Error())
			return err
		}
	}

	return nil
}

// rotateImages rotates out images based on the "retain" number set in the Image class
func rotateImages(class string, cfg config.ImageClass, dryRun bool) error {
	var wg sync.WaitGroup
	var errs []error

	launchConfigs, err := GetLaunchConfigurations("")
	if err != nil {
		return errors.New("Error while retrieving the list of assets to exclude from rotation!")
	}
	excludedImages := launchConfigs.LockedImageIds()

	regions := regions.GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()

			// Get the images of this class in this region
			images, err := GetImagesByTag(*region.RegionName, "Class", class, false)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering image list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}

			// Exclude the images being used in Launch Configurations
			for i, image := range images {
				if excludedImages[image.ImageID] {
					terminal.Information("Image [" + image.Name + " (" + image.ImageID + ") ] is being used in a launch configuration, skipping!")
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

// waitForImage waits for an Image to complete being created
func waitForImage(imageID, region string, dryRun bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Wait for the snapshot to complete.
	waitParams := &ec2.DescribeImagesInput{
		ImageIds: []*string{aws.String(imageID)},
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

// createImage is the private function without terminal prompts
func createImage(instanceID, name, region string, dryRun bool) (*ec2.CreateImageOutput, error) {

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Create the Image
	params := &ec2.CreateImageInput{
		InstanceId: aws.String(instanceID),
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
		DryRun:   aws.Bool(dryRun),
		NoReboot: aws.Bool(false), // force a reboot
	}
	createImageResp, err := svc.CreateImage(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return createImageResp, errors.New(awsErr.Message())
		}
	}

	return createImageResp, err
}

// Marshal parses the response from the aws sdk into an awsm Image
func (i *Image) Marshal(image *ec2.Image, region string) {
	var snapshotID, volSize string
	root := aws.StringValue(image.RootDeviceType)

	if root == "ebs" {
		for _, mapping := range image.BlockDeviceMappings {

			if *mapping.DeviceName == *image.RootDeviceName {
				snapshotID = aws.StringValue(mapping.Ebs.SnapshotId)
				volSize = fmt.Sprintf("%d GB", *mapping.Ebs.VolumeSize)
			}
		}
	}

	i.Name = GetTagValue("Name", image.Tags)
	i.Class = GetTagValue("Class", image.Tags)
	i.CreationDate, _ = time.Parse("2006-01-02T15:04:05.000Z", aws.StringValue(image.CreationDate))
	i.ImageID = aws.StringValue(image.ImageId)
	i.State = aws.StringValue(image.State)
	i.Root = root
	i.SnapshotID = snapshotID
	i.VolumeSize = volSize
	i.Region = region
	i.AmiName = aws.StringValue(image.Name)

	// Fall back to AMI Name
	if i.Name == "" {
		i.Name = i.AmiName
	}

}

// DeleteImages deletes one or more AMI images based on the search and optional region input
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
		return errors.New("No Images found!")
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
			ImageId: aws.String(image.ImageID),
			DryRun:  aws.Bool(dryRun),
		}

		_, err := svc.DeregisterImage(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				terminal.ErrorLine(awsErr.Message())
			} else {
				terminal.ErrorLine(err.Error())
			}
		} else {
			terminal.Delta("Deleted Image [" + image.Name + "] in [" + image.Region + "]!")
		}
	}

	return nil
}

// Len returns the number of Images within the Images slice
func (i Images) Len() int {
	return len(i)
}

// Swap swaps the position of two Images within the Image slice
func (i Images) Swap(k, j int) {
	i[k], i[j] = i[j], i[k]
}

// Less returns true if the Image at index k is newer than the Image at index j
func (i Images) Less(k, j int) bool {
	return i[k].CreationDate.After(i[j].CreationDate)
}

// PrintTable Prints an ascii table of the list of Amazon Machine Images
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
