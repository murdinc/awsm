package aws

import (
	"os/user"

	"github.com/kataras/iris"
	"github.com/murdinc/awsm/config"
	"github.com/toqueteos/webbrowser"
)

type Content struct {
	Title        string
	Type         string
	URLKey       string
	Data         map[string]interface{}
	Configs      interface{}
	AZList       []string
	RenderLayout bool
	Errors       []string
	Host         string
	//ClassFormURL string
}

func RunDashboard(devMode bool) {
	currentUser, _ := user.Current()
	guiLocation := currentUser.HomeDir + "/.awsm/gui/awsm-default-gui" // TODO accept custom theme directories

	// logger middleware
	//log := logger.New()

	//iris := iris.New()

	iris.Config.DisableBanner = true

	// Template Configuration
	iris.Config.Render.Template.Directory = guiLocation
	iris.Config.Render.Template.Layout = "templates/layout.html"

	// Static Asset Folders
	iris.StaticWeb("/js", guiLocation, 0)
	iris.StaticWeb("/css", guiLocation, 0)
	iris.StaticWeb("/fonts", guiLocation, 0)
	iris.StaticWeb("/static", guiLocation, 0)

	// Form Handler
	iris.Post("/create/:type", postData)

	// Index
	iris.Get("/", index)

	// Pages
	pages := iris.Party("/dashboard")
	pages.Get("/", getDashboard)
	pages.Get("/:page", getDashboard)

	// Modals
	modals := iris.Party("/modal")
	modals.Get("/:action/:type", getModal)
	modals.Get("/:action/:type/:class", getModal)

	if !devMode {
		webbrowser.Open("http://localhost:8080/dashboard") // TODO race condition?
	}

	iris.Listen(":8080") // TODO optionally configurable port #
}

// index redirect to dashboard
func index(ctx *iris.Context) {
	ctx.Redirect("/dashboard")
}

// ===================================
// Builds all the different full pages
func getDashboard(ctx *iris.Context) {

	page := ctx.Param("page")

	switch page {
	case "loadbalancers":
		loadbalancersPage(ctx)
	case "autoscalegroups":
		autoscalegroupsPage(ctx)
	case "launchconfigurations":
		launchconfigurationsPage(ctx)
	case "scalingpolicies":
		scalingpoliciesPage(ctx)
	case "alarms":
		alarmsPage(ctx)
	case "vpcs":
		vpcsPage(ctx)
	case "subnets":
		subnetsPage(ctx)
	case "routetables":
		routetablesPage(ctx)
	case "internetgateways":
		internetgatewaysPage(ctx)
	case "dhcpoptionsets":
		dhcpoptionssetsPage(ctx)
	case "elasticips":
		elasticipsPage(ctx)
	case "instances":
		instancesPage(ctx)
	case "images":
		imagesPage(ctx)
	case "volumes":
		volumesPage(ctx)
	case "snapshots":
		snapshotsPage(ctx)
	case "securitygroups":
		securitygroupsPage(ctx)
	default:
		ctx.Render("templates/dashboard.html", Content{Title: "Dashboard", Type: "Dashboard", URLKey: "dashboard", RenderLayout: true, Host: ctx.HostString()})
	}
}

// ===================================
// Recieves all the different forms
func postData(ctx *iris.Context) {

	var err error

	typ := ctx.Param("type")

	switch typ {
	case "instance":
		err = LaunchInstance(ctx.PostFormValue("class"), ctx.PostFormValue("sequence"), ctx.PostFormValue("az"), true)

	case "vpc":

	default:
		ctx.Render("templates/404-modal.html", Content{Title: "404", Type: "404", URLKey: "error"})
	}

	if err != nil {
		ctx.JSON(200, map[string]interface{}{"success": false, "error": err.Error()})
	} else {
		ctx.JSON(200, map[string]interface{}{"success": true})
	}

}

