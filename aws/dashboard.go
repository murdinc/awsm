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

		/*
			case "new-loadbalancer":
				//newloadbalancerModal(ctx)
			case "new-autoscalegroup":
				//newautoscalegroupModal(ctx)
			case "new-launchconfiguration":
				//newlaunchconfigurationModal(ctx)
			case "new-scalingpolicy":
				//newscalingpolicieModal(ctx)
			case "new-alarm":
				//newalarmModal(ctx)
			case "new-vpc":
				//newvpcModal(ctx)
			case "new-subnet":
				//newsubnetModal(ctx)
			case "new-routetable":
				//newroutetableModal(ctx)
			case "new-internetgateway":
				//newinternetgatewayModal(ctx)
			case "new-dhcpoptionset":
				//newdhcpoptionssetModal(ctx)
			case "new-elasticip":
				//newelasticipModal(ctx)

			case "new-image":
				//newimageModal(ctx)
			case "new-volume":
				//newvolumeModal(ctx)
			case "new-snapshot":
				//newsnapshotModal(ctx)
			case "new-securitygroup":
				//newsecuritygroupModal(ctx)
		*/
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

		/*
			case "new-loadbalancer":
				//newloadbalancerModal(ctx)
			case "new-autoscalegroup":
				//newautoscalegroupModal(ctx)
			case "new-launchconfiguration":
				//newlaunchconfigurationModal(ctx)
			case "new-scalingpolicy":
				//newscalingpolicieModal(ctx)
			case "new-alarm":
				//newalarmModal(ctx)
			case "new-vpc":
				//newvpcModal(ctx)
			case "new-subnet":
				//newsubnetModal(ctx)
			case "new-routetable":
				//newroutetableModal(ctx)
			case "new-internetgateway":
				//newinternetgatewayModal(ctx)
			case "new-dhcpoptionset":
				//newdhcpoptionssetModal(ctx)
			case "new-elasticip":
				//newelasticipModal(ctx)
			case "new-image":
				//newimageModal(ctx)
			case "new-volume":
				//newvolumeModal(ctx)
			case "new-snapshot":
				//newsnapshotModal(ctx)
			case "new-securitygroup":
				//newsecuritygroupModal(ctx)
		*/
	default:
		//ctx.Render("templates/404-modal.html", Content{Title: "404", Type: "404"})
	}
}

// ===================================
// EC2 Instances

func instancesPage(ctx *iris.Context) {
	instances, errs := GetInstances("")

	data := make(map[string]interface{})
	data["Instances"] = instances

	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
			ctx.Write("Error gathering instance list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/instances.html", Content{Title: "Instances", Type: "Instance", Data: data, RenderLayout: true})
}

func newInstanceModal(ctx *iris.Context) {
	data := make(map[string]interface{})
	azList := AZList()
	configs, err := config.GetAllConfigNames("ec2")
	if err != nil {
		data["Errors"] = append(data["Errors"].([]string), err.Error())
	}

	data["Configs"] = configs
	data["AZList"] = azList

	ctx.Render("templates/new-instance-modal.html", Content{Title: "New Instance", Type: "Instance", Data: data})
}

func manageInstanceClassesModal(ctx *iris.Context) {
	data := make(map[string]interface{})
	configs, err := config.GetAllConfigNames("ec2")
	if err != nil {
		data["Errors"] = append(data["Errors"].([]string), err.Error())
	}

	data["Configs"] = configs

	ctx.Render("templates/manage-classes-modal.html", Content{Title: "Mange Instance Classes", Type: "Instance", Data: data, ClassFormURL: "edit-instance-class"})
}

func instanceClassForm(ctx *iris.Context) {
	data := make(map[string]interface{})
	class := ctx.Param("class")

	var cfg config.InstanceClassConfig
	err := cfg.LoadConfig(class)

	if err != nil {
		data["Errors"] = append(data["Errors"].([]string), err.Error())
	}

	data["ClassName"] = class
	data["ClassConfig"] = cfg

	ctx.Render("templates/instance-class-form.html", Content{Title: "Edit Instance Class", Type: "Instance", Data: data})
}

// ===================================
// AMI

func imagesPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	images, errs := GetImages("")
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["Images"] = images

	ctx.Render("templates/images.html", Content{Title: "Images", Type: "Image", Data: data, RenderLayout: true})
}

// ===================================
// EBS

func volumesPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	volumes, errs := GetVolumes()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["Volumes"] = volumes

	ctx.Render("templates/volumes.html", Content{Title: "Volumes", Type: "Volume", Data: data, RenderLayout: true})
}

func snapshotsPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	snapshots, errs := GetSnapshots()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["Snapshots"] = snapshots

	ctx.Render("templates/snapshots.html", Content{Title: "Snapshots", Type: "Snapshot", Data: data, RenderLayout: true})
}

// ===================================
// Security Groups

func securitygroupsPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	securitygroups, errs := GetSecurityGroups()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["SecurityGroups"] = securitygroups

	ctx.Render("templates/securitygroups.html", Content{Title: "Security Groups", Type: "Security Group", Data: data, RenderLayout: true})
}

// ===================================
// Load Balancers

func loadbalancersPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	loadbalancers, errs := GetLoadBalancers()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["LoadBalancers"] = loadbalancers

	ctx.Render("templates/loadbalancers.html", Content{Title: "Load Balancers", Type: "Load Balancer", Data: data, RenderLayout: true})
}

// ===================================
// Auto Scaling

func launchconfigurationsPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	launchconfigurations, errs := GetLaunchConfigurations()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["LaunchConfigurations"] = launchconfigurations

	ctx.Render("templates/launchconfigurations.html", Content{Title: "Launch Configurations", Type: "Launch Configuration", Data: data, RenderLayout: true})
}

func autoscalegroupsPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	autoscalegroups, errs := GetAutoScaleGroups()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["AutoScaleGroups"] = autoscalegroups

	ctx.Render("templates/autoscalegroups.html", Content{Title: "Auto Scale Groups", Type: "Auto Scale Group", Data: data, RenderLayout: true})
}

func scalingpoliciesPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	scalingpolicies, errs := GetScalingPolicies()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["ScalingPolicies"] = scalingpolicies

	ctx.Render("templates/scalingpolicies.html", Content{Title: "Scaling Policies", Type: "Scaling Policy", Data: data, RenderLayout: true})
}

// ===================================
// VPCs / Networking

func vpcsPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	vpcs, errs := GetVpcs()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["VPCs"] = vpcs

	ctx.Render("templates/vpcs.html", Content{Title: "VPCs", Type: "VPC", Data: data, RenderLayout: true})
}

func subnetsPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	subnets, errs := GetSubnets()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["Subnets"] = subnets

	ctx.Render("templates/subnets.html", Content{Title: "Subnets", Type: "Subnet", Data: data, RenderLayout: true})
}

func routetablesPage(ctx *iris.Context) {
	/*
		routetables, errs := GetRouteTables()
		if errs != nil {
			for _, err := range errs {
				data["Errors"] = append(data["Errors"].([]string), err.Error())
			}
		}
		ctx.Render("routetables.html", Content{Title: "Route Tables", Data: routetables, RenderLayout: true})
	*/
}

func internetgatewaysPage(ctx *iris.Context) {
	/*
		internetgateways, errs := GetRouteTables()
		if errs != nil {
			for _, err := range errs {
				data["Errors"] = append(data["Errors"].([]string), err.Error())
			}
		}
		ctx.Render("internetgateways.html", Content{Title: "Internet Gateways", Data: internetgateways, RenderLayout: true})
	*/
}

func dhcpoptionssetsPage(ctx *iris.Context) {
	/*
		dhcpoptionssets, errs := GetRouteTables()
		if errs != nil {
			for _, err := range errs {
				data["Errors"] = append(data["Errors"].([]string), err.Error())
			}
		}
		ctx.Render("dhcpoptionssets.html", Content{Title: "DHCP Options Sets", Data: dhcpoptionssets, RenderLayout: true})
	*/
}

func elasticipsPage(ctx *iris.Context) {
	/*
		elasticips, errs := GetElasticIps()
		if errs != nil {
			for _, err := range errs {
				data["Errors"] = append(data["Errors"].([]string), err.Error())
			}
		}
		ctx.Render("elasticips.html", Content{Title: "Elastic IPs", Data: elasticips, RenderLayout: true})
	*/
}

// ===================================
// CloudWatch Alarms

func alarmsPage(ctx *iris.Context) {
	data := make(map[string]interface{})

	alarms, errs := GetAlarms()
	if errs != nil {
		for _, err := range errs {
			data["Errors"] = append(data["Errors"].([]string), err.Error())
		}
	}

	data["Alarms"] = alarms

	ctx.Render("templates/alarms.html", Content{Title: "Alarms", Type: "Alarm", Data: data, RenderLayout: true})
}
