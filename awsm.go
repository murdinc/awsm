package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/murdinc/awsm/aws"
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
	app.Usage = "AWS iMproved CLI"
	app.Version = "1.0"
	app.Author = "Ahmad A"
	app.Email = "send@ahmad.pizza"
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:        "attachVolume",
			ShortName:   "",
			Description: "Attach an AWS EBS Volume",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "copyImage",
			ShortName:   "",
			Description: "Copy an AWS Machine Image to another region",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "copySnapshot",
			ShortName:   "",
			Description: "Copy an AWS EBS Snapshot to another region",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "createAddress",
			ShortName:   "",
			Description: "Create an AWS Elastic IP Address (for use in a VPC or EC2-Classic)",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "createAutoScaleGroup",
			ShortName:   "",
			Description: "Create an AWS AutoScaling Group",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "createImage",
			ShortName:   "",
			Description: "Create an AWS Machine Image from a running instance",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "createLaunchConfiguration",
			ShortName:   "",
			Description: "Create an AWS AutoScaling Launch Configuration",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "createSnapshot",
			ShortName:   "",
			Description: "Create an AWS EBS snapshot of a volume",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "createVolume",
			ShortName:   "",
			Description: "Create an AWS EBS volume (from a class snapshot or blank)",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "deleteAutoScaleGroup",
			ShortName:   "",
			Description: "Delete an AWS AutoScaling Group",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "deleteImage",
			ShortName:   "",
			Description: "Delete an AWS Machine Image",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "deleteLaunchConfiguration",
			ShortName:   "",
			Description: "Delete an AWS AutoScaling Launch Configuration",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "deleteSnapshot",
			ShortName:   "",
			Description: "Delete an AWS EBS Snapshot",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "deleteVolume",
			ShortName:   "",
			Description: "Delete an AWS EBS Volume",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "detachVolume",
			ShortName:   "",
			Description: "Detach an AWS EBS Volume",
			Action: func(c *cli.Context) error {
				// anotha one

				return nil
			},
		},
		{
			Name:        "stopInstances",
			ShortName:   "",
			Description: "Stop AWS instance(s)",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "pauseInstances",
			ShortName:   "",
			Description: "Pause AWS instance(s)",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "launchInstance",
			ShortName:   "",
			Description: "Launch an EC2 instance",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "listAddresses",
			ShortName:   "",
			Description: "Lists all AWS Elastic IP Addresses",
			Action: func(c *cli.Context) error {
				addresses, err := aws.GetAddresses()
				if err != nil {
					return cli.NewExitError("Error Listing Elastic IP Addresses!", 1)
				} else {
					addresses.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listAlarms",
			ShortName:   "",
			Description: "Lists all CloudWatch Alarms",
			Action: func(c *cli.Context) error {
				alarms, err := aws.GetAlarms()
				if err != nil {
					return cli.NewExitError("Error Listing Alarms!", 1)
				} else {
					alarms.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listAutoScaleGroups",
			ShortName:   "",
			Description: "Lists all AutoScale Groups",
			Action: func(c *cli.Context) error {
				groups, err := aws.GetAutoScaleGroups()
				if err != nil {
					return cli.NewExitError("Error Listing Auto Scale Groups!", 1)
				} else {
					groups.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listImages",
			ShortName:   "",
			Description: "Lists all AWS Machine Images owned by us",
			Action: func(c *cli.Context) error {
				images, err := aws.GetImages()
				if err != nil {
					return cli.NewExitError("Error Listing Images!", 1)
				} else {
					images.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listInstances",
			ShortName:   "",
			Description: "Lists all AWS EC2 Instances",
			Action: func(c *cli.Context) error {
				instances, err := aws.GetInstances()
				if err != nil {
					return cli.NewExitError("Error Listing Instances!", 1)
				} else {
					instances.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listLaunchConfigurations",
			ShortName:   "",
			Description: "Lists all Launch Configurations",
			Action: func(c *cli.Context) error {
				launchConfigs, err := aws.GetLaunchConfigurations()
				if err != nil {
					return cli.NewExitError("Error Listing Launch Configurations!", 1)
				} else {
					launchConfigs.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listLoadBalancers",
			ShortName:   "",
			Description: "Lists all Elastic Load Balancers",
			Action: func(c *cli.Context) error {
				loadBalancers, err := aws.GetLoadBalancers()
				if err != nil {
					return cli.NewExitError("Error Listing Load Balancers!", 1)
				} else {
					loadBalancers.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listScalingPolicies",
			ShortName:   "",
			Description: "Lists all Scaling Policies",
			Action: func(c *cli.Context) error {
				policies, err := aws.GetScalingPolicies()
				if err != nil {
					return cli.NewExitError("Error Listing Auto Scaling Policies!", 1)
				} else {
					policies.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listSecurityGroups",
			ShortName:   "",
			Description: "Lists all Security Groups",
			Action: func(c *cli.Context) error {
				groups, err := aws.GetSecurityGroups()
				if err != nil {
					return cli.NewExitError("Error Listing Security Groups!", 1)
				} else {
					groups.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listSnapshots",
			ShortName:   "",
			Description: "Lists all AWS EBS Snapshots",
			Action: func(c *cli.Context) error {
				snapshots, err := aws.GetSnapshots()
				if err != nil {
					return cli.NewExitError("Error Listing Snapshots!", 1)
				} else {
					snapshots.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listSubnets",
			ShortName:   "",
			Description: "Lists all AWS Subnets",
			Action: func(c *cli.Context) error {
				subnets, err := aws.GetSubnets()
				if err != nil {
					return cli.NewExitError("Error Listing Subnets!", 1)
				} else {
					subnets.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listVolumes",
			ShortName:   "",
			Description: "Lists all AWS EBS Volumes",
			Action: func(c *cli.Context) error {
				volumes, err := aws.GetVolumes()
				if err != nil {
					return cli.NewExitError("Error Listing Volumes!", 1)
				} else {
					volumes.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "listVpcs",
			ShortName:   "",
			Description: "Lists all AWS Vpcs",
			Action: func(c *cli.Context) error {
				vpcs, err := aws.GetVpcs()
				if err != nil {
					return cli.NewExitError("Error Listing VPCs!", 1)
				} else {
					vpcs.PrintTable()
				}
				return nil
			},
		},
		{
			Name:        "resumeProcesses",
			ShortName:   "",
			Description: "Resume all autoscaling processes on a specific autoscaling group",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "runCommand",
			ShortName:   "",
			Description: "Run a command on a set of instances",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "suspendProcesses",
			ShortName:   "",
			Description: "Stop all autoscaling processes on a specific autoscaling group",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
			},
		},
		{
			Name:        "updateAutoScaleGroup",
			ShortName:   "",
			Description: "Update an AWS AutoScaling Group",
			Action: func(c *cli.Context) error {
				// anotha one
				return nil
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
