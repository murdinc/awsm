package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/cli"
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

func GetImages() (*Images, error) {
	var wg sync.WaitGroup

	imgList := new(Images)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionImages(region.RegionName, imgList)
			if err != nil {
				cli.ShowErrorMessage("Error gathering image list", err.Error())
			}
		}(region)
	}
	wg.Wait()

	return imgList, nil
}

func GetRegionImages(region *string, imgList *Images) error {
	svc := ec2.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeImages(&ec2.DescribeImagesInput{Owners: []*string{aws.String("self")}})

	if err != nil {
		return err
	}

	img := make(Images, len(result.Images))
	for i, image := range result.Images {
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

		img[i] = Image{
			Name:         GetTagValue("Name", image.Tags),
			Class:        GetTagValue("Class", image.Tags),
			CreationDate: aws.StringValue(image.CreationDate),
			ImageId:      aws.StringValue(image.ImageId),
			State:        aws.StringValue(image.State),
			Root:         root,
			SnapshotId:   snapshotId,
			VolumeSize:   volSize,
			Region:       fmt.Sprintf(*region),
		}
	}
	*imgList = append(*imgList, img[:]...)

	return nil
}

func (i *Images) PrintTable() {
	collumns := []string{"Name", "Class", "Creation Date", "Image Id", "State", "Root", "Snapshot Id", "Volume Size", "Region"}

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

	printTable(collumns, rows)
}
