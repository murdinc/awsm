package api

import (
	"fmt"
	"net/http"

	"github.com/murdinc/awsm/aws"
	"github.com/murdinc/awsm/config"
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

func putWidgets(w http.ResponseWriter, r *http.Request) {
	resp, err := config.LoadAllWidgets()

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"dashboard": resp, "success": true})
}

func getWidgetOptions(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string][]string)
	resp["availableWidgets"] = []string{"events", "alarms"}
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

func getSecurityBulletins(w http.ResponseWriter, r *http.Request) {
	items, err := aws.GetSecurityBulletins()
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"securityBulletins": items, "success": true})
}

func TestGetStatusWidget() {
	status, err := aws.GetEvents()
	if err != nil {
		println(err.Error())
		return
	}

	for _, event := range status {
		fmt.Printf("%v\n", event)
	}
}
