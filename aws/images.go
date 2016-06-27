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
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type Images []Image

type Image struct {
	Name         string
	Class        string
	CreationDate string
	ImageId      string
	State        string
	Root         string
	SnapshotId   string
	VolumeSize   string
	Region       string
}

func GetLatestImageByTag(region, key, value string) (Image, error) {

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

	if err != nil {
		return Image{}, err
	}

	count := len(result.Images)

	switch count {
	case 0:
		return Image{}, errors.New("No Image found with [" + key + "] of [" + value + "] in [" + region + "], Aborting!")
	case 1:
		image := new(Image)
		image.Marshall(result.Images[0], region)
		return *image, nil
	}

	imgList := make(Images, len(result.Images))
	for i, image := range result.Images {
		imgList[i].Marshall(image, region)
	}

	sort.Sort(imgList)

	return imgList[0], nil
}

func GetImages(search string) (*Images, []error) {
	var wg sync.WaitGroup
	var errs []error

	imgList := new(Images)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionImages(*region.RegionName, imgList, search, false)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering image list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return imgList, errs
}

func CreateImage(search, class, name string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Class Config
	var imgCfg config.ImageClassConfig
	err := imgCfg.LoadConfig(class)
	if err != nil {
		return err
	} else {
		terminal.Information("Found Image Class Configuration for [" + class + "]!")
	}

	// Locate the Instance
	instances, _ := GetInstances(search, true)
	if len(*instances) == 0 {
		return errors.New("No running instances found for your search terms.")
	}
	if len(*instances) > 1 {
		instances.PrintTable()
		return errors.New("Please limit your search to return only one instance.")
	}

	instance := (*instances)[0]
	region := instance.Region

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))

	// Create the Image

	params := &ec2.CreateImageInput{
		InstanceId: aws.String(instance.InstanceId),
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
			return errors.New(awsErr.Message())
		}
		return err
	}

	terminal.Information("Created Image [" + *createImageResp.ImageId + "] named [" + name + "] in [" + region + "]!")

	// Add Tags
	snapshotTagsParams := &ec2.CreateTagsInput{
		Resources: []*string{
			createImageResp.ImageId,
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
	_, err = svc.CreateTags(snapshotTagsParams)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			return errors.New(awsErr.Message())
		}
		return err
	}

	return nil

}

func (i *Image) Marshall(image *ec2.Image, region string) {
	var snapshotId, volSize string
	root := aws.StringValue(image.RootDeviceType)
	//fmt.Println(root)

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
	i.CreationDate = aws.StringValue(image.CreationDate)
	i.ImageId = aws.StringValue(image.ImageId)
	i.State = aws.StringValue(image.State)
	i.Root = root
	i.SnapshotId = snapshotId
	i.VolumeSize = volSize
	i.Region = region
}

func GetRegionImages(region string, imgList *Images, search string, searchClass bool) error {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeImages(&ec2.DescribeImagesInput{Owners: []*string{aws.String("self")}})

	if err != nil {
		return err
	}

	img := make(Images, len(result.Images))
	for i, image := range result.Images {
		img[i].Marshall(image, region)
	}

	if search != "" {
		if searchClass { // Specific class search
			for i, in := range img {
				if in.Class == search {
					*imgList = append(*imgList, img[i])
				}
			}
		} else { // General search
			term := regexp.MustCompile(search)
		Loop:
			for i, in := range img {
				rInst := reflect.ValueOf(in)

				for k := 0; k < rInst.NumField(); k++ {
					sVal := rInst.Field(k).String()

					if term.MatchString(sVal) {
						*imgList = append(*imgList, img[i])
						continue Loop
					}
				}
			}
		}
	} else {
		*imgList = append(*imgList, img[:]...)
	}

	return nil
}

// Functions for sorting
func (i *Image) Timestamp() time.Time {
	timestamp, err := time.Parse("2006-01-02T15:04:05.000Z", i.CreationDate)
	if err != nil {
		fmt.Println(err)
		terminal.ErrorLine("Error parsing the timestamp for image [" + i.ImageId + "]!")
	}

	return timestamp
}

func (s Images) Len() int {
	return len(s)
}

func (s Images) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Images) Less(i, j int) bool {
	return s[i].Timestamp().After(s[j].Timestamp())
}

func (i *Images) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.Class,
			val.CreationDate,
			val.ImageId,
			val.State,
			val.Root,
			val.SnapshotId,
			val.VolumeSize,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Class", "Creation Date", "Image Id", "State", "Root", "Snapshot Id", "Volume Size", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
