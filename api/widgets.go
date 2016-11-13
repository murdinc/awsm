package api

import (
	"io/ioutil"
	"net/http"

	"github.com/SlyMarbo/rss"
	"github.com/murdinc/awsm/aws"
	"github.com/murdinc/awsm/config"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

func getWidgets(w http.ResponseWriter, r *http.Request) {
	resp, err := config.LoadAllWidgets()

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"widgets": resp, "success": true})
}

func getWidgetNames(w http.ResponseWriter, r *http.Request) {
	resp, err := config.LoadAllWidgetNames()

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"widgetNames": resp, "success": true})
}

func getWidgetByName(w http.ResponseWriter, r *http.Request) {
	widgetName := chi.URLParam(r, "widgetName")
	resp, err := config.LoadWidget(widgetName)

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"widgetName": widgetName, "widget": resp, "success": true})
}

func getWidgetOptions(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string][]string)
	resp["availableWidgets"] = []string{"events", "alarms", "rss"}
	render.JSON(w, r, map[string]interface{}{"options": resp, "success": true})
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	events, err := aws.GetEvents()
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"events": events, "success": true})
}

func getFeed(w http.ResponseWriter, r *http.Request) {
	feedName := chi.URLParam(r, "feedName")

	feedSettings, err := config.LoadWidget(feedName)
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	var items config.FeedItems

	feed, err := rss.Fetch(feedSettings.RssURL)
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	for _, i := range feed.Items {
		items.ParseItem(i)
	}

	items, err = config.SaveFeed(feedName, items, feedSettings.Count)
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"feed": items, "success": true})
}

func deleteWidget(w http.ResponseWriter, r *http.Request) {
	widgetName := chi.URLParam(r, "widgetName")

	err := config.DeleteWidget(widgetName)
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"success": true})
}

func putWidget(w http.ResponseWriter, r *http.Request) {
	widgetName := chi.URLParam(r, "widgetName")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{"Error Reading Body!", err.Error()}})
		return
	}

	// check for empty body?
	if len(data) == 0 {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{"No widget object was passed!"}})
		return
	}

	widget, err := config.SaveWidget(widgetName, data)

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{"Error saving Class!", err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"widget": widget, "success": true})
}
