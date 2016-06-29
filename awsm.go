package main

import (
	"os"

	"github.com/murdinc/awsm/aws"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
)

// Main Function
////////////////..........
func main() {

	found := aws.CheckCreds()
	if !found {
		return
	}

	var dryRun bool
	var force bool

	app := cli.NewApp()
	app.Name = "awsm"
	app.Usage = "AWSM Interface"
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
			Usage: "Attach an AWS EBS Volume to and EC2 Instance",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "volume",
					Description: "The volume to attach",
					Optional:    false,
				},
				cli.Argument{
					Name:        "instance",
					Description: "The instance to attach the volume to",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.AttachVolume(c.NamedArg("volume"), c.NamedArg("instance"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "copyImage",
			Usage: "Copy an AWS Machine Image to another region",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The image to copy",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region to copy the image to",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CopyImage(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "copySnapshot",
			Usage: "Copy an AWS EBS Snapshot to another region",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The snapshot to copy",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region to copy the snapshot to",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CopySnapshot(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createAddress",
			Usage: "Create an AWS Elastic IP Address",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "region",
					Description: "The region to create the elastic ip in",
					Optional:    false,
				},
				cli.Argument{
					Name:        "domain",
					Description: "The domain to create the elastic ip in (classic or vpc)",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateAddress(c.NamedArg("region"), c.NamedArg("domain"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
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
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The AMI to create an image of",
					Optional:    false,
				},
				cli.Argument{
					Name:        "class",
					Description: "The class of the new image",
					Optional:    false,
				},
				cli.Argument{
					Name:        "name",
					Description: "The name of the new image",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateImage(c.NamedArg("search"), c.NamedArg("class"), c.NamedArg("name"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
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
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The volume to create a snapshot from",
					Optional:    false,
				},
				cli.Argument{
					Name:        "class",
					Description: "The class of the new snapshot",
					Optional:    false,
				},
				cli.Argument{
					Name:        "name",
					Description: "The name of the new snapshot",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateSnapshot(c.NamedArg("search"), c.NamedArg("class"), c.NamedArg("name"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createVolume",
			Usage: "Create an AWS EBS volume",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "class",
					Description: "The class of the new volume",
					Optional:    false,
				},
				cli.Argument{
					Name:        "name",
					Description: "The name of the new volume",
					Optional:    false,
				},
				cli.Argument{
					Name:        "az",
					Description: "The Availability Zone to create the volume in",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateVolume(c.NamedArg("class"), c.NamedArg("name"), c.NamedArg("az"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
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
			Name:  "deleteAddresses",
			Usage: "Delete AWS Elastic IP Addresses",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The search term for the elastic ip to delete",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region to delete the elastic ip from",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteAddresses(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteAutoScaleGroup",
			Usage: "Delete AWS AutoScaling Groups",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "deleteIAMUsers",
			Usage: "Delete AWS IAM Users",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The search term for iam username",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteIAMUsers(c.NamedArg("search"))
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteImages",
			Usage: "Delete AWS Machine Images",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The search term for images to delete",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the images (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteImages(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteKeyPairs",
			Usage: "Delete AWS KeyPairs",
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
			Usage: "Delete AWS AutoScaling Launch Configurations",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "deleteSnapshots",
			Usage: "Delete AWS EBS Snapshots",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The search term for snapshots to delete",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the snapshots (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteSnapshots(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteSimpleDBDomains",
			Usage: "Delete AWS SimpleDB Domains",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The search term for simpleDB domain to delete",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the DBs (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteSimpleDBDomain(c.NamedArg("search"), c.NamedArg("region"))
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteVolumes",
			Usage: "Delete AWS EBS Volumes",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "volume",
					Description: "The volume to delete",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the volumes (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {

				err := aws.DeleteVolumes(c.NamedArg("volume"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteSubnets",
			Usage: "Delete AWS VPC Subnets",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The search term for Subnets to delete",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the subnets (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteSubnets(c.NamedArg("search"), c.NamedArg("region"), dryRun)
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
					Name:        "search",
					Description: "The search term for VPCs to delete",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the VPCs (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteVpcs(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "detachVolume",
			Usage: "Detach an AWS EBS Volume",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "volume",
					Description: "The volume to detach",
					Optional:    false,
				},
				cli.Argument{
					Name:        "instance",
					Description: "The instance to detach the volume from",
					Optional:    false,
				},
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "force",
					Destination: &force,
					Usage:       "force detach",
				},
			},
			Action: func(c *cli.Context) error {

				err := aws.DetachVolume(c.NamedArg("volume"), c.NamedArg("instance"), force, dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "stopInstances",
			Usage: "Stop AWS instances",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The search term for instance to stop",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.StopInstances(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "startInstances",
			Usage: "Start AWS instances",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The search term for Instance to start",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.StartInstances(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "rebootInstances",
			Usage: "Reboot AWS instances",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The search term for instance to reboot",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.RebootInstances(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "terminateInstances",
			Usage: "Terminate AWS instances",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "name",
					Description: "The search term for instance to terminate",
					Optional:    false,
				},
				cli.Argument{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.TerminateInstances(c.NamedArg("name"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
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
			Usage: "Lists AWS Elastic IP Addresses",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				addresses, errs := aws.GetAddresses(c.NamedArg("search"), false)
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
			Usage: "Lists CloudWatch Alarms",
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
			Usage: "Lists AutoScale Groups",
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
			Usage: "Lists IAM Users",
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
			Usage: "Lists AWS Machine Images owned by us",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				images, errs := aws.GetImages(c.NamedArg("search"), false)
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
			Usage: "Lists AWS EC2 Instances",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				instances, errs := aws.GetInstances(c.NamedArg("search"), false)
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
			Usage: "Lists AWS Key Pairs",
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
			Usage: "Lists Launch Configurations",
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
			Usage: "Lists Elastic Load Balancers",
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
			Usage: "Lists Scaling Policies",
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
			Usage: "Lists Security Groups",
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
			Usage: "Lists AWS EBS Snapshots",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				snapshots, errs := aws.GetSnapshots(c.NamedArg("search"), false)
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
			Usage: "Lists AWS Subnets",
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
			Usage: "Lists AWS SimpleDB Domains",
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
			Usage: "Lists AWS EBS Volumes",
			Arguments: []cli.Argument{
				cli.Argument{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				volumes, errs := aws.GetVolumes(c.NamedArg("search"), false)
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
			Usage: "Lists AWS Vpcs",
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
			Usage: "Resume autoscaling processes on a specific autoscaling group",
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
			Usage: "Stop autoscaling processes on a specific autoscaling group",
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
