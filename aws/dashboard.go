package aws

import (
	"os/user"

	"github.com/kataras/iris"
	"github.com/murdinc/awsm/config"
	"github.com/toqueteos/webbrowser"
)

type Content struct {
	Title    string
	Type     string
	Subtitle string
	Data     map[string]interface{}
	Configs  interface{}
	AZList   []string
	Modal    bool // determines if the layout renders or not
}

func RunDashboard(devMode bool) {
	currentUser, _ := user.Current()
	guiLocation := currentUser.HomeDir + "/.awsm/gui/awsm-default-gui"

	api := iris.New()

	api.Config().Render.Template.Directory = guiLocation
	api.Config().Render.Template.Layout = "templates/layout.html"
	api.StaticWeb("/js", guiLocation, 0)
	api.StaticWeb("/css", guiLocation, 0)
	api.StaticWeb("/fonts", guiLocation, 0)
	api.StaticWeb("/static", guiLocation, 0)

	// Index and Dashboard
	api.Get("/", index)
	api.Get("/dashboard", dashboard)

	api.Get("/dashboard/:page", dashboard)
	api.Get("/modal/:modal", modal)

	// Classes
	//iris.Get("/classes/:configType", getclasses)
	//iris.Put("/class/:configType", putclass)

	if !devMode {
		webbrowser.Open("http://localhost:8080/dashboard") // TODO race condition?
	}

	api.Listen(":8080") // TODO optionally configurable port #
}

func index(ctx *iris.Context) {
	ctx.Redirect("/dashboard")
}

func dashboard(ctx *iris.Context) {

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
		ctx.Render("templates/dashboard.html", Content{Title: "Dashboard", Type: "Dashboard"})
	}
}

func modal(ctx *iris.Context) {

	modal := ctx.Param("modal")

	switch modal {
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
	case "new-instance":
		newInstanceModal(ctx)
	case "manage-instance-classes":
		manageInstanceClassesModal(ctx)
	case "new-image":
		//newimageModal(ctx)
	case "new-volume":
		//newvolumeModal(ctx)
	case "new-snapshot":
		//newsnapshotModal(ctx)
	case "new-securitygroup":
		//newsecuritygroupModal(ctx)
	default:
		//ctx.Render("templates/404-modal.html", Content{Title: "404", Type: "404"})
	}
}

func instancesPage(ctx *iris.Context) {
	instances, errs := GetInstances("")

	data := make(map[string]interface{})
	data["Instances"] = instances

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering instance list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/instances.html", Content{Title: "Instances", Type: "Instance", Data: data})
}

func newInstanceModal(ctx *iris.Context) {

	configs, err := config.GetAllConfigNames("ec2")
	azList := AZList()

	data := make(map[string]interface{})
	data["Configs"] = configs
	data["AZList"] = azList

	if err != nil {
		ctx.Write("Error gathering instance class configs: %s\n", err.Error())
	}
	ctx.Render("templates/new-instance-modal.html", Content{Title: "New Instance", Type: "Instance", Data: data, Modal: true})
}

func manageInstanceClassesModal(ctx *iris.Context) {

	configs, err := config.GetAllConfigNames("ec2")

	data := make(map[string]interface{})
	data["Configs"] = configs

	if err != nil {
		ctx.Write("Error gathering instance class configs: %s\n", err.Error())
	}
	ctx.Render("templates/manage-instance-class-modal.html", Content{Title: "New Instance", Type: "Instance", Data: data, Modal: true})
}

func imagesPage(ctx *iris.Context) {
	images, errs := GetImages("")

	data := make(map[string]interface{})
	data["Images"] = images

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering image list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/images.html", Content{Title: "Images", Type: "Image", Subtitle: "Amazon Machine Images", Data: data})
}

func volumesPage(ctx *iris.Context) {
	volumes, errs := GetVolumes()

	data := make(map[string]interface{})
	data["Volumes"] = volumes

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering volume list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/volumes.html", Content{Title: "Volumes", Type: "Volume", Subtitle: "Amazon Elastic Block Storage Volumes", Data: data})
}

func snapshotsPage(ctx *iris.Context) {
	snapshots, errs := GetSnapshots()

	data := make(map[string]interface{})
	data["Snapshots"] = snapshots

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering snapshot list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/snapshots.html", Content{Title: "Snapshots", Type: "Snapshot", Subtitle: "Amazon Elastic Block Storage Snapshots", Data: data})
}

