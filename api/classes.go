package api

import (
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/murdinc/awsm/config"
)

func exportClasses(w http.ResponseWriter, r *http.Request) {

	resp, err := config.Export()

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"classes": resp, "success": true})
}

func getClasses(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	resp, err := config.LoadAllClasses(classType)

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"classType": classType, "classes": resp, "success": true})
}

func getClassOptions(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	resp, err := config.LoadAllClassOptions(classType)

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"classType": classType, "classOptions": resp, "success": true})
}

func getClassNames(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	resp, err := config.LoadAllClassNames(classType)

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"classType": classType, "classNames": resp, "success": true})
}

func getClassByName(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	className := chi.URLParam(r, "className")
	resp, err := config.LoadClassByName(classType, className)

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"classType": classType, "className": className, "class": resp, "success": true})
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

func putClass(w http.ResponseWriter, r *http.Request) {
	classType := chi.URLParam(r, "classType")
	className := chi.URLParam(r, "className")

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{"Error Reading Body!", err.Error()}})
		return
	}

	// check for empty body?
	if len(data) == 0 {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{"No class object was passed!"}})
		return
	}

	var class interface{}

	switch classType {

	case "vpcs":
		class, err = config.SaveVpcClass(className, data)

	case "subnets":
		class, err = config.SaveSubnetClass(className, data)

	case "instances":
		class, err = config.SaveInstanceClass(className, data)

	case "volumes":
		class, err = config.SaveVolumeClass(className, data)

	case "snapshots":
		class, err = config.SaveSnapshotClass(className, data)

	case "images":
		class, err = config.SaveImageClass(className, data)

	case "autoscalegroups":
		class, err = config.SaveAutoscalingGroupClass(className, data)

	case "launchconfigurations":
		class, err = config.SaveLaunchConfigurationClass(className, data)

	case "loadbalancers":
		class, err = config.SaveLoadBalancerClass(className, data)

	case "scalingpolicies":
		class, err = config.SaveScalingPolicyClass(className, data)

	case "alarms":
		class, err = config.SaveAlarmClass(className, data)

	case "securitygroups":
		class, err = config.SaveSecurityGroupClass(className, data)

	case "keypairs":
		class, err = config.SaveKeyPairClass(className, data)

	}

	if err != nil {
		render.JSON(w, r, map[string]interface{}{"success": false, "errors": []string{"Error saving Class!", err.Error()}})
		return
	}

	render.JSON(w, r, map[string]interface{}{"classType": classType, "class": class, "success": true})
}
