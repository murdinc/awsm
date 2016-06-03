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
			Name:  "dashboard",
			Usage: "Launch the awsm Dashboard GUI",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "dev-mode",
					Destination: &dryRun,
					Usage:       "dev-mode (Don't reopen dashboard on each restart)",
				},
			},
			Action: func(c *cli.Context) error {
				aws.RunDashboard(c.Bool("dev-mode"))
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
			Name:  "createIAMUser",
			Usage: "Create an IAM User",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "username",
					Description: "The username for this IAM user",
					Optional:    false,
				},
				cli.Argument{
					Name:        "path",
					Description: "The optional path for the user",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateIAMUser(c.NamedArg("username"), c.NamedArg("path"))
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
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
			Name:  "createKeyPair",
			Usage: "Create and upload an AWS Key Pair",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "name",
					Description: "The name of the key pair",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				errs := aws.CreateAndImportKeyPair(c.NamedArg("name"), dryRun)
				if errs != nil {
					return cli.NewExitError("Error Creating KeyPair!", 1)
				} else {
					terminal.Information("Done!")
				}
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
					terminal.ErrorLine(err.Error())
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
			Name:  "createVpc",
			Usage: "Create an AWS VPC",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "class",
					Description: "The class of VPC to create",
					Optional:    false,
				},
				cli.Argument{
					Name:        "name",
					Description: "The name of the VPC",
					Optional:    false,
				},
				cli.Argument{
					Name:        "ip",
					Description: "The IP address of this VPC (not including CIDR)",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region to create the VPC in",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateVpc(c.NamedArg("class"), c.NamedArg("name"), c.NamedArg("ip"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createSubnet",
			Usage: "Create an AWS VPC Subnet",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "class",
					Description: "The class of Subnet to create",
					Optional:    false,
				},
				cli.Argument{
					Name:        "name",
					Description: "The name of the Subnet",
					Optional:    false,
				},
				cli.Argument{
					Name:        "vpc",
					Description: "The VPC to create the Subnet in",
					Optional:    false,
				},
				cli.Argument{
					Name:        "ip",
					Description: "The IP address of this Subnet (not including CIDR)",
					Optional:    false,
				},
				cli.Argument{
					Name:        "az",
					Description: "The Availability Zone to create the Subnet in",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateSubnet(c.NamedArg("class"), c.NamedArg("name"), c.NamedArg("vpc"), c.NamedArg("ip"), c.NamedArg("az"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
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
			Name:  "deleteIAMUser",
			Usage: "Delete an AWS Machine Image",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "username",
					Description: "The username of the IAM User to delete",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteIAMUser(c.NamedArg("username"))
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
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
			Name:  "deleteKeyPairs",
			Usage: "Delete an AWS KeyPair",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "name",
					Description: "The name of the AWS KeyPair to delete",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				errs := aws.DeleteKeyPairs(c.NamedArg("name"), dryRun)
				if errs != nil {
					return cli.NewExitError("Errors Deleting KeyPair!", 1)
				} else {
					terminal.Information("Done!")
				}
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
					terminal.ErrorLine(err.Error())
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
			Name:  "deleteSubnets",
			Usage: "Delete AWS VPC Subnets",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "name",
					Description: "The name of the Subnet",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the subnet (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteSubnets(c.NamedArg("name"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteVpcs",
			Usage: "Delete AWS VPCs",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "name",
					Description: "The name of the VPC",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the vpc (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteVpcs(c.NamedArg("name"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
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
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "name",
					Description: "The name of the Instance",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				errs := aws.KillInstances(c.NamedArg("name"), c.NamedArg("region"), dryRun)
				if errs != nil {
					return cli.NewExitError("Error Terminating Instances!", 1)
				}
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
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "listAddresses",
			Usage: "Lists all AWS Elastic IP Addresses",
			Action: func(c *cli.Context) error {
				addresses, errs := aws.GetAddresses()
				if errs != nil {
					return cli.NewExitError("Error Listing Addresses!", 1)
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
				alarms, errs := aws.GetAlarms()
				if errs != nil {
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
				groups, errs := aws.GetAutoScaleGroups()
				if errs != nil {
					return cli.NewExitError("Error Listing Auto Scale Groups!", 1)
				} else {
					groups.PrintTable()
				}
				return nil
			},
		},
		{
			Name:  "listIAMUsers",
			Usage: "Lists all IAM Users",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				iam, errs := aws.GetIAMUsers(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing IAM Users!", 1)
				} else {
					iam.PrintTable()
				}
				return nil
			},
		},
		{
			Name:  "listImages",
			Usage: "Lists all AWS Machine Images owned by us",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				images, errs := aws.GetImages(c.NamedArg("search"))
				if errs != nil {
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
				instances, errs := aws.GetInstances(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Instances!", 1)
				} else {
					instances.PrintTable()
				}
				return nil
			},
		},
		{
			Name:  "listKeyPairs",
			Usage: "Lists all AWS Key Pairs",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				keyPairs, errs := aws.GetKeyPairs(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Key Pairs!", 1)
				} else {
					keyPairs.PrintTable()
				}
				return nil
			},
		},
		{
			Name:  "listLaunchConfigurations",
			Usage: "Lists all Launch Configurations",
			Action: func(c *cli.Context) error {
				launchConfigs, errs := aws.GetLaunchConfigurations()
				if errs != nil {
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
				loadBalancers, errs := aws.GetLoadBalancers()
				if errs != nil {
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
				policies, errs := aws.GetScalingPolicies()
				if errs != nil {
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
				groups, errs := aws.GetSecurityGroups()
				if errs != nil {
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
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				snapshots, errs := aws.GetSnapshots(c.NamedArg("search"))
				if errs != nil {
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
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				subnets, errs := aws.GetSubnets(c.NamedArg("search"))
				if errs != nil {
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
				domains, errs := aws.GetSimpleDBDomains(c.NamedArg("search"))
				if errs != nil {
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
				volumes, errs := aws.GetVolumes()
				if errs != nil {
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
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				vpcs, errs := aws.GetVpcs(c.NamedArg("search"))
				if errs != nil {
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