func securitygroupsPage(ctx *iris.Context) {
	securitygroups, errs := GetSecurityGroups()

	data := make(map[string]interface{})
	data["SecurityGroups"] = securitygroups

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering security group list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/securitygroups.html", Content{Title: "Security Groups", Type: "Security Group", Data: data})
}

func loadbalancersPage(ctx *iris.Context) {
	loadbalancers, errs := GetLoadBalancers()

	data := make(map[string]interface{})
	data["LoadBalancers"] = loadbalancers

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering load balancer list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/loadbalancers.html", Content{Title: "Load Balancers", Type: "Load Balancer", Data: data})
}

func launchconfigurationsPage(ctx *iris.Context) {
	launchconfigurations, errs := GetLaunchConfigurations()

	data := make(map[string]interface{})
	data["LaunchConfigurations"] = launchconfigurations

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering launch configurations list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/launchconfigurations.html", Content{Title: "Launch Configurations", Type: "Launch Configuration", Data: data})
}

func autoscalegroupsPage(ctx *iris.Context) {
	autoscalegroups, errs := GetAutoScaleGroups()

	data := make(map[string]interface{})
	data["AutoScaleGroups"] = autoscalegroups

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering auto scale group list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/autoscalegroups.html", Content{Title: "Auto Scale Groups", Type: "Auto Scale Group", Data: data})
}

func scalingpoliciesPage(ctx *iris.Context) {
	scalingpolicies, errs := GetScalingPolicies()

	data := make(map[string]interface{})
	data["ScalingPolicies"] = scalingpolicies

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering scaling policy list list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/scalingpolicies.html", Content{Title: "Scaling Policies", Type: "Scaling Policy", Data: data})
}

func subnetsPage(ctx *iris.Context) {
	subnets, errs := GetSubnets()

	data := make(map[string]interface{})
	data["Subnets"] = subnets

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering subnet list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/subnets.html", Content{Title: "Subnets", Type: "Subnet", Data: data})
}

func routetablesPage(ctx *iris.Context) {
	/*
		routetables, errs := GetRouteTables()
		if errs != nil {
			for _, err := range errs {
				ctx.Write("Error gathering alarm list: %s\n", err.Error())
			}
		}
		ctx.Render("routetables.html", Content{Title: "Route Tables", Subtitle: "Amazon Virtual Private Cloud Route Tables", Data: routetables})
	*/
}

func internetgatewaysPage(ctx *iris.Context) {
	/*
		internetgateways, errs := GetRouteTables()
		if errs != nil {
			for _, err := range errs {
				ctx.Write("Error gathering alarm list: %s\n", err.Error())
			}
		}
		ctx.Render("internetgateways.html", Content{Title: "Internet Gateways", Subtitle: "Amazon Virtual Private Cloud Internet Gateways", Data: internetgateways})
	*/
}

func dhcpoptionssetsPage(ctx *iris.Context) {
	/*
		dhcpoptionssets, errs := GetRouteTables()
		if errs != nil {
			for _, err := range errs {
				ctx.Write("Error gathering dhcp options sets list: %s\n", err.Error())
			}
		}
		ctx.Render("dhcpoptionssets.html", Content{Title: "DHCP Options Sets", Subtitle: "Amazon Virtual Private Cloud DHCP Options Sets", Data: dhcpoptionssets})
	*/
}

func elasticipsPage(ctx *iris.Context) {
	/*
		elasticips, errs := GetElasticIps()
		if errs != nil {
			for _, err := range errs {
				ctx.Write("Error gathering elastic ip list: %s\n", err.Error())
			}
		}
		ctx.Render("elasticips.html", Content{Title: "Elastic IPs", Subtitle: "Amazon Elastic IPs", Data: elasticips})
	*/
}

func alarmsPage(ctx *iris.Context) {
	alarms, errs := GetAlarms()

	data := make(map[string]interface{})
	data["Alarms"] = alarms

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering alarm list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/alarms.html", Content{Title: "Alarms", Type: "Alarm", Subtitle: "Amazon CloudWatch Alarms", Data: data})
}

func vpcsPage(ctx *iris.Context) {
	vpcs, errs := GetVpcs()

	data := make(map[string]interface{})
	data["VPCs"] = vpcs

	if errs != nil {
		for _, err := range errs {
			ctx.Write("Error gathering vpc list: %s\n", err.Error())
		}
	}
	ctx.Render("templates/vpcs.html", Content{Title: "VPCs", Type: "VPC", Subtitle: "Amazon Virtual Private Networks", Data: data})
}
