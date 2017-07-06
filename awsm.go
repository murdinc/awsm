package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/user"
	"regexp"

	"github.com/murdinc/awsm/api"
	"github.com/murdinc/awsm/aws"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
)

// Main Function
////////////////..........
func main() {

	var dryRun bool
	var force bool
	var double bool  // optional flag when updating an auto-scale group
	var details bool // optional flag when listing command invocations

	app := cli.NewApp()
	app.Name = "awsm"
	app.Usage = "AWS Interface"
	app.Version = "0.1.1"
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
			Name:   "check",
			Usage:  "Check / repair the awsm config",
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				terminal.Information("The awsm config looks good!")
				return nil
			},
		},
		{
			Name:   "api",
			Usage:  "Start the awsm api server",
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				return api.StartAPI(false)
			},
		},
		{
			Name:   "dashboard",
			Usage:  "Launch the awsm Dashboard GUI",
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				return api.StartAPI(true)
			},
		},
		{
			Name:  "associateRouteTable",
			Usage: "Associate a Route Table to a Subnet",
			Arguments: []cli.Argument{
				{
					Name:        "routetable",
					Description: "The route table to attach",
					Optional:    false,
				},
				{
					Name:        "subnet",
					Description: "The subnet to associate the route table with",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.AssociateRouteTable(c.NamedArg("routetable"), c.NamedArg("subnet"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "attachIAMRolePolicy",
			Usage: "Attach an IAM Policy to a IAM Role",
			Arguments: []cli.Argument{
				{
					Name:        "role",
					Description: "The name of the role to attach the policy to",
					Optional:    false,
				},
				{
					Name:        "policy",
					Description: "The name of the policy to attach to the role",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.AttachIAMRolePolicy(c.NamedArg("role"), c.NamedArg("policy"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "attachInternetGateway",
			Usage: "Attach an Internet Gateway to a VPC",
			Arguments: []cli.Argument{
				{
					Name:        "gateway",
					Description: "The internet gateway to attach",
					Optional:    false,
				},
				{
					Name:        "vpc",
					Description: "The vpc to attach the internet gateway to",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.AttachInternetGateway(c.NamedArg("gateway"), c.NamedArg("vpc"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "attachVolume",
			Usage: "Attach an EBS Volume to an EC2 Instance",
			Arguments: []cli.Argument{
				{
					Name:        "volume",
					Description: "The volume to attach",
					Optional:    false,
				},
				{
					Name:        "instance",
					Description: "The instance to attach the volume to",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.AttachVolume(c.NamedArg("volume"), c.NamedArg("instance"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "installKeyPair",
			Usage: "Installs a Key Pair locally",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the key pair",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.InstallKeyPair(c.NamedArg("class"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "copyImage",
			Usage: "Copy a Machine Image to another region",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The image to copy",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to copy the image to",
					Optional:    false,
				},
			},
			Before: setupCheck,
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
			Usage: "Copy an EBS Snapshot to another region",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The snapshot to copy",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to copy the snapshot to",
					Optional:    false,
				},
			},
			Before: setupCheck,
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
			Usage: "Create an Elastic IP Address",
			Arguments: []cli.Argument{
				{
					Name:        "region",
					Description: "The region to create the elastic ip in",
					Optional:    false,
				},
				{
					Name:        "domain",
					Description: "The domain to create the elastic ip in (classic or vpc)",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				_, err := aws.CreateAddress(c.NamedArg("region"), c.NamedArg("domain"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createAutoScaleAlarms",
			Usage: "Create a AutoScaling Alarms",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the autoscaling groups to create",
					Optional:    false,
				},
				{
					Name:        "search",
					Description: "The AutoScale Groups to create this alarm in",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateAutoScaleAlarms(c.NamedArg("class"), c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createAutoScaleGroups",
			Usage: "Create an AutoScaling Groups",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the autoscaling groups to create",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateAutoScaleGroups(c.NamedArg("class"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createIAMUser",
			Usage: "Create an IAM User",
			Arguments: []cli.Argument{
				{
					Name:        "username",
					Description: "The username for this IAM user",
					Optional:    false,
				},
				{
					Name:        "path",
					Description: "The optional path for the user",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateIAMUser(c.NamedArg("username"), c.NamedArg("path"))
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createIAMPolicy",
			Usage: "Create an IAM Policy",
			Arguments: []cli.Argument{
				{
					Name:        "name",
					Description: "The name for this IAM policy",
					Optional:    false,
				},
				{
					Name:        "document",
					Description: "The document file for this IAM policy",
					Optional:    false,
				},
				{
					Name:        "path",
					Description: "The optional path for this IAM policy",
					Optional:    true,
				},
				{
					Name:        "description",
					Description: "The optional description for this IAM policy",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {

				doc, err := ioutil.ReadFile(c.NamedArg("document"))
				if err != nil {
					terminal.ErrorLine(err.Error())
					return err
				}

				_, err = aws.CreateIAMPolicy(c.NamedArg("name"), string(doc), c.NamedArg("path"), c.NamedArg("description"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createInternetGateway",
			Usage: "Create an Internet Gateway",
			Arguments: []cli.Argument{
				{
					Name:        "name",
					Description: "The name of the internet gateway",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to create the internet gateway in",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				_, err := aws.CreateInternetGateway(c.NamedArg("name"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createImage",
			Usage: "Create a Machine Image from a running instance",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the new image",
					Optional:    false,
				},
				{
					Name:        "search",
					Description: "The instance to create the image from (optional, defaults setting in class configuration)",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateImage(c.NamedArg("class"), c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createLaunchConfigurations",
			Usage: "Create an AutoScaling Launch Configurations",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the launch configuration groups to create",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateLaunchConfigurations(c.NamedArg("class"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createLoadBalancer",
			Usage: "Create a Load Balancer",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the load balancer groups to create",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to create the load balancer in",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateLoadBalancer(c.NamedArg("class"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createKeyPair",
			Usage: "Create a Key Pair in the specified region",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the key pair",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to create the keypair in",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateKeyPair(c.NamedArg("class"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createResourceRecord",
			Usage: "Create a Route53 Resource Record",
			Arguments: []cli.Argument{
				{
					Name:        "record",
					Description: "The record to create (www.stage1.example.com)",
					Optional:    false,
				},
				{
					Name:        "value",
					Description: "The value of the resource record (defaults to instance IP)",
					Optional:    true,
				},
				{
					Name:        "ttl",
					Description: "The ttl of the resource record (defaults to 300)",
					Optional:    true,
				},
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "force",
					Destination: &force,
					Usage:       "force (UPSERT, no prompt)",
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateResourceRecord(c.NamedArg("record"), c.NamedArg("value"), c.NamedArg("ttl"), force, dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createRouteTable",
			Usage: "Create a Route Table",
			Arguments: []cli.Argument{
				{
					Name:        "name",
					Description: "The name of the route table",
					Optional:    false,
				},
				{
					Name:        "vpc",
					Description: "The vpc to create the route table in",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				_, err := aws.CreateRouteTable(c.NamedArg("name"), c.NamedArg("vpc"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createScalingPolicies",
			Usage: "Create Scaling Policies",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of security group to create",
					Optional:    false,
				},
				{
					Name:        "search",
					Description: "The search term for the autoscaling groups to create scaling policies in",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateScalingPolicy(c.NamedArg("class"), c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createSecurityGroup",
			Usage: "Create a Security Groups",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of security group to create",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to create the security group in",
					Optional:    false,
				},
				{
					Name:        "vpc",
					Description: "The vpc to create the security group in (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateSecurityGroup(c.NamedArg("class"), c.NamedArg("region"), c.NamedArg("vpc"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createSimpleDBDomain",
			Usage: "Create a SimpleDB Domain",
			Arguments: []cli.Argument{
				{
					Name:        "domain",
					Description: "The domain of the db",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the db",
					Optional:    false,
				},
			},
			Before: setupCheck,
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
			Usage: "Create an EBS snapshot of a volume",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the new snapshot",
					Optional:    false,
				},
				{
					Name:        "search",
					Description: "The volume to create the snapshot from (optional, defaults setting in class configuration)",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.CreateSnapshot(c.NamedArg("class"), c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createVolume",
			Usage: "Create an EBS volume",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the new volume",
					Optional:    false,
				},
				{
					Name:        "name",
					Description: "The name of the new volume",
					Optional:    false,
				},
				{
					Name:        "az",
					Description: "The Availability Zone to create the volume in",
					Optional:    false,
				},
			},
			Before: setupCheck,
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
			Usage: "Create a VPC",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of VPC to create",
					Optional:    false,
				},
				{
					Name:        "name",
					Description: "The name of the VPC",
					Optional:    false,
				},
				{
					Name:        "ip",
					Description: "The IP address of this VPC (not including CIDR)",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to create the VPC in",
					Optional:    false,
				},
			},
			Before: setupCheck,
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
			Usage: "Create a VPC Subnet",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of Subnet to create",
					Optional:    false,
				},
				{
					Name:        "name",
					Description: "The name of the Subnet",
					Optional:    false,
				},
				{
					Name:        "vpc",
					Description: "The VPC to create the Subnet in",
					Optional:    false,
				},
				{
					Name:        "ip",
					Description: "The IP address of this Subnet (not including CIDR)",
					Optional:    false,
				},
				{
					Name:        "az",
					Description: "The Availability Zone to create the Subnet in",
					Optional:    true,
				},
			},
			Before: setupCheck,
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
			Usage: "Delete Elastic IP Addresses",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the elastic ip to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to delete the elastic ip from",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteAddresses(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteAutoScaleGroups",
			Usage: "Delete AutoScaling Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the autoscaling group to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to delete the autoscaling group from",
					Optional:    true,
				},
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "force",
					Destination: &force,
					Usage:       "force (Force deletes all instances and lifecycle actions)",
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteAutoScaleGroups(c.NamedArg("search"), c.NamedArg("region"), force, dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteIAMInstanceProfiles",
			Usage: "Delete IAM Instance Profiles",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for iam instance profiles",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteIAMInstanceProfiles(c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteIAMPolicies",
			Usage: "Delete IAM Policies",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for iam policy",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteIAMPolicies(c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteIAMRoles",
			Usage: "Delete IAM Roles",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for iam role",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteIAMRoles(c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteIAMUsers",
			Usage: "Delete IAM Users",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for iam username",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteIAMUsers(c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteInternetGateway",
			Usage: "Delete an Internet Gateway",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for internet gateway",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteInternetGateway(c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteImages",
			Usage: "Delete Machine Images",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for images to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the images (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
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
			Usage: "Delete KeyPairs",
			Arguments: []cli.Argument{
				{
					Name:        "name",
					Description: "The name of the KeyPair to delete",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				errs := aws.DeleteKeyPairs(c.NamedArg("name"), dryRun)
				if errs != nil {
					return cli.NewExitError("Errors Deleting KeyPair!", 1)

				}
				return nil
			},
		},
		{
			Name:  "deleteLaunchConfigurations",
			Usage: "Delete AutoScaling Launch Configurations",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the launch configuration to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to delete the launch configuration from",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteLaunchConfigurations(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteLoadBalancers",
			Usage: "Delete Load Balancer(s)",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the load balancer to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to create the load balancer in",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteLoadBalancers(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteResourceRecords",
			Usage: "Delete Route53 Resource Records",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the resource record to delete",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteResourceRecords(c.NamedArg("search"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteSecurityGroups",
			Usage: "Delete Security Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the security group to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to delete the security group from",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteSecurityGroups(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteSnapshots",
			Usage: "Delete EBS Snapshots",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for snapshots to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the snapshots (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
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
			Usage: "Delete SimpleDB Domains",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for simpleDB domain to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the DBs (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteSimpleDBDomains(c.NamedArg("search"), c.NamedArg("region"))
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteVolumes",
			Usage: "Delete EBS Volumes",
			Arguments: []cli.Argument{
				{
					Name:        "volume",
					Description: "The volume to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the volumes (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
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
			Usage: "Delete VPC Subnets",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for Subnets to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the subnets (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
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
			Usage: "Delete VPCs",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for VPCs to delete",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the VPCs (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeleteVpcs(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deregisterInstances",
			Usage: "Deregister Instances from SSM Inventory",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The inventory to search for",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DeregisterInstances(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "detachInternetGateway",
			Usage: "Detach an Internet Gateway from a VPC",
			Arguments: []cli.Argument{
				{
					Name:        "gateway",
					Description: "The internet gateway to detach",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DetachInternetGateway(c.NamedArg("gateway"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "detachVolume",
			Usage: "Detach an EBS Volume",
			Arguments: []cli.Argument{
				{
					Name:        "volume",
					Description: "The volume to detach",
					Optional:    false,
				},
				{
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
			Before: setupCheck,
			Action: func(c *cli.Context) error {

				err := aws.DetachVolume(c.NamedArg("volume"), c.NamedArg("instance"), force, dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "disassociateRouteTable",
			Usage: "Disassociate a Route Table from a Subnet",
			Arguments: []cli.Argument{
				{
					Name:        "routetable",
					Description: "The route table to disassociate",
					Optional:    false,
				},
				{
					Name:        "subnet",
					Description: "The subnet to disassociate the route table from",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.DisassociateRouteTable(c.NamedArg("routetable"), c.NamedArg("subnet"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "getIAMInstanceProfile",
			Usage: "Get an IAM Instance Profile",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the iam instance profile",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				instanceProfile, err := aws.GetIAMInstanceProfile(c.NamedArg("search"))

				if err != nil {
					terminal.ErrorLine(err.Error())
				}

				instProfilesSlice := aws.IAMInstanceProfiles{instanceProfile}
				instProfilesSlice.PrintTable()

				return nil
			},
		},
		{
			Name:  "getIAMPolicy",
			Usage: "Get an IAM Policy",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the iam policy document",
					Optional:    false,
				},
				{
					Name:        "version",
					Description: "The version of the iam policy document to retrieve",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				policyDocument, err := aws.GetIAMPolicyDocument(c.NamedArg("search"), c.NamedArg("version"))

				if err != nil {
					terminal.ErrorLine(err.Error())
				}

				// TODO specify output file and write to that instead?
				fmt.Println(policyDocument.Document)

				return nil
			},
		},
		{
			Name:  "getIAMUser",
			Usage: "Get an IAM User",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The username to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				iam, err := aws.GetIAMUser(c.NamedArg("search"))
				if err != nil {
					terminal.ErrorLine(err.Error())
					return err
				}

				userSlice := aws.IAMUsers{iam}
				userSlice.PrintTable()

				return nil
			},
		},
		{
			Name:  "getInventory",
			Usage: "Get SSM Inventory",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The inventory to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				inventory, errs := aws.GetInventory(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Inventory!", 1)
				}

				inventory.PrintTable()
				return nil
			},
		},
		{
			Name:  "stopInstances",
			Usage: "Stop instances",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for instance to stop",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "force",
					Destination: &force,
					Usage:       "force (Force deletes all instances and lifecycle actions)",
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.StopInstances(c.NamedArg("search"), c.NamedArg("region"), force, dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "startInstances",
			Usage: "Start instances",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for Instance to start",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
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
			Usage: "Reboot instances",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for instance to reboot",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.RebootInstances(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "refreshVolume",
			Usage: "Refreshe an EBS Volume on an EC2 Instance",
			Arguments: []cli.Argument{
				{
					Name:        "volume",
					Description: "The volume to refresh",
					Optional:    false,
				},
				{
					Name:        "instance",
					Description: "The instance to refresh the volume on",
					Optional:    false,
				},
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "force",
					Destination: &force,
					Usage:       "force refresh",
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.RefreshVolume(c.NamedArg("volume"), c.NamedArg("instance"), force, dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "terminateInstances",
			Usage: "Terminate instances",
			Arguments: []cli.Argument{
				{
					Name:        "name",
					Description: "The search term for instance to terminate",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region of the instance (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
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
				{
					Name:        "class",
					Description: "The class of the instance (dev, stage, etc)",
					Optional:    false,
				},
				{
					Name:        "sequence",
					Description: "The sequence of the instance (1...100)",
					Optional:    false,
				},
				{
					Name:        "az",
					Description: "The availability zone to launch the instance in (us-west-2a, us-east-1a, etc)",
					Optional:    false,
				},
			},
			Before: setupCheck,
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
			Usage: "List Elastic IP Addresses",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				addresses, errs := aws.GetAddresses(c.NamedArg("search"), false)
				if errs != nil {
					return cli.NewExitError("Error Listing Addresses!", 1)
				}
				addresses.PrintTable()
				return nil
			},
		},
		{
			Name:  "listAlarms",
			Usage: "List CloudWatch Alarms",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				alarms, errs := aws.GetAlarms(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Alarms!", 1)
				}
				alarms.PrintTable()

				return nil
			},
		},
		{
			Name:  "listAutoScaleGroups",
			Usage: "List AutoScale Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				groups, errs := aws.GetAutoScaleGroups(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Auto Scale Groups!", 1)
				}
				groups.PrintTable()

				return nil
			},
		},
		{
			Name:  "listBuckets",
			Usage: "List S3 Buckets",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				groups, errs := aws.GetBuckets(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing S3 Buckets!", 1)
				}
				groups.PrintTable()

				return nil
			},
		},
		{
			Name:  "listCommandInvocations",
			Usage: "List SSM Command Invocations",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "details",
					Destination: &details,
					Usage:       "details (Shows the output of each command)",
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				commandInvocations, errs := aws.ListCommandInvocations(c.NamedArg("search"), details)
				if errs != nil {
					return cli.NewExitError("Error Listing Command Invocations!", 1)
				}
				if details {
					commandInvocations.PrintOutput()
				} else {
					commandInvocations.PrintTable()
				}

				return nil
			},
		},
		{
			Name:  "listHostedZones",
			Usage: "List Route53 Hosted Zones",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				hostedZones, errs := aws.GetHostedZones(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Hosted Zones!", 1)
				}
				hostedZones.PrintTable()

				return nil
			},
		},

		{
			Name:  "listIAMInstanceProfiles",
			Usage: "List IAM Instance Profiles",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				iam, errs := aws.GetIAMInstanceProfiles(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing IAM Instance Profiles!", 1)
				}
				iam.PrintTable()

				return nil
			},
		},
		{
			Name:  "listIAMPolicies",
			Usage: "List IAM Policies",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				iam, errs := aws.GetIAMPolicies(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing IAM Policies!", 1)
				}
				iam.PrintTable()

				return nil
			},
		},
		{
			Name:  "listIAMRoles",
			Usage: "List IAM Roles",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				iam, errs := aws.GetIAMRoles(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing IAM Roles!", 1)
				}
				iam.PrintTable()

				return nil
			},
		},
		{
			Name:  "listIAMUsers",
			Usage: "List IAM Users",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				iam, errs := aws.GetIAMUsers(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing IAM Users!", 1)
				}
				iam.PrintTable()

				return nil
			},
		},
		{
			Name:  "listImages",
			Usage: "List Machine Images owned by us",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				images, errs := aws.GetImages(c.NamedArg("search"), false)
				if errs != nil {
					return cli.NewExitError("Error Listing Images!", 1)
				}
				images.PrintTable()

				return nil
			},
		},
		{
			Name:  "listInstances",
			Usage: "List EC2 Instances",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				instances, errs := aws.GetInstances(c.NamedArg("search"), false)
				if errs != nil {
					return cli.NewExitError("Error Listing Instances!", 1)
				}
				instances.PrintTable()

				return nil
			},
		},
		{
			Name:  "listInternetGateways",
			Usage: "List VPC Internet Gateways",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				internetGateways, errs := aws.GetInternetGateways(c.NamedArg("search"), false)
				if errs != nil {
					return cli.NewExitError("Error Listing Internet Gateways!", 1)
				}
				internetGateways.PrintTable()

				return nil
			},
		},
		{
			Name:  "listKeyPairs",
			Usage: "List Key Pairs",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				keyPairs, errs := aws.GetKeyPairs(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Key Pairs!", 1)
				}
				keyPairs.PrintTable()

				return nil
			},
		},
		{
			Name:  "listLaunchConfigurations",
			Usage: "List Launch Configurations",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				launchConfigs, errs := aws.GetLaunchConfigurations(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Launch Configurations!", 1)
				}
				launchConfigs.PrintTable()

				return nil
			},
		},
		{
			Name:  "listLoadBalancers",
			Usage: "List Elastic Load Balancers",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				loadBalancers, errs := aws.GetLoadBalancers(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Load Balancers!", 1)
				}
				loadBalancers.PrintTable()

				return nil
			},
		},
		{
			Name:  "listResourceRecords",
			Usage: "List Route53 Resource Records",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				resourceRecords, errs := aws.GetResourceRecords(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Resource Records!", 1)
				}
				resourceRecords.PrintTable()

				return nil
			},
		},
		{
			Name:  "listRouteTables",
			Usage: "List VPC Internet Gateways",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				internetGateways, errs := aws.GetRouteTables(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Route Tables!", 1)
				}
				internetGateways.PrintTable()

				return nil
			},
		},
		{
			Name:  "listScalingPolicies",
			Usage: "List Scaling Policies",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				policies, errs := aws.GetScalingPolicies(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Auto Scaling Policies!", 1)
				}
				policies.PrintTable()

				return nil
			},
		},
		{
			Name:  "listSecurityGroups",
			Usage: "List Security Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				groups, errs := aws.GetSecurityGroups(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Security Groups!", 1)
				}
				groups.PrintTable()

				return nil
			},
		},
		{
			Name:  "listSnapshots",
			Usage: "List EBS Snapshots",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				snapshots, errs := aws.GetSnapshots(c.NamedArg("search"), false)
				if errs != nil {
					return cli.NewExitError("Error Listing Snapshots!", 1)
				}
				snapshots.PrintTable()

				return nil
			},
		},
		{
			Name:  "listSSMInstances",
			Usage: "List SSM Instances",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				instances, errs := aws.GetSSMInstances(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing SSM Instances!", 1)
				}
				instances.PrintTable()

				return nil
			},
		},
		{
			Name:  "listSubnets",
			Usage: "List Subnets",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				subnets, errs := aws.GetSubnets(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Subnets!", 1)
				}
				subnets.PrintTable()

				return nil
			},
		},
		{
			Name:  "listSimpleDBDomains",
			Usage: "List SimpleDB Domains",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				domains, errs := aws.GetSimpleDBDomains(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing Simple DB Domains!", 1)
				}
				domains.PrintTable()

				return nil
			},
		},
		{
			Name:  "listVolumes",
			Usage: "List EBS Volumes",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				volumes, errs := aws.GetVolumes(c.NamedArg("search"), false)
				if errs != nil {
					return cli.NewExitError("Error Listing Volumes!", 1)
				}
				volumes.PrintTable()

				return nil
			},
		},
		{
			Name:  "listVpcs",
			Usage: "List Vpcs",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				vpcs, errs := aws.GetVpcs(c.NamedArg("search"))
				if errs != nil {
					return cli.NewExitError("Error Listing VPCs!", 1)
				}
				vpcs.PrintTable()

				return nil
			},
		},
		{
			Name:  "resumeProcesses",
			Usage: "Resume scaling processes on Autoscaling Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the autoscaling group to resume",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to resume the processes in",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.ResumeProcesses(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "runCommand",
			Usage: "Run a command on a set of EC2 Instances",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the instances to run the command on",
					Optional:    false,
				},
				{
					Name:        "command",
					Description: "The command to run on the instances",
					Optional:    false,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				cmdInvocations, err := aws.RunCommand(c.NamedArg("search"), c.NamedArg("command"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				} else {
					cmdInvocations.PrintOutput()
				}
				return nil
			},
		},
		{
			Name:  "suspendProcesses",
			Usage: "Suspend scaling processes on Autoscaling Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for the autoscaling group to suspend",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to suspend the processes in",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.SuspendProcesses(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "updateAutoScaleGroups",
			Usage: "Update AutoScaling Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term of the autoscaling group to update",
					Optional:    false,
				},
				{
					Name:        "version",
					Description: "The version of the launch configuration group to use (defaults to the most recent)",
					Optional:    true,
				},
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "double",
					Destination: &double,
					Usage:       "double (Doubles the desired-capacity and max-capacity)",
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.UpdateAutoScaleGroups(c.NamedArg("search"), c.NamedArg("version"), double, dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "updateLoadBalancers",
			Usage: "Update Load Balancers",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term of the load balancers to update",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to update the load balancers in (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.UpdateLoadBalancers(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "updateSecurityGroups",
			Usage: "Update Security Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term of the security groups to update",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to update the security groups in (optional)",
					Optional:    true,
				},
			},
			Before: setupCheck,
			Action: func(c *cli.Context) error {
				err := aws.UpdateSecurityGroups(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "installAutocomplete",
			Usage: "Install awsm autocomplete",
			Action: func(c *cli.Context) error {
				err := installAutocomplete()
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}

func installAutocomplete() error {

	currentUser, _ := user.Current()
	bashrcPath := currentUser.HomeDir + "/.bashrc"

	autocompleteScriptPath := currentUser.HomeDir + "/.awsm_autocomplete"
	autcompleteLine := "PROG=awsm source " + autocompleteScriptPath

	// Check for the autocomplete script
	if _, err := os.Stat(autocompleteScriptPath); os.IsNotExist(err) {
		terminal.Delta("autocomplete script [" + autocompleteScriptPath + "] does not exist, creating one...")

		f, err := os.Create(autocompleteScriptPath)
		defer f.Close()
		if err != nil {
			return err
		}
		f.WriteString(autocompleteScript)
		f.Sync()
	} else {
		terminal.Information("autocomplete script [" + autocompleteScriptPath + "] exists, checking if .bashrc file exists...")
	}

	// check for .bashrc file
	if _, err := os.Stat(bashrcPath); os.IsNotExist(err) {
		terminal.Delta(".bashrc file [" + bashrcPath + "] does not exist, creating one...")
		f, err := os.Create(bashrcPath)
		defer f.Close()
		if err != nil {
			return err
		}
		f.Sync()
	}

	terminal.Information(".bashrc file [" + bashrcPath + "] exists, checking if autocomplete has been set up already...")

	bashrcFile, err := ioutil.ReadFile(bashrcPath)
	if err != nil {
		terminal.ErrorLine(err.Error())
	}

	term := regexp.MustCompile("\n" + autcompleteLine)
	exists := term.MatchString(string(bashrcFile))

	if exists {
		terminal.Information(".bashrc file already contains the path the the autocomplete script")
	} else {
		terminal.Delta("adding autocomplete line to your .bashrc...")
		f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		defer f.Close()
		if err != nil {
			return err
		}
		_, err = f.WriteString(autcompleteLine)
		if err != nil {
			return err
		}
	}

	terminal.Information("Done!")

	return nil
}

func setupCheck(c *cli.Context) error {

	// Creds Check
	found, accountId := aws.CheckCreds()
	if !found {
		terminal.ErrorLine("No Credentials Available, Aborting!")
		os.Exit(0)
	}

	// DB Check
	if !config.CheckDB() {
		create := terminal.BoxPromptBool("No awsm database found!", "Do you want to create one now?")
		if !create {
			terminal.Information("Ok, maybe next time.. ")
			os.Exit(0)
		}

		// Check if we already have an awsm KeyPair on aws already or not
		generateAwsmKeyPair := true
		keyPairs, _ := aws.GetKeyPairs("awsm")
		if len(*keyPairs) > 0 {
			generateAwsmKeyPair = false
		}

		// Create the SimpleDB Domain
		err := config.CreateAwsmDatabase(generateAwsmKeyPair)
		if err != nil {
			return err
		}

		var policyDocument string
		region := "us-east-1" // TODO handle default region preference
		dbArn := "arn:aws:sdb:" + region + ":" + accountId + ":domain/awsm"

		t := template.New("")
		t, err = t.Parse(awsmDBPolicy)
		if err == nil {
			buff := bytes.NewBufferString("")
			t.Execute(buff, dbArn)
			policyDocument = buff.String()
		}

		// Create the awsm-db IAM Policy granting access to the newly created awsm simpledb domain
		policyARN, err := aws.CreateIAMPolicy("awsm-db", policyDocument, "", "", false)
		if err != nil {
			return err
		}

		// Create the awsm IAM Role
		_, err = aws.CreateIAMRole("awsm", awsmRolePolicyDocument, "", false)
		if err != nil {
			return err
		}

		// Create the awsm IAM Instance Profile Role
		_, err = aws.CreateIAMInstanceProfile("awsm", "", false)
		if err != nil {
			return err
		}

		// Attach the awsm IAM Role to the awsm IAM Instance Profile
		err = aws.AddIAMRoleToInstanceProfile("awsm", "awsm", false)
		if err != nil {
			return err
		}

		// Attach the awsm-db IAM Policy to the awsm IAM Instance Profile Role
		err = aws.AttachIAMRolePolicyByARN("awsm", policyARN, false)
		if err != nil {
			return err
		}

		// TODO attach admin policy to awsm also?
	}

	return nil
}

var awsmDBPolicy = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sdb:*"
            ],
            "Resource": [
                "{{.}}"
            ]
        }
    ]
}`

var awsmRolePolicyDocument = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

var autocompleteScript = `#! /bin/bash

: ${PROG:=$(basename ${BASH_SOURCE})}

_cli_bash_autocomplete() {
     local cur opts base
     COMPREPLY=()
     cur="${COMP_WORDS[COMP_CWORD]}"
     opts=$( ${COMP_WORDS[@]:0:$COMP_CWORD} --generate-bash-completion )
     COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
     return 0
 }

 complete -F _cli_bash_autocomplete $PROG`
