# awsm
> AWS Interface

[![Build Status](https://travis-ci.org/murdinc/awsm.svg)](https://travis-ci.org/murdinc/awsm)

#### Note: This is still in early beta. It is not recommended for use in production environments.

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
* associateRouteTable - "Associate a Route Table to a Subnet"
* attachIAMRolePolicy - "Attach an IAM Policy to a IAM Role"
* attachInternetGateway - "Attach an Internet Gateway to a VPC"
* attachVolume - "Attach an EBS Volume to an EC2 Instance"
* installKeyPair - "Installs a Key Pair locally"
* copyImage - "Copy a Machine Image to another region"
* copySnapshot - "Copy an EBS Snapshot to another region"
* createAddress - "Create an Elastic IP Address"
* createAutoScaleGroups - "Create an AutoScaling Groups"
* createIAMUser - "Create an IAM User"
* createIAMPolicy - "Create an IAM Policy"
* createInternetGateway - "Create an Internet Gateway"
* createImage - "Create a Machine Image from a running instance"
* createLaunchConfigurations - "Create an AutoScaling Launch Configurations"
* createLoadBalancer - "Create a Load Balancer"
* createKeyPair - "Create a Key Pair in the specified region"
* createResourceRecord - "Create a Route53 Resource Record"
* createRouteTable - "Create a Route Table"
* createSecurityGroup - "Create a Security Groups"
* createSimpleDBDomain - "Create a SimpleDB Domain"
* createSnapshot - "Create an EBS snapshot of a volume"
* createVolume - "Create an EBS volume"
* createVpc - "Create a VPC"
* createSubnet - "Create a VPC Subnet"
* deleteAddresses - "Delete Elastic IP Addresses"
* deleteAutoScaleGroups - "Delete AutoScaling Groups"
* deleteIAMInstanceProfiles - "Delete IAM Instance Profiles"
* deleteIAMPolicies - "Delete IAM Policies"
* deleteIAMRoles - "Delete IAM Roles"
* deleteIAMUsers - "Delete IAM Users"
* deleteInternetGateway - "Delete an Internet Gateway"
* deleteImages - "Delete Machine Images"
* deleteKeyPairs - "Delete KeyPairs"
* deleteLaunchConfigurations - "Delete AutoScaling Launch Configurations"
* deleteLoadBalancers - "Delete Load Balancer(s)""
* deleteResourceRecords - "Delete Route53 Resource Records"
* deleteSecurityGroups - "Delete Security Groups"
* deleteSnapshots - "Delete EBS Snapshots"
* deleteSimpleDBDomains - "Delete SimpleDB Domains"
* deleteVolumes - "Delete EBS Volumes"
* deleteSubnets - "Delete VPC Subnets"
* deleteVpcs - "Delete VPCs"
* deregisterInstances - "Deregister Instances from SSM Inventory"
* detachInternetGateway - "Detach an Internet Gateway from a VPC"
* detachVolume - "Detach an EBS Volume"
* disassociateRouteTable - "Disassociate a Route Table from a Subnet"
* getIAMInstanceProfile - "Get an IAM Instance Profile"
* getIAMPolicy - "Get an IAM Policy"
* getIAMUser - "Get an IAM User"
* getInventory - "Get SSM Inventory"
* stopInstances - "Stop instances"
* startInstances - "Start instances"
* rebootInstances - "Reboot instances"
* refreshVolume - "Refreshe an EBS Volume on an EC2 Instance"
* terminateInstances - "Terminate instances"
* launchInstance - "Launch an EC2 instance"
* listAddresses - "List Elastic IP Addresses"
* listAlarms - "List CloudWatch Alarms"
* listAutoScaleGroups - "List AutoScale Groups"
* listBuckets - "List S3 Buckets"
* listCommandInvocations - "List SSM Command Invocations"
* listHostedZones - "List Route53 Hosted Zones"
* listIAMInstanceProfiles - "List IAM Instance Profiles"
* listIAMPolicies - "List IAM Policies"
* listIAMRoles - "List IAM Roles"
* listIAMUsers - "List IAM Users"
* listImages - "List Machine Images owned by us"
* listInstances - "List EC2 Instances
* listInternetGateways - "List VPC Internet Gateways"
* listKeyPairs - "List Key Pairs"
* listLaunchConfigurations - "List Launch Configurations"
* listLoadBalancers - "List Elastic Load Balancers"
* listResourceRecords - "List Route53 Resource Records"
* listRouteTables - "List VPC Internet Gateways"
* listScalingPolicies - "List Scaling Policies"
* listSecurityGroups - "List Security Groups"
* listSnapshots - "List EBS Snapshots"
* listSSMInstances - "List SSM Instances"
* listSubnets - "List Subnets"
* listSimpleDBDomains - "List SimpleDB Domains"
* listVolumes - "List EBS Volumes"
* listVpcs - "List Vpcs"
* resumeProcesses - "Resume scaling processes on Autoscaling Groups"
* runCommand - "Run a command on a set of EC2 Instances"
* suspendProcesses - "Suspend scaling processes on Autoscaling Groups"
* updateAutoScaleGroups - "Update AutoScaling Groups"
* updateLoadBalancers - "Update Load Balancers"
* updateSecurityGroups - "Update Security Groups"
* installAutocomplete - "Install awsm autocomplete"

## Roadmap

* Adding support for Application ELBs
* Config to JSON import and export (partially complete)


Also, check out [awsmDashboard](https://github.com/murdinc/awsmDashboard) which feeds into this project.