// ===================================
// Builds all the different modals
func getModal(ctx *iris.Context) {

	action := ctx.Param("action")
	typ := ctx.Param("type")

	switch action {
	case "new":
		switch typ {
		case "instance":
			newInstanceModal(ctx)
		case "vpc":
			newVpcModal(ctx)
		default:
			ctx.Render("templates/404-modal.html", Content{Title: "404", Type: "404", URLKey: "error"})
		}

	case "manage":
		switch typ {
		case "instance":
			manageInstanceClassesModal(ctx)
		case "vpc":
			manageVpcClassesModal(ctx)
		default:
			ctx.Render("templates/404-modal.html", Content{Title: "404", Type: "404", URLKey: "error"})
		}

	case "edit":
		switch typ {
		case "instance":
			instanceClassForm(ctx)
		case "vpc":
			vpcClassForm(ctx)
		default:
			ctx.Render("templates/404-modal.html", Content{Title: "404", Type: "404", URLKey: "error"})
		}

	default:
		ctx.Render("templates/404-modal.html", Content{Title: "404", Type: "404", URLKey: "error"})
	}
}

// ===================================
// EC2 Instances

func instancesPage(ctx *iris.Context) {
	data := make(map[string]interface{})
	instances, errs := GetInstances("")

	data["Instances"] = instances
	data["Errors"] = errs

	ctx.Render("templates/instances.html", Content{Title: "Instances", Type: "Instance", URLKey: "instance", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

func newInstanceModal(ctx *iris.Context) {

	data := make(map[string]interface{})
	azList := AZList()
	configs, err := config.GetAllConfigNames("ec2")
	if err != nil {
		data["Errors"] = []error{err}
	}

	data["Configs"] = configs
	data["AZList"] = azList

	ctx.Render("templates/new-instance-modal.html", Content{Title: "New Instance", Type: "Instance", URLKey: "instance", Data: data})
}

func manageInstanceClassesModal(ctx *iris.Context) {

	data := make(map[string]interface{})
	configs, err := config.GetAllConfigNames("ec2")
	if err != nil {
		data["Errors"] = []error{err}
	}

	data["Configs"] = configs

	ctx.Render("templates/manage-classes-modal.html", Content{Title: "Mange Instance Classes", Type: "Instance", URLKey: "instance", Data: data})
}

func instanceClassForm(ctx *iris.Context) {

	data := make(map[string]interface{})
	class := ctx.Param("class")

	var cfg config.InstanceClassConfig
	err := cfg.LoadConfig(class)
	if err != nil {
		data["Errors"] = []error{err}
	}

	data["ClassName"] = class
	data["ClassConfig"] = cfg

	ctx.Render("templates/instance-class-form.html", Content{Title: "Edit Instance Class", Type: "Instance", URLKey: "instance", Data: data})
}

// ===================================
// AMI

func imagesPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	images, errs := GetImages("")

	data["Images"] = images
	data["Errors"] = errs

	ctx.Render("templates/images.html", Content{Title: "Images", Type: "Image", URLKey: "image", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

// ===================================
// EBS

func volumesPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	volumes, errs := GetVolumes("")

	data["Volumes"] = volumes
	data["Errors"] = errs

	ctx.Render("templates/volumes.html", Content{Title: "Volumes", Type: "Volume", URLKey: "volume", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

func snapshotsPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	snapshots, errs := GetSnapshots("")

	data["Snapshots"] = snapshots
	data["Errors"] = errs

	ctx.Render("templates/snapshots.html", Content{Title: "Snapshots", Type: "Snapshot", URLKey: "snapshot", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

// ===================================
// Security Groups

func securitygroupsPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	securitygroups, errs := GetSecurityGroups()

	data["SecurityGroups"] = securitygroups
	data["Errors"] = errs

	ctx.Render("templates/securitygroups.html", Content{Title: "Security Groups", Type: "Security Group", URLKey: "security-group", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

// ===================================
// Load Balancers

func loadbalancersPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	loadbalancers, errs := GetLoadBalancers()

	data["LoadBalancers"] = loadbalancers
	data["Errors"] = errs

	ctx.Render("templates/loadbalancers.html", Content{Title: "Load Balancers", Type: "Load Balancer", URLKey: "load-balancer", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

// ===================================
// Auto Scaling

func launchconfigurationsPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	launchconfigurations, errs := GetLaunchConfigurations()

	data["LaunchConfigurations"] = launchconfigurations
	data["Errors"] = errs

	ctx.Render("templates/launchconfigurations.html", Content{Title: "Launch Configurations", Type: "Launch Configuration", URLKey: "launch-configuration", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

func autoscalegroupsPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	autoscalegroups, errs := GetAutoScaleGroups()

	data["AutoScaleGroups"] = autoscalegroups
	data["Errors"] = errs

	ctx.Render("templates/autoscalegroups.html", Content{Title: "Auto Scale Groups", Type: "Auto Scale Group", URLKey: "auto-scale-group", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

func scalingpoliciesPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	scalingpolicies, errs := GetScalingPolicies()

	data["ScalingPolicies"] = scalingpolicies
	data["Errors"] = errs

	ctx.Render("templates/scalingpolicies.html", Content{Title: "Scaling Policies", Type: "Scaling Policy", URLKey: "scaling-policy", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

// ===================================
// VPCs

func vpcsPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	vpcs, errs := GetVpcs("")

	data["VPCs"] = vpcs
	data["Errors"] = errs

	ctx.Render("templates/vpcs.html", Content{Title: "VPCs", Type: "VPC", URLKey: "vpc", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

func newVpcModal(ctx *iris.Context) {

	data := make(map[string]interface{})
	regionList := GetRegionList()

	configs, err := config.LoadAllVpcConfigs()
	if err != nil {
		data["Errors"] = []error{err}
	}

	data["Configs"] = configs
	data["Regions"] = regionList

	ctx.Render("templates/new-vpc-modal.html", Content{Title: "New Vpc", Type: "Vpc", URLKey: "vpc", Data: data})
}

func manageVpcClassesModal(ctx *iris.Context) {

	data := make(map[string]interface{})
	configs, err := config.GetAllConfigNames("vpc")
	if err != nil {
		data["Errors"] = []error{err}
	}

	data["Configs"] = configs

	ctx.Render("templates/manage-classes-modal.html", Content{Title: "Mange VPC Classes", Type: "VPC", URLKey: "vpc", Data: data})
}

func vpcClassForm(ctx *iris.Context) {
	data := make(map[string]interface{})
	class := ctx.Param("class")

	var cfg config.VpcClassConfig
	err := cfg.LoadConfig(class)
	if err != nil {
		data["Errors"] = []error{err}
	}

	data["ClassName"] = class
	data["ClassConfig"] = cfg

	ctx.Render("templates/vpc-class-form.html", Content{Title: "Edit VPC Class", Type: "VPC", URLKey: "vpc", Data: data})
}

// ===================================
// Subnets

func subnetsPage(ctx *iris.Context) {
	data := make(map[string]interface{})
	subnets, errs := GetSubnets("")

	data["Subnets"] = subnets
	data["Errors"] = errs

	ctx.Render("templates/subnets.html", Content{Title: "Subnets", Type: "Subnet", URLKey: "subnet", Data: data, RenderLayout: true, Host: ctx.HostString()})
}

func routetablesPage(ctx *iris.Context) {

}

func internetgatewaysPage(ctx *iris.Context) {

}

func dhcpoptionssetsPage(ctx *iris.Context) {

}

func elasticipsPage(ctx *iris.Context) {

}

// ===================================
// CloudWatch Alarms

func alarmsPage(ctx *iris.Context) {

	data := make(map[string]interface{})

	alarms, errs := GetAlarms()

	data["Alarms"] = alarms
	data["Errors"] = errs

	ctx.Render("templates/alarms.html", Content{Title: "Alarms", Type: "Alarm", URLKey: "alarm", Data: data, RenderLayout: true, Host: ctx.HostString()})
}
