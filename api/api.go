package api

import "github.com/kataras/iris"

func StartApi() {

	iris.Config.DisableBanner = true

	api := iris.Party("/api")

	// Assets
	assets := api.Party("/assets")
	assets.Get("/:assetType", getAssets)

	// Configs
	classes := api.Party("/classes")
	classes.Get("/:classType", getClasses)
	classes.Get("/:classType/names", getClassNames)
	classes.Get("/:classType/name/:className", getClassByName)
	//classes.Post("/:classType/name/:className", postConfig)
	//classes.Patch("/:classType/name/:className", patchConfig)
	//classes.Delete("/:classType/name/:className", deleteConfig)

	// Listen
	iris.Listen(":8080")
}
