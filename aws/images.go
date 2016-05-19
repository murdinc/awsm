package aws

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/terminal"
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

func GetLatestImage(class, region string) (Image, error) {
	imgList := new(Images)

	err := GetRegionImages(aws.String(region), imgList, class, true)
	if err != nil {
		return Image{}, err
	}

	sort.Sort(imgList)

	return (*imgList)[0], nil
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
			err := GetRegionImages(region.RegionName, imgList, search, false)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering image list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return imgList, errs
}

func GetRegionImages(region *string, imgList *Images, search string, searchClass bool) error {
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
