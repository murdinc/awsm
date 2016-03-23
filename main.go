package main

import (
	"os"

	"github.com/murdinc/awsm/aws"
	"github.com/murdinc/cli"
)

// Main Function
////////////////..........
func main() {

	found := aws.CheckCreds()
	if !found {
		return
	}

	app := cli.NewApp()
	app.Name = "awsm"
	app.Usage = "AWS Improved Interface"
	app.Version = "1.0"
	app.Author = "Ahmad Abdo"
	app.Email = "send@ahmad.pizza"

	app.Commands = []cli.Command{
		{
			Name:        "attachVolume",
			ShortName:   "",
			Example:     "",
			Description: "Attach an AWS EBS Volume",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "copyImage",
			ShortName:   "",
			Example:     "",
			Description: "Copy an AWS Machine Image to another region",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "copySnapshot",
			ShortName:   "",
			Example:     "",
			Description: "Copy an AWS EBS Snapshot to another region",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "createAddress",
			ShortName:   "",
			Example:     "",
			Description: "Create an AWS Elastic IP Address (for use in a VPC or EC2-Classic)",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "createAutoScaleGroup",
			ShortName:   "",
			Example:     "",
			Description: "Create an AWS AutoScaling Group",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "createImage",
			ShortName:   "",
			Example:     "",
			Description: "Create an AWS Machine Image from a running instance",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "createLaunchConfiguration",
			ShortName:   "",
			Example:     "",
			Description: "Create an AWS AutoScaling Launch Configuration",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "createSnapshot",
			ShortName:   "",
			Example:     "",
			Description: "Create an AWS EBS snapshot of a volume",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "createVolume",
			ShortName:   "",
			Example:     "",
			Description: "Create an AWS EBS volume (from a class snapshot or blank)",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "deleteAutoScaleGroup",
			ShortName:   "",
			Example:     "",
			Description: "Delete an AWS AutoScaling Group",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "deleteImage",
			ShortName:   "",
			Example:     "",
			Description: "Delete an AWS Machine Image",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "deleteLaunchConfiguration",
			ShortName:   "",
			Example:     "",
			Description: "Delete an AWS AutoScaling Launch Configuration",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "deleteSnapshot",
			ShortName:   "",
			Example:     "",
			Description: "Delete an AWS EBS Snapshot",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "deleteVolume",
			ShortName:   "",
			Example:     "",
			Description: "Delete an AWS EBS Volume",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "detachVolume",
			ShortName:   "",
			Example:     "",
			Description: "Detach an AWS EBS Volume",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		/*
			{
				Name:        "doHost",
				ShortName:   "",
				Example:     "",
				Description: "Add/Update/Delete Route53 Records based on config data and AWS resources",
				Action: func(c *cli.Context) {
					// anotha one
				},
			},
		*/
		{
			Name:        "stopInstances",
			ShortName:   "",
			Example:     "",
			Description: "Stop AWS instance(s)",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "pauseInstances",
			ShortName:   "",
			Example:     "",
			Description: "Pause AWS instance(s)",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "launchInstance",
			ShortName:   "",
			Example:     "",
			Description: "Launch an EC2 instance",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "listAddresses",
			ShortName:   "",
			Example:     "",
			Description: "Lists all AWS Elastic IP Addresses",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "listAlarms",
			ShortName:   "",
			Example:     "",
			Description: "Lists all CloudWatch Alarms",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "listAutoScaleGroups",
			ShortName:   "",
			Example:     "",
			Description: "Lists all AutoScale Groups",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "listImages",
			ShortName:   "",
			Example:     "",
			Description: "Lists all AWS Machine Images owned by us",
			Action: func(c *cli.Context) {
				images, err := aws.GetImages()
				if err != nil {
					cli.ShowErrorMessage("Error Listing Images!", err.Error())
				} else {
					images.PrintTable()
				}
			},
		},
		{
			Name:        "listInstances",
			ShortName:   "",
			Example:     "",
			Description: "Lists all AWS EC2 Instances",
			Action: func(c *cli.Context) {
				instances, err := aws.GetInstances()
				if err != nil {
					cli.ShowErrorMessage("Error Listing Instances!", err.Error())
				} else {
					instances.PrintTable()
				}
			},
		},
		{
			Name:        "listLaunchConfigurations",
			ShortName:   "",
			Example:     "",
			Description: "Lists all Launch Configurations",
			Action: func(c *cli.Context) {
				launchConfigs, err := aws.GetLaunchConfigurations()
				if err != nil {
					cli.ShowErrorMessage("Error Listing Launch Configurations!", err.Error())
				} else {
					launchConfigs.PrintTable()
				}
			},
		},
		{
			Name:        "listLoadBalancers",
			ShortName:   "",
			Example:     "",
			Description: "Lists all Elastic Load Balancers",
			Action: func(c *cli.Context) {
				loadBalancers, err := aws.GetLoadBalancers()
				if err != nil {
					cli.ShowErrorMessage("Error Listing Load Balancers!", err.Error())
				} else {
					loadBalancers.PrintTable()
				}
			},
		},
		{
			Name:        "listScalingPolicies",
			ShortName:   "",
			Example:     "",
			Description: "Lists all Scaling Policies",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "listSecurityGroups",
			ShortName:   "",
			Example:     "",
			Description: "Lists all Security Groups",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "listSnapshots",
			ShortName:   "",
			Example:     "",
			Description: "Lists all AWS EBS Snapshots",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "listSubnets",
			ShortName:   "",
			Example:     "",
			Description: "Lists all AWS Subnets",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "listVolumes",
			ShortName:   "",
			Example:     "",
			Description: "Lists all AWS EBS Volumes",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "listVpcs",
			ShortName:   "",
			Example:     "",
			Description: "Lists all AWS Vpcs",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "resumeProcesses",
			ShortName:   "",
			Example:     "",
			Description: "Resume all autoscaling processes on a specific autoscaling group",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "runCommand",
			ShortName:   "",
			Example:     "",
			Description: "Run a command on a set of instances",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "suspendProcesses",
			ShortName:   "",
			Example:     "",
			Description: "Stop all autoscaling processes on a specific autoscaling group",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
		{
			Name:        "updateAutoScaleGroup",
			ShortName:   "",
			Example:     "",
			Description: "Update an AWS AutoScaling Group",
			Action: func(c *cli.Context) {
				// anotha one
			},
		},
	}

	app.Run(os.Args)
}

/*
func getConfig() *config.CrusherConfig {
	// Check Config
	cfg, err := config.ReadConfig()
	if err != nil || len(cfg.Servers) == 0 {
		// No Config Found, ask if we want to create one
		create := cli.BoxPromptBool("Crusher configuration file not found or empty!", "Do you want to add some servers now?")
		if !create {
			cli.Information("Alright then, maybe next time.. ")
			os.Exit(0)
		}
		// Add Some Servers to our config
		cfg.AddServer()
		os.Exit(0)
	}

	return cfg
}
*/
