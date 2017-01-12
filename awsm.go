package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/murdinc/awsm/api"
	"github.com/murdinc/awsm/aws"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
)

// Main Function
////////////////..........
func main() {

	// Setup Check
	err := setupCheck()
	if err != nil {
		terminal.ErrorLine(err.Error())
		return
	}

	var dryRun bool
	var force bool
	var double bool // optional flag when updating an auto-scale group

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
			Name:  "api",
			Usage: "Start the awsm api server",
			Action: func(c *cli.Context) error {
				api.StartAPI()
				return nil
			},
		},
		/*
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
		*/
		{
			Name:  "attachIAMRolePolicy",
			Usage: "Attach an AWS IAM Policy to a IAM Role",
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
			Action: func(c *cli.Context) error {
				err := aws.AttachIAMRolePolicy(c.NamedArg("role"), c.NamedArg("policy"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "attachVolume",
			Usage: "Attach an AWS EBS Volume to an EC2 Instance",
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
			Usage: "Copy an AWS Machine Image to another region",
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
			Action: func(c *cli.Context) error {
				err := aws.CreateAddress(c.NamedArg("region"), c.NamedArg("domain"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createAutoScaleGroups",
			Usage: "Create an AWS AutoScaling Groups",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the autoscaling groups to create",
					Optional:    false,
				},
			},
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
			Name:  "createImage",
			Usage: "Create an AWS Machine Image from a running instance",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the new image",
					Optional:    false,
				},
				{
					Name:        "name",
					Description: "The name of the new image",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateImage(c.NamedArg("class"), c.NamedArg("name"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createLaunchConfigurations",
			Usage: "Create an AWS AutoScaling Launch Configurations",
			Arguments: []cli.Argument{
				{
					Name:        "class",
					Description: "The class of the launch configuration groups to create",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateLaunchConfigurations(c.NamedArg("class"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createKeyPair",
			Usage: "Create an AWS Key Pair in the specified region",
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
			Action: func(c *cli.Context) error {
				err := aws.CreateKeyPair(c.NamedArg("class"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "createSecurityGroup",
			Usage: "Create an AWS Security Groups",
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
			Usage: "Create an AWS SimpleDB Domain",
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
				{
					Name:        "class",
					Description: "The class of the new snapshot",
					Optional:    false,
				},
				{
					Name:        "name",
					Description: "The name of the new snapshot",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.CreateSnapshot(c.NamedArg("class"), c.NamedArg("name"), dryRun)
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
			Usage: "Delete AWS AutoScaling Groups",
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
			Usage: "Delete AWS IAM Instance Profiles",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for iam instance profiles",
					Optional:    false,
				},
			},
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
			Usage: "Delete AWS IAM Policies",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for iam policy",
					Optional:    false,
				},
			},
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
			Usage: "Delete AWS IAM Roles",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for iam role",
					Optional:    false,
				},
			},
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
			Usage: "Delete AWS IAM Users",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The search term for iam username",
					Optional:    false,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.DeleteIAMUsers(c.NamedArg("search"), dryRun)
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
				{
					Name:        "name",
					Description: "The name of the AWS KeyPair to delete",
					Optional:    false,
				},
			},
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
			Usage: "Delete AWS AutoScaling Launch Configurations",
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
			Action: func(c *cli.Context) error {
				err := aws.DeleteLaunchConfigurations(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
		{
			Name:  "deleteSecurityGroups",
			Usage: "Delete AWS Security Groups",
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
			Usage: "Delete AWS EBS Snapshots",
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
			Usage: "Delete AWS EBS Volumes",
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
			Action: func(c *cli.Context) error {

				err := aws.DetachVolume(c.NamedArg("volume"), c.NamedArg("instance"), force, dryRun)
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
					Name:        "username",
					Description: "The username to search for",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				iam, err := aws.GetIAMUser(c.NamedArg("username"))
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
			Name:  "stopInstances",
			Usage: "Stop AWS instances",
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
			Usage: "Start AWS instances",
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
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists CloudWatch Alarms",
			Action: func(c *cli.Context) error {
				alarms, errs := aws.GetAlarms()
				if errs != nil {
					return cli.NewExitError("Error Listing Alarms!", 1)
				}
				alarms.PrintTable()

				return nil
			},
		},
		{
			Name:  "listAutoScaleGroups",
			Usage: "Lists AutoScale Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Name:  "listIAMInstanceProfiles",
			Usage: "Lists IAM Instance Profiles",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists IAM Policies",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists IAM Roles",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists IAM Users",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists AWS Machine Images owned by us",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists AWS EC2 Instances",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Name:  "listKeyPairs",
			Usage: "Lists AWS Key Pairs",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists Launch Configurations",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists Elastic Load Balancers",
			Action: func(c *cli.Context) error {
				loadBalancers, errs := aws.GetLoadBalancers()
				if errs != nil {
					return cli.NewExitError("Error Listing Load Balancers!", 1)
				}
				loadBalancers.PrintTable()

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
				}
				policies.PrintTable()

				return nil
			},
		},
		{
			Name:  "listSecurityGroups",
			Usage: "Lists Security Groups",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists AWS EBS Snapshots",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Name:  "listSubnets",
			Usage: "Lists AWS Subnets",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists AWS SimpleDB Domains",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists AWS EBS Volumes",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Lists AWS Vpcs",
			Arguments: []cli.Argument{
				{
					Name:        "search",
					Description: "The keyword to search for",
					Optional:    true,
				},
			},
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
			Usage: "Resume scaling processes on autoscaling groups",
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
			Usage: "Run a command on a set of instances",
			Action: func(c *cli.Context) error {
				// TODO
				return nil
			},
		},
		{
			Name:  "suspendProcesses",
			Usage: "Suspend scaling processes on autoscaling groups",
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
					Description: "The search term autoscaling group to update",
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
			Action: func(c *cli.Context) error {
				err := aws.UpdateAutoScaleGroups(c.NamedArg("search"), c.NamedArg("version"), double, dryRun)
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
					Description: "The search term autoscaling group to update",
					Optional:    false,
				},
				{
					Name:        "region",
					Description: "The region to update the security groups in (optional)",
					Optional:    true,
				},
			},
			Action: func(c *cli.Context) error {
				err := aws.UpdateSecurityGroups(c.NamedArg("search"), c.NamedArg("region"), dryRun)
				if err != nil {
					terminal.ErrorLine(err.Error())
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}

func setupCheck() error {
	if !config.CheckDB() {
		create := terminal.BoxPromptBool("No awsm database found!", "Do you want to create one now?")
		if !create {
			terminal.Information("Ok then, maybe next time.. ")
			os.Exit(0)
		}

		// Get the current user
		iamUser, err := aws.GetIAMUser("")
		if err != nil {
			return err
		}

		// Parse the ARN to get the account ID
		parsedArn, err := aws.ParseArn(iamUser.Arn)
		if err != nil {
			return err
		}

		// Create the SimpleDB Domain
		err = config.CreateAwsmDatabase()
		if err != nil {
			return err
		}

		var policyDocument string
		region := "us-east-1" // TODO handle default region preference
		accountId := parsedArn.AccountID
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
