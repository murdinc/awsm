package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// Widgets is a map of Widgets
type Widgets map[string]Widget

// Widget is a single widget class
type Widget struct {
	Index   int  `json:"index" awsmWidget:"Index"`
	Enabled bool `json:"enabled" awsmWidget:"Enabled"`
}

// DefaultWidgets returns the default Widget classes
func DefaultWidgets() Widgets {
	defaultWidgets := make(Widgets)

	defaultWidgets["events"] = Widget{
		Index:   0,
		Enabled: true,
	}
	defaultWidgets["securitybulletins"] = Widget{
		Index:   1,
		Enabled: true,
	}
	defaultWidgets["alarms"] = Widget{
		Index:   2,
		Enabled: false,
	}

	return defaultWidgets
}

// SaveWidget reads and unmarshals a byte slice and inserts it into the db
func SaveWidget(widgetName string, data []byte) (widget Widget, err error) {
	err = json.Unmarshal(data, &widget)
	if err != nil {
		return
	}

	err = InsertClasses("widgets", Widgets{widgetName: widget})
	return
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

			case "Index":
				cfg.Index, _ = strconv.Atoi(val)

			case "Enabled":
				cfg.Enabled, _ = strconv.ParseBool(val)

			}
		}
		d[name] = *cfg
	}
}
