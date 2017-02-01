package config

import (
	"sort"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/simpledb"
)

// FeedItems is a map of FeedItems
type FeedItems []FeedItem

// FeedItem is a single RSS feed item, see models.FeedItem for the model used for output and awsmTable parsing.
type FeedItem struct {
	ID       string    `json:"guid"`
	Link     string    `json:"link"`
	Date     time.Time `json:"date"`
	Title    string    `json:"title"`
	Summary  string    `json:"summary"`
	Category string    `json:"category"`
}

func (f FeedItems) Len() int {
	return len(f)
}

func (f FeedItems) Less(i, j int) bool {
	return f[i].Date.After(f[j].Date)
}

func (f FeedItems) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// SaveScalingPolicyClass reads and unmarshals a byte slice and inserts it into the db
func SaveFeed(feedName string, latest FeedItems, max int) (feed FeedItems, err error) {

	svc := simpledb.New(session.New(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference
	existing, _ := LoadAllFeedItems(feedName)

	sort.Sort(existing)
	sort.Sort(latest)

	if len(latest) < 1 {
		return existing, nil
	}

	// Delete the existing items
	DeleteItemsByType("feeditems/" + feedName)

	itemsMap := make(map[string][]*simpledb.ReplaceableAttribute)
	for index, feedItem := range latest {
		if index > max {
			break
		}

		itemsMap[feedItem.ID] = append(itemsMap[feedItem.ID], BuildAttributes(feedItem, "feeditems/"+feedName)...)
	}

	count := len(itemsMap)

Loop:
	for _, feedItem := range existing {
		itemsMap[feedItem.ID] = append(itemsMap[feedItem.ID], BuildAttributes(feedItem, "feeditems/"+feedName)...)
		count++
		if count < max {
			continue Loop
		}
	}

	items := make([]*simpledb.ReplaceableItem, len(itemsMap))
	for item, attributes := range itemsMap {

		i := &simpledb.ReplaceableItem{
			Attributes: attributes,
			Name:       aws.String(item),
		}
		items = append(items, i)

	}

	params := &simpledb.BatchPutAttributesInput{
		DomainName: aws.String("awsm"),
		Items:      items,
	}

	_, err = svc.BatchPutAttributes(params)
	if err != nil {
		return latest, err
	}

	itms, err := LoadAllFeedItems(feedName)

	sort.Sort(itms)
	return itms, err

}

func (i *FeedItems) ParseItem(item *rss.Item) {
	*i = append(*i, FeedItem{
		ID:       item.ID,
		Link:     item.Link,
		Date:     item.Date,
		Title:    item.Title,
		Category: item.Category,
		//Summary:  item.Summary,
	})
}

// LoadScalingPolicyClass loads a Scaling Policy Class by its name
func LoadAllFeedItems(feedName string) (FeedItems, error) {
	items, err := GetItemsByType("feeditems/" + feedName)
	fi := make(FeedItems, len(items))

	if err != nil {
		return fi, err
	}

	fi.Marshal(items)

	//println(fi[0].Date.String())
	return fi, nil
}

// Marshal puts items from SimpleDB into a Scaling Policy Class
func (f FeedItems) Marshal(items []*simpledb.Item) {
	for i, item := range items {
		cfg := new(FeedItem)
		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "ID":
				cfg.ID = val

			case "Link":
				cfg.Link = val

			case "Date":
				cfg.Date, _ = time.Parse("2006-01-02 15:04:05 +0000 UTC", val)

			case "Title":
				cfg.Title = val

			case "Summary":
				cfg.Summary = val

			}
		}
		f[i] = *cfg
	}
}
