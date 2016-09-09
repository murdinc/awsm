package api

import (
	"github.com/kataras/iris"
	"github.com/murdinc/awsm/config"
)

func getClasses(ctx *iris.Context) {

	// Get the classType
	classType := ctx.Param("classType")

	resp, err := config.LoadAllClasses(classType)

	if err == nil {
		ctx.JSON(iris.StatusOK, map[string]interface{}{"classType": classType, "classes": resp, "success": true})
	} else {
		ctx.JSON(iris.StatusForbidden, map[string]interface{}{"classType": classType, "classes": resp, "success": false, "errors": []string{err.Error()}})
	}

}

func getClassNames(ctx *iris.Context) {

	// Get the classType
	classType := ctx.Param("classType")

	resp, err := config.LoadAllClassNames(classType)

	if err == nil {
		ctx.JSON(iris.StatusOK, map[string]interface{}{"classType": classType, "classNames": resp, "success": true})
	} else {
		ctx.JSON(iris.StatusForbidden, map[string]interface{}{"classType": classType, "classNames": resp, "success": false, "errors": []string{err.Error()}})
	}

}

func getClassByName(ctx *iris.Context) {

	// Get the classType & className
	classType := ctx.Param("classType")
	className := ctx.Param("className")

	resp, err := config.LoadClassByName(classType, className)

	if err == nil {
		ctx.JSON(iris.StatusOK, map[string]interface{}{"classType": classType, "class": resp, "success": true})
	} else {
		ctx.JSON(iris.StatusForbidden, map[string]interface{}{"classType": classType, "class": resp, "success": false, "errors": []string{err.Error()}})
	}
}
