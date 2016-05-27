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
	Data         map[string]interface{}
	Configs      interface{}
	AZList       []string
	RenderLayout bool
	Errors       []string
	ClassFormURL string
}

func RunDashboard(devMode bool) {
	currentUser, _ := user.Current()
	guiLocation := currentUser.HomeDir + "/.awsm/gui/awsm-default-gui" // TODO accept custom theme directories

	api := iris.New()

	// Template Configuration
	api.Config().Render.Template.Directory = guiLocation
	api.Config().Render.Template.Layout = "templates/layout.html"

	// Static Asset Folders
	api.StaticWeb("/js", guiLocation, 0)
	api.StaticWeb("/css", guiLocation, 0)
	api.StaticWeb("/fonts", guiLocation, 0)
	api.StaticWeb("/static", guiLocation, 0)

	// Index and Dashboard
	api.Get("/", index)
	api.Get("/dashboard", getDashboard)

	// Template builders
	api.Get("/dashboard/:page", getDashboard)
	api.Get("/modal/:modal", getModal)
	api.Get("/form/:form/:class", getForm)

	// Form Handlers
	//api.Post("/form/:form", postForm)

	if !devMode {
		webbrowser.Open("http://localhost:8080/dashboard") // TODO race condition?
	}

	api.Listen(":8080") // TODO optionally configurable port #
}

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
		ctx.Render("templates/dashboard.html", Content{Title: "Dashboard", Type: "Dashboard", RenderLayout: true})
	}
}

// ===================================
// Builds all the different modals
func getModal(ctx *iris.Context) {

	modal := ctx.Param("modal")

	switch modal {

	// EC2 Instances
	case "new-instance":
		newInstanceModal(ctx)
	case "manage-instance-classes":
		manageInstanceClassesModal(ctx)

	// VPC
	case "new-vpc":
		newVpcModal(ctx)
	case "manage-vpc-classes":
		manageVpcClassesModal(ctx)

	default:
		//ctx.Render("templates/404-modal.html", Content{Title: "404", Type: "404"})
	}
}

// ===================================
// Builds all the different forms
func getForm(ctx *iris.Context) {

	form := ctx.Param("form")

	switch form {

	// EC2 Instances
	case "edit-instance-class":
		instanceClassForm(ctx)

	// VPC
	case "edit-vpc-class":
		vpcClassForm(ctx)

	default:
		//ctx.Render("templates/404-modal.html", Content{Title: "404", Type: "404"})
	}
}

// ===================================
// EC2 Instances

func instancesPage(ctx *iris.Context) {
	var errSlice []string
	instances, errs := GetInstances("")

	data := make(map[string]interface{})

	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
			ctx.Write("Error gathering instance list: %s\n", err.Error())
		}
	}

	data["Instances"] = instances
	data["Errors"] = errSlice

	ctx.Render("templates/instances.html", Content{Title: "Instances", Type: "Instance", Data: data, RenderLayout: true})
}

func newInstanceModal(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})
	azList := AZList()
	configs, err := config.GetAllConfigNames("ec2")
	if err != nil {
		errSlice = append(errSlice, err.Error())
	}

	data["Configs"] = configs
	data["AZList"] = azList
	data["Errors"] = errSlice

	ctx.Render("templates/new-instance-modal.html", Content{Title: "New Instance", Type: "Instance", Data: data})
}

func manageInstanceClassesModal(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})
	configs, err := config.GetAllConfigNames("ec2")
	if err != nil {
		errSlice = append(errSlice, err.Error())
	}

	data["Configs"] = configs
	data["Errors"] = errSlice

	ctx.Render("templates/manage-classes-modal.html", Content{Title: "Mange Instance Classes", Type: "Instance", Data: data, ClassFormURL: "edit-instance-class"})
}

func instanceClassForm(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})
	class := ctx.Param("class")

	var cfg config.InstanceClassConfig
	err := cfg.LoadConfig(class)

	if err != nil {
		errSlice = append(errSlice, err.Error())
	}

	data["ClassName"] = class
	data["ClassConfig"] = cfg
	data["Errors"] = errSlice

	ctx.Render("templates/instance-class-form.html", Content{Title: "Edit Instance Class", Type: "Instance", Data: data})
}

// ===================================
// AMI

func imagesPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	images, errs := GetImages("")
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["Images"] = images
	data["Errors"] = errSlice

	ctx.Render("templates/images.html", Content{Title: "Images", Type: "Image", Data: data, RenderLayout: true})
}

// ===================================
// EBS

func volumesPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	volumes, errs := GetVolumes()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["Volumes"] = volumes
	data["Errors"] = errSlice

	ctx.Render("templates/volumes.html", Content{Title: "Volumes", Type: "Volume", Data: data, RenderLayout: true})
}

func snapshotsPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	snapshots, errs := GetSnapshots()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["Snapshots"] = snapshots
	data["Errors"] = errSlice

	ctx.Render("templates/snapshots.html", Content{Title: "Snapshots", Type: "Snapshot", Data: data, RenderLayout: true})
}

// ===================================
// Security Groups

func securitygroupsPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	securitygroups, errs := GetSecurityGroups()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["SecurityGroups"] = securitygroups
	data["Errors"] = errSlice

	ctx.Render("templates/securitygroups.html", Content{Title: "Security Groups", Type: "Security Group", Data: data, RenderLayout: true})
}

// ===================================
// Load Balancers

func loadbalancersPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	loadbalancers, errs := GetLoadBalancers()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["LoadBalancers"] = loadbalancers
	data["Errors"] = errSlice

	ctx.Render("templates/loadbalancers.html", Content{Title: "Load Balancers", Type: "Load Balancer", Data: data, RenderLayout: true})
}

// ===================================
// Auto Scaling

func launchconfigurationsPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	launchconfigurations, errs := GetLaunchConfigurations()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["LaunchConfigurations"] = launchconfigurations
	data["Errors"] = errSlice

	ctx.Render("templates/launchconfigurations.html", Content{Title: "Launch Configurations", Type: "Launch Configuration", Data: data, RenderLayout: true})
}

func autoscalegroupsPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	autoscalegroups, errs := GetAutoScaleGroups()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["AutoScaleGroups"] = autoscalegroups
	data["Errors"] = errSlice

	ctx.Render("templates/autoscalegroups.html", Content{Title: "Auto Scale Groups", Type: "Auto Scale Group", Data: data, RenderLayout: true})
}

func scalingpoliciesPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	scalingpolicies, errs := GetScalingPolicies()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["ScalingPolicies"] = scalingpolicies
	data["Errors"] = errSlice

	ctx.Render("templates/scalingpolicies.html", Content{Title: "Scaling Policies", Type: "Scaling Policy", Data: data, RenderLayout: true})
}

// ===================================
// VPCs

func vpcsPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	vpcs, errs := GetVpcs()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["VPCs"] = vpcs
	data["Errors"] = errSlice

	ctx.Render("templates/vpcs.html", Content{Title: "VPCs", Type: "VPC", Data: data, RenderLayout: true})
}

func newVpcModal(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})
	regionList := GetRegionList()

	configs, err := config.LoadAllVpcConfigs()
	if err != nil {
		errSlice = append(errSlice, err.Error())
	}

	data["Configs"] = configs
	data["Regions"] = regionList
	data["Errors"] = errSlice

	ctx.Render("templates/new-vpc-modal.html", Content{Title: "New Vpc", Type: "Vpc", Data: data})
}

func manageVpcClassesModal(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})
	configs, err := config.GetAllConfigNames("vpc")
	if err != nil {
		errSlice = append(errSlice, err.Error())
	}

	data["Configs"] = configs
	data["Errors"] = errSlice

	ctx.Render("templates/manage-classes-modal.html", Content{Title: "Mange Instance Classes", Type: "Instance", Data: data, ClassFormURL: "edit-vpc-class"})
}

func vpcClassForm(ctx *iris.Context) {
	var errSlice []string
	data := make(map[string]interface{})
	class := ctx.Param("class")

	var cfg config.VpcClassConfig
	err := cfg.LoadConfig(class)

	if err != nil {
		errSlice = append(errSlice, err.Error())
	}

	data["ClassName"] = class
	data["ClassConfig"] = cfg
	data["Errors"] = errSlice

	ctx.Render("templates/vpc-class-form.html", Content{Title: "Edit VPC Class", Type: "VPC", Data: data})
}

// ===================================
// Subnets

func subnetsPage(ctx *iris.Context) {
	var errSlice []string

	data := make(map[string]interface{})

	subnets, errs := GetSubnets()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["Subnets"] = subnets
	data["Errors"] = errSlice

	ctx.Render("templates/subnets.html", Content{Title: "Subnets", Type: "Subnet", Data: data, RenderLayout: true})
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
	var errSlice []string

	data := make(map[string]interface{})

	alarms, errs := GetAlarms()
	if errs != nil {
		for _, err := range errs {
			errSlice = append(errSlice, err.Error())
		}
	}

	data["Alarms"] = alarms
	data["Errors"] = errSlice

	ctx.Render("templates/alarms.html", Content{Title: "Alarms", Type: "Alarm", Data: data, RenderLayout: true})
}
