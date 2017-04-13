# awsm
> AWS Interface

[![Build Status](https://travis-ci.org/murdinc/awsm.svg)](https://travis-ci.org/murdinc/awsm)

#### Note: This tool is not yet complete, but it is nearing initial release.

## Intro
**awsm** is a CLI for building and maintaining your infrastructure on Amazon Web Services. It's designed to streamline many of the tasks involved with setting up and scaling infrastructure across multiple AWS Regions. It's goal is to introduce as few new concepts as possible, and provide powerful tools that require only a few inputs to use.

**[awsmDashboard](https://github.com/murdinc/awsmDashboard)** is a web interface for configuring awsm. The configuration of Classes (see Features) are done though the web interface, and you can also see a list of all of your current AWS services through the Dashboard.

## Features
**Class** (short for classification) is a group of settings for any AWS service, stored in a SimpleDB database by awsm. Classes can be used to bootstrap assets in any AWS region, allowing you to configure once, and run anywhere.

**Propagation** allows you to (optionally) copy/backup assets to other regions when you create them. Currently: EBS Snapshots, AMI Images, and Launch Configurations are available for propagation - allowing you to automatically have access to the latest versions of those as you create them.

**Retention** (also optional) is the number of previous versions of assets to retain. Older EBS Snapshots, AMI's, and Launch Configurations can be rotated out as new ones are created, automating the task of clearing them out. EBS Snapshots and AMI's that are referenced in existing Launch Configurations are never touched.


## Installation
To install awsm, simply copy/paste the following command into your terminal:
```
curl -s http://dl.sudoba.sh/get/awsm | sh
```


## Configuration
The first time you run awsm on a machine, it will ask you to provide an AWS Access ID and Secret Key. Once those are saved, it will create a simpleDB Domain named `awsm` if one does not already exist, and load the default starter awsm classes.

Note: When running awsm on an EC2 instance that was launched with an IAM Instance Profile, you will not need to enter your Key and Secret.


## Commands (CLI)
* dashboard - "Launch the awsm Dashboard GUI"
* attachVolume - "Attach an AWS EBS Volume to an EC2 Instance"
* copyImage - "Copy an AWS Machine Image to another region"
* copySnapshot - "Copy an AWS EBS Snapshot to another region"
* createAddress - "Create an AWS Elastic IP Address"
* createIAMUser - "Create an IAM User"
* createImage - "Create an AWS Machine Image from a running instance"
* createKeyPair - "Create and upload an AWS Key Pair"
* createSimpleDBDomain - "Create an AWS SimpleDB Domain"
* createSnapshot - "Create an AWS EBS snapshot of a volume"
* createVolume - "Create an AWS EBS volume"
* createVpc - "Create an AWS VPC"
* createSubnet - "Create an AWS VPC Subnet"
* deleteAddresses - "Delete AWS Elastic IP Addresses"
* deleteIAMRoles - "Delete AWS IAM Roles"
* deleteIAMUsers - "Delete AWS IAM Users"
* deleteImages - "Delete AWS Machine Images"
* deleteKeyPairs - "Delete AWS KeyPairs"
* deleteSnapshots - "Delete AWS EBS Snapshots"
* deleteSimpleDBDomains - "Delete AWS SimpleDB Domains"
* deleteVolumes - "Delete AWS EBS Volumes"
* deleteSubnets - "Delete AWS VPC Subnets"
* deleteVpcs - "Delete AWS VPCs"
* detachVolume - "Detach an AWS EBS Volume"
* resumeProcesses - "Resume scaling processes on autoscaling groups"
* stopInstances - "Stop AWS instances"
* startInstances - "Start AWS instances"
* suspendProcesses - "Suspend scaling processes on autoscaling groups"
* rebootInstances - "Reboot AWS instances"
* terminateInstances - "Terminate AWS instances"
* launchInstance - "Launch an EC2 instance"
* listAddresses - "Lists AWS Elastic IP Addresses"
* listAlarms - "Lists CloudWatch Alarms"
* listAutoScaleGroups - "Lists AutoScale Groups"
* listIAMInstanceProfiles = "List IAM Instance Profiles"
* listIAMRoles - "Lists IAM Roles"
* listIAMUsers - "Lists IAM Users"
* listImages - "Lists AWS Machine Images owned by us"
* listInstances - "Lists AWS EC2 Instances"
* listKeyPairs - "Lists AWS Key Pairs"
* listLaunchConfigurations - "Lists Launch Configurations"
* listLoadBalancers - "Lists Elastic Load Balancers"
* listScalingPolicies - "Lists Scaling Policies"
* listSecurityGroups - "Lists Security Groups"
* listSnapshots - "Lists AWS EBS Snapshots"
* listSubnets - "Lists AWS Subnets"
* listSimpleDBDomains - "Lists AWS SimpleDB Domains"
* listVolumes - "Lists AWS EBS Volumes"
* listVpcs - "Lists AWS Vpcs"
* createAutoScaleGroup - "Create an AWS AutoScaling Group"
* createLaunchConfiguration - "Create an AWS AutoScaling Launch Configuration"
* deleteAutoScaleGroup - "Delete AWS AutoScaling Groups"
* deleteLaunchConfiguration - "Delete AWS AutoScaling Launch Configurations"
* updateAutoScaleGroup - "Update an AWS AutoScaling Group"

## Roadmap

* Adding support for Application ELBs
* runCommand - "Run a command on a set of instances"
* Config to JSON import and export



Also, check out [awsmDashboard](https://github.com/murdinc/awsmDashboard) which feeds into this project.


