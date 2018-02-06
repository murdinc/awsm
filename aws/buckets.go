package aws

import (
	"math/rand"
	"os"
	"reflect"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// Buckets represents a slice of AWS S3 Buckets
type Buckets []Bucket

// Bucket represents a single S3 Bucket
type Bucket models.Bucket

// GetBuckets returns a list of S3 Buckets that match the provided search term
func GetBuckets(search string) (*Buckets, error) {

	bucketList := new(Buckets)
	regions := GetRegionListWithoutIgnored()

	rand.Seed(time.Now().UnixNano())
	region := regions[rand.Intn(len(regions))] // pick a random region

	err := GetRegionBuckets(*region.RegionName, bucketList, search)

	return bucketList, err
}

// GetRegionBuckets returns a list of Buckets for a given region into the provided Buckets slice
func GetRegionBuckets(region string, bucketList *Buckets, search string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := s3.New(sess)

	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return err
	}

	bucket := make(Buckets, len(result.Buckets))
	for i, b := range result.Buckets {
		bucket[i].Marshal(b)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, kp := range bucket {
			rBucket := reflect.ValueOf(kp)

			for k := 0; k < rBucket.NumField(); k++ {
				sVal := rBucket.Field(k).String()

				if term.MatchString(sVal) {
					*bucketList = append(*bucketList, bucket[i])
					continue Loop
				}
			}
		}
	} else {
		*bucketList = append(*bucketList, bucket[:]...)
	}

	return nil
}

// Marshal parses the response from the aws sdk into an awsm Bucket
func (b *Bucket) Marshal(bucket *s3.Bucket) {
	b.Name = aws.StringValue(bucket.Name)
	b.CreationDate = aws.TimeValue(bucket.CreationDate)
}

// PrintTable Prints an ascii table of the list of KeyPairs
func (b *Buckets) PrintTable() {
	if len(*b) == 0 {
		terminal.ShowErrorMessage("Warning", "No Buckets Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*b))

	for index, key := range *b {
		models.ExtractAwsmTable(index, key, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
