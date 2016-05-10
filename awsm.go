package main

import (
	"os"

	"github.com/murdinc/awsm/aws"
	"github.com/murdinc/awsm/terminal"
	"github.com/murdinc/cli"
)

// Main Function
////////////////..........
func main() {

	found := aws.CheckCreds()
	if !found {
		return
	}

	var dryRun bool

	app := cli.NewApp()
	app.Name = "awsm"
	app.Usage = "AWS iMproved CLI"
	app.Version = "1.0"
	app.Author = "Ahmad A"
	app.Email = "send@ahmad.pizza"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "dry-run",
			Destination: &dryRun,
			Usage:       "dry-run (Don't make any real changes)",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "test",
			Usage: "TEST",
			Action: func(c *cli.Context) error {
				// TODO
				//config.CheckConfig()
				return nil
			},
		},
		{
			Name:  "attachVolume",
			Usage: "Attach an AWS EBS Volume",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "copyImage",
			Usage: "Copy an AWS Machine Image to another region",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "copySnapshot",
			Usage: "Copy an AWS EBS Snapshot to another region",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "createAddress",
			Usage: "Create an AWS Elastic IP Address (for use in a VPC or EC2-Classic)",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "createAutoScaleGroup",
			Usage: "Create an AWS AutoScaling Group",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "createImage",
			Usage: "Create an AWS Machine Image from a running instance",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "createLaunchConfiguration",
			Usage: "Create an AWS AutoScaling Launch Configuration",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "createSimpleDBDomain",
			Usage: "Create an AWS SimpleDB Domain",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "domain",
					Description: "The domain of the db",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the db",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateSimpleDBDomain(c.NamedArg("domain"), c.NamedArg("region"))
				if err != nil {
					terminal.ShowErrorMessage("Error", err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createSnapshot",
			Usage: "Create an AWS EBS snapshot of a volume",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "createVolume",
			Usage: "Create an AWS EBS volume (from a class snapshot or blank)",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "deleteAutoScaleGroup",
			Usage: "Delete an AWS AutoScaling Group",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "deleteImage",
			Usage: "Delete an AWS Machine Image",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "deleteLaunchConfiguration",
			Usage: "Delete an AWS AutoScaling Launch Configuration",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "deleteSnapshot",
			Usage: "Delete an AWS EBS Snapshot",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "deleteSimpleDBDomains",
			Usage: "Delete an AWS SimpleDB Domain",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "domain",
					Description: "The domain of the db",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the db (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteSimpleDBDomain(c.NamedArg("domain"), c.NamedArg("region"))
				if err != nil {
					terminal.ShowErrorMessage("Error", err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteVolume",
			Usage: "Delete an AWS EBS Volume",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "detachVolume",
			Usage: "Detach an AWS EBS Volume",
			Action: func(c *cli.Context) error {
				// TODO

				return nil
			},
		},
		{
			Name:  "stopInstances",
			Usage: "Stop AWS instance(s)",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "pauseInstances",
			Usage: "Pause AWS instance(s)",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "killInstances",
			Usage: "Kill AWS instance(s)",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "launchInstance",
			Usage: "Launch an EC2 instance",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "class",
					Description: "The class of the instance (dev, stage, etc)",
					Optional:    false,
				},
				cli.Argument{
					Name:        "sequence",
					Description: "The sequence of the instance (1...100)",
					Optional:    false,
				},
				cli.Argument{
					Name:        "az",
					Description: "The availability zone to launch the instance in (us-west-2a, us-east-1a, etc)",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.LaunchInstance(c.NamedArg("class"), c.NamedArg("sequence"), c.NamedArg("az"), dryRun)
				if err != nil {
					terminal.ShowErrorMessage("Error", err.Error())
				}
				return nil
			},
		},
		{
			Name:  "listAddresses",
			Usage: "Lists all AWS Elastic IP Addresses",
			Action: func(c *cli.Context) error {
				addresses, err := aws.GetAddresses()
				if err != nil {
					terminal.ShowErrorMessage("Error", err.Error())
					return nil
				} else {
					addresses.PrintTable()
				}
				return nil
			},
		},
		{
			Name:  "listAlarms",
			Usage: "Lists all CloudWatch Alarms",
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
			Name:  "listAutoScaleGroups",
			Usage: "Lists all AutoScale Groups",
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
			Name:  "listImages",
			Usage: "Lists all AWS Machine Images owned by us",
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
			Name:  "listInstances",
			Usage: "Lists all AWS EC2 Instances",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				instances, err := aws.GetInstances(c.NamedArg("search"))
				if err != nil {
					return cli.NewExitError("Error Listing Instances!", 1)
				} else {
					instances.PrintTable()
				}
				return nil
			},
		},
		{
			Name:  "listLaunchConfigurations",
			Usage: "Lists all Launch Configurations",
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
			Name:  "listLoadBalancers",
			Usage: "Lists all Elastic Load Balancers",
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
			Name:  "listScalingPolicies",
			Usage: "Lists all Scaling Policies",
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
			Name:  "listSecurityGroups",
			Usage: "Lists all Security Groups",
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
			Name:  "listSnapshots",
			Usage: "Lists all AWS EBS Snapshots",
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
			Name:  "listSubnets",
			Usage: "Lists all AWS Subnets",
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
			Name:  "listSimpleDBDomains",
			Usage: "Lists all AWS SimpleDB Domains",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				domains, err := aws.GetSimpleDBDomains(c.NamedArg("search"))
				if err != nil {
					return cli.NewExitError("Error Listing Simple DB Domains!", 1)
				} else {
					domains.PrintTable()
				}
				return nil
			},
		},
		{
			Name:  "listVolumes",
			Usage: "Lists all AWS EBS Volumes",
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
			Name:  "listVpcs",
			Usage: "Lists all AWS Vpcs",
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
			Name:  "resumeProcesses",
			Usage: "Resume all autoscaling processes on a specific autoscaling group",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "runCommand",
			Usage: "Run a command on a set of instances",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "suspendProcesses",
			Usage: "Stop all autoscaling processes on a specific autoscaling group",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "updateAutoScaleGroup",
			Usage: "Update an AWS AutoScaling Group",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
	}

	app.Run(os.Args)
}
