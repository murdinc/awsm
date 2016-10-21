package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/murdinc/awsm/config"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

func getClasses(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	resp, err := config.LoadAllClasses(classType)

	if err == nil {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "classes": resp, "success": true})
	} else {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "classes": resp, "success": false, "errors": []string{err.Error()}})
	}
}

func getClassNames(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	resp, err := config.LoadAllClassNames(classType)

	if err == nil {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "classNames": resp, "success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"classType": classType, "classNames": resp, "success": true})
}

func getClassByName(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	className := chi.URLParam(r, "className")
	resp, err := config.LoadClassByName(classType, className)

	if err == nil {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "class": resp, "success": false, "errors": []string{err.Error()}})
		return
	}
	render.JSON(w, r, map[string]interface{}{"classType": classType, "className": className, "class": resp, "success": true})
}

func putClass(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	className := chi.URLParam(r, "className")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	// check for empty body?
	if len(body) == 0 {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{"No class object was passed!"}})
		return
	}

	switch classType {
	case "instances":
		var t config.InstanceClass

		err = json.Unmarshal(body, &t)
		if err != nil {
			render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
			return
		}

		err = config.InsertClasses(classType, config.InstanceClasses{className: t})
		if err != nil {
			render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
			return
		}
	}

	render.JSON(w, r, map[string]interface{}{"success": true})
}

func deleteClass(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	className := chi.URLParam(r, "className")

	err := config.DeleteClass(classType, className)
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}
	render.JSON(w, r, map[string]interface{}{"success": true})
}
