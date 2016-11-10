package aws

import (
	"github.com/SlyMarbo/rss"
	"github.com/murdinc/awsm/config"
)

func GetSecurityBulletins() (config.FeedItems, error) {
	var items config.FeedItems

	feedUrl := "http://aws.amazon.com/security/security-bulletins/feed/"
	securityBulletins, err := rss.Fetch(feedUrl)

	if err != nil {
		return items, err
	}

	for _, i := range securityBulletins.Items {
		items.ParseItem(i)
	}

	items, err = config.SaveFeed("securitybulletins", items, 20)

	//sort.Sort(items)
	return items, err
}
