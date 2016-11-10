package aws

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/murdinc/awsm/models"
)

// ARN represents a single Amazon Resource Number
type Events []models.Event

func GetEvents() (Events, error) {
	var events Events

	resp, err := http.Get("http://status.aws.amazon.com/data.json")
	if err != nil {
		return events, err
	}
	defer resp.Body.Close()

	var eventMap map[string][]interface{}
	err = json.NewDecoder(resp.Body).Decode(&eventMap)
	if err != nil {
		return events, err
	}

	for _, e := range eventMap["current"] {
		events.parseEvent(e, "current")
	}
	for _, e := range eventMap["archive"] {
		events.parseEvent(e, "archive")
	}

	sort.Sort(events)
	return events, nil
}

func (e *Events) parseEvent(ev interface{}, eventType string) {
	event := ev.(map[string]interface{})

	archiveStatus := false
	if eventType == "archive" {
		archiveStatus = true
	}

	i, err := strconv.ParseInt(event["date"].(string), 10, 64)
	if err != nil {
		panic(err)
	}
	date := time.Unix(i, 0)

	*e = append(*e, models.Event{
		Date:        date,
		Details:     event["details"].(string),
		Service:     event["service"].(string),
		ServiceName: event["service_name"].(string),
		Summary:     strings.TrimSpace(event["summary"].(string)),
		Archive:     archiveStatus,
	})
}

func (e Events) Len() int {
	return len(e)
}

func (e Events) Less(i, j int) bool {
	return e[i].Date.After(e[j].Date)
}

func (e Events) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
