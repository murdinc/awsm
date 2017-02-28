package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/murdinc/terminal"
)

// Widgets is a map of Widgets
type Widgets map[string]Widget

type WidgetSlice []Widget

// Widget is a single widget class
type Widget struct {
	WidgetType string `json:"widgetType" awsmWidget:"Widget Type"`
	Title      string `json:"title" awsmWidget:"Title"`
	RssURL     string `json:"rssUrl" awsmWidget:"RSS URL"`
	Index      int    `json:"index"`
	Enabled    bool   `json:"enabled" awsmWidget:"Enabled"`
	Count      int    `json:"count" awsmWidget:"Count"`
	Name       string `json:"-"` // temporary only
}

func (f WidgetSlice) Len() int {
	return len(f)
}

func (f WidgetSlice) Less(i, j int) bool {
	return f[i].Index < f[j].Index
}

func (f WidgetSlice) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// DefaultWidgets returns the default Widget classes
func DefaultWidgets() Widgets {
	defaultWidgets := make(Widgets)

	defaultWidgets["events"] = Widget{
		WidgetType: "events",
		Title:      "AWS Events",
		Index:      0,
		Enabled:    true,
	}
	defaultWidgets["securitybulletins"] = Widget{
		WidgetType: "rss",
		Title:      "AWS Security Bulletins",
		RssURL:     "http://aws.amazon.com/security/security-bulletins/feed/",
		Count:      10,
		Index:      1,
		Enabled:    true,
	}
	defaultWidgets["awsblog"] = Widget{
		WidgetType: "rss",
		Title:      "AWS Blog",
		RssURL:     "http://feeds.feedburner.com/AmazonWebServicesBlog",
		Count:      10,
		Index:      2,
		Enabled:    true,
	}

	return defaultWidgets
}

// SaveWidget reads and unmarshals a byte slice and inserts it into the db
func SaveWidget(widgetName string, data []byte) (widget Widget, err error) {
	err = json.Unmarshal(data, &widget)
	if err != nil {
		return
	}

	err = Insert("widgets", Widgets{widgetName: widget})
	return
}

// DeleteWidget deletes a widget from SimpleDB
func DeleteWidget(widgetName string) error {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})) // TODO handle default region preference
	svc := simpledb.New(sess)

	itemName := "widgets/" + widgetName

	params := &simpledb.DeleteAttributesInput{
		DomainName: aws.String("awsm"),
		ItemName:   aws.String(itemName),
	}

	terminal.Delta("Deleting Widget item [" + itemName + "]...")
	_, err := svc.DeleteAttributes(params)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// LoadWidget returns a single Widget by its name
func LoadWidget(name string) (Widget, error) {
	cfgs := make(Widgets)
	item, err := GetItemByName("widgets", name)
	if err != nil {
		return cfgs[name], err
	}
	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllWidgets returns all Image classes
func LoadAllWidgets() (Widgets, error) {
	cfgs := make(Widgets)
	items, err := GetItemsByType("widgets")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts items from SimpleDB into Widgets
func (d Widgets) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "widgets/", "", -1)
		cfg := new(Widget)

		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "WidgetType":
				cfg.WidgetType = val

			case "Title":
				cfg.Title = val

			case "RssURL":
				cfg.RssURL = val

			case "Index":
				cfg.Index, _ = strconv.Atoi(val)

			case "Count":
				cfg.Count, _ = strconv.Atoi(val)

			case "Enabled":
				cfg.Enabled, _ = strconv.ParseBool(val)

			}
		}
		d[name] = *cfg
	}
}

// LoadAllWidgetNames loads all widget names
func LoadAllWidgetNames() ([]string, error) {
	// Check for the awsm db
	if !CheckDB() {
		return nil, nil
	}

	// Get the widgets
	items, err := GetItemsByType("widget")
	if err != nil {
		return nil, err
	}

	names := make([]string, len(items))
	for i, item := range items {
		names[i] = strings.Replace(*item.Name, "widget/", "", -1)
	}

	return names, nil
}
