package api

import (
	"net/http"

	"github.com/murdinc/awsm/config"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

func getClasses(w http.ResponseWriter, r *http.Request) {
	// Get the classType
	classType := chi.URLParam(r, "classType")

	resp, err := config.LoadAllClasses(classType)

	if err == nil {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "classes": resp, "success": true})
	} else {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "classes": resp, "success": false, "errors": []string{err.Error()}})
	}
}

func getClassNames(w http.ResponseWriter, r *http.Request) {
	// Get the classType
	classType := chi.URLParam(r, "classType")

	resp, err := config.LoadAllClassNames(classType)

	if err == nil {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "classNames": resp, "success": true})
	} else {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "classNames": resp, "success": false, "errors": []string{err.Error()}})
	}

}

func getClassByName(w http.ResponseWriter, r *http.Request) {
	// Get the classType & className
	classType := chi.URLParam(r, "classType")
	className := chi.URLParam(r, "className")

	resp, err := config.LoadClassByName(classType, className)

	if err == nil {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "className": className, "class": resp, "success": true})
	} else {
		render.JSON(w, r, map[string]interface{}{"classType": classType, "class": resp, "success": false, "errors": []string{err.Error()}})
	}
}
