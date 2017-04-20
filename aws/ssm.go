package aws

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	humanize "github.com/dustin/go-humanize"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

var ssmRegions = []string{
	"us-east-1", "us-east-2", "us-west-1", "us-west-2",
	"ap-southeast-1", "ap-southeast-2", "ap-northeast-1",
	"ap-northeast-2", "eu-central-1", "eu-west-1", "sa-east-1",
}

// Inventory represents a slice of SSM Entities
type Inventory []Entity

// Entity represents a single SSM Entity
type Entity models.Entity

// SSMInstances represents a slice of SSM Instances
type SSMInstances []SSMInstance

// SSMInstance represents a single SSM Instance
type SSMInstance models.SSMInstance

// CommandInvocations represents a slice of SSM Command Invocations
type CommandInvocations []CommandInvocation

// CommandInvocation represents a single SSM Command Invocation
type CommandInvocation models.CommandInvocation

// CommandPlugins represents a slice of SSM Command Plugins
type CommandPlugins []CommandPlugin

// CommandPlugin represents a single SSM Command Plugin
type CommandPlugin models.CommandPlugin

// GetSSMInstanceById returns a single SSM Instance by its instance ID and region
func GetSSMInstanceById(region, id string) (SSMInstance, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ssm.New(sess)

	params := &ssm.DescribeInstanceInformationInput{
		InstanceInformationFilterList: []*ssm.InstanceInformationFilter{
			{
				Key: aws.String("InstanceIds"),
				ValueSet: []*string{
					aws.String(id),
				},
			},
		},
	}

	result, err := svc.DescribeInstanceInformation(params)
	if err != nil {
		return SSMInstance{}, err
	}

	if len(result.InstanceInformationList) == 0 {
		return SSMInstance{}, errors.New("No SSM Instances found matching instance ID [" + id + "] in [" + region + "].")
	}

	instList := new(Instances)
	GetRegionInstances(region, instList, "", true)

	instances := make(SSMInstances, len(result.InstanceInformationList))
	for i, inst := range result.InstanceInformationList {
		instances[i].Marshal(inst, region, instList)
	}

	return instances[0], err
}

// GetSSMInstances returns a slice of SSMInstances that math the provided optional search term
func GetSSMInstances(search string) (*SSMInstances, []error) {
	var wg sync.WaitGroup
	var errs []error

	ssmInstList := new(SSMInstances)

	for _, region := range ssmRegions {
		wg.Add(1)

		go func(region string) {
			defer wg.Done()
			err := GetRegionSSMInstances(region, ssmInstList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering instance list for region [%s]", region), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}

	wg.Wait()

	return ssmInstList, errs
}

// GetRegionSSMInstances returns a slice of Instances into the passed Instances slice based on the provided region and search term, and optional running flag
func GetRegionSSMInstances(region string, ssmInstList *SSMInstances, search string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ssm.New(sess)

	result, err := svc.DescribeInstanceInformation(&ssm.DescribeInstanceInformationInput{})
	if err != nil {
		return err
	}

	instList := new(Instances)
	GetRegionInstances(region, instList, "", true)

	instances := make(SSMInstances, len(result.InstanceInformationList))
	for i, inst := range result.InstanceInformationList {
		instances[i].Marshal(inst, region, instList)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, in := range instances {
			rInst := reflect.ValueOf(in)

			for k := 0; k < rInst.NumField(); k++ {
				sVal := rInst.Field(k).String()

				if term.MatchString(sVal) {
					*ssmInstList = append(*ssmInstList, instances[i])
					continue Loop
				}
			}
		}
	} else {
		*ssmInstList = append(*ssmInstList, instances[:]...)
	}

	return nil
}

func GetInventory(search string) (*Inventory, []error) {
	var wg sync.WaitGroup
	var errs []error

	invList := new(Inventory)

	for _, region := range ssmRegions {
		wg.Add(1)

		go func(region string) {
			defer wg.Done()
			err := GetRegionInventory(region, invList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering inventory list for region [%s]", region), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}

	wg.Wait()

	return invList, errs
}

// GetRegionInventory returns a slice of Inventory into the passed Inventories slice based on the provided region and search term, and optional running flag
func GetRegionInventory(region string, invList *Inventory, search string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ssm.New(sess)

	result, err := svc.GetInventory(&ssm.GetInventoryInput{})
	if err != nil {
		return err
	}

	instList := new(Instances)
	GetRegionInstances(region, instList, "", true)

	inventory := make(Inventory, len(result.Entities))
	for i, entity := range result.Entities {
		inventory[i].Marshal(entity, region, instList)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, in := range inventory {
			rInst := reflect.ValueOf(in)

			for k := 0; k < rInst.NumField(); k++ {
				sVal := rInst.Field(k).String()

				if term.MatchString(sVal) {
					*invList = append(*invList, inventory[i])
					continue Loop
				}
			}
		}
	} else {
		*invList = append(*invList, inventory[:]...)
	}

	return nil
}

// ListCommandInvocations returns a list of Command Invocations
func ListCommandInvocations(search string, details bool) (*CommandInvocations, []error) {
	var wg sync.WaitGroup
	var errs []error

	cmdInvocationsList := new(CommandInvocations)

	for _, region := range ssmRegions {
		wg.Add(1)

		go func(region string) {
			defer wg.Done()
			err := GetRegionCommandInvocations(region, cmdInvocationsList, search, details)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering command invocations list for region [%s]", region), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}

	wg.Wait()

	return cmdInvocationsList, errs
}

// GetRegionCommandInvocations returns a slice of Command Invocations into the passed CommandInvocations slice based on the provided region and search term, and optional details flag
func GetRegionCommandInvocations(region string, cmdInvocationsList *CommandInvocations, search string, details bool) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ssm.New(sess)

	result, err := svc.ListCommandInvocations(&ssm.ListCommandInvocationsInput{
		Details: aws.Bool(details),
	})
	if err != nil {
		return err
	}

	instList := new(Instances)
	GetRegionInstances(region, instList, "", false)

	cmdInvo := make(CommandInvocations, len(result.CommandInvocations))
	for i, invocation := range result.CommandInvocations {
		cmdInvo[i].Marshal(invocation, region, instList)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, in := range cmdInvo {
			rInst := reflect.ValueOf(in)

			for k := 0; k < rInst.NumField(); k++ {
				sVal := rInst.Field(k).String()

				if term.MatchString(sVal) {
					*cmdInvocationsList = append(*cmdInvocationsList, cmdInvo[i])
					continue Loop
				}
			}
		}
	} else {
		*cmdInvocationsList = append(*cmdInvocationsList, cmdInvo[:]...)
	}

	return nil
}

// GetRegionCommandInvocationssByCommandID returns a slice of Command Invocations based on the provided region and commandId, and optional details flag
func GetRegionCommandInvocationsByCommandID(region string, commandId string, details bool) (CommandInvocations, error) {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := ssm.New(sess)

	result, err := svc.ListCommandInvocations(&ssm.ListCommandInvocationsInput{
		CommandId: aws.String(commandId),
		Details:   aws.Bool(details),
	})
	if err != nil {
		return CommandInvocations{}, err
	}

	if len(result.CommandInvocations) == 0 {
		return CommandInvocations{}, nil
	}

	instList := new(Instances)
	GetRegionInstances(region, instList, "", false)

	cmdInvocations := make(CommandInvocations, len(result.CommandInvocations))
	for i, invocation := range result.CommandInvocations {
		cmdInvocations[i].Marshal(invocation, region, instList)
	}

	return cmdInvocations, nil
}

// RunCommand runs a command on one or more ec2 instances.
func RunCommand(search, command string, dryRun bool) (*CommandInvocations, error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	instList, errs := GetSSMInstances(search)
	if errs != nil {
		return &CommandInvocations{}, errors.New("Error gathering Instance list")
	}

	if len(*instList) > 0 {
		// Print the table
		instList.PrintTable()
	} else {
		return &CommandInvocations{}, errors.New("No SSM Instances found matching your search term, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to run the command [" + command + "] on these instances?") {
		return &CommandInvocations{}, errors.New("Aborting!")
	}

	// Run Em
	commandInvocations, err := runCommand(instList, command, dryRun)
	if err != nil {
		return commandInvocations, err
	}

	terminal.Information("Done!")

	return commandInvocations, nil
}

// private function without the confirmation terminal prompts
func runCommand(instList *SSMInstances, command string, dryRun bool) (*CommandInvocations, error) {

	regionInstanceIds := make(map[string][]string)
	regionInstanceNames := make(map[string][]string)

	for _, instance := range *instList {
		regionInstanceIds[instance.Region] = append(regionInstanceIds[instance.Region], instance.InstanceID)
		regionInstanceNames[instance.Region] = append(regionInstanceNames[instance.Region], instance.Name)
	}

	cmdInvocationsCombined := new(CommandInvocations)

	// Bail if on a dryRun
	// TODO more
	if dryRun {
		return cmdInvocationsCombined, nil
	}

	var wg sync.WaitGroup
	for region, instanceIds := range regionInstanceIds {
		wg.Add(1)

		terminal.Delta("Sending Command [" + command + "] to instances [" + strings.Join(regionInstanceNames[region], ", ") + "] in [" + region + "]!")

		go func(region string, instanceIds []string, command string) {
			defer wg.Done()
			sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
			svc := ssm.New(sess)

			params := &ssm.SendCommandInput{
				DocumentName: aws.String("AWS-RunShellScript"),
				InstanceIds:  aws.StringSlice(instanceIds),

				Parameters: map[string][]*string{
					"commands": {
						aws.String(command),
					},
				},
				Comment: aws.String("awsm sendCommand: " + command),
			}

			resp, err := svc.SendCommand(params)
			if err != nil {
				terminal.ErrorLine(err.Error())
				return
			} else {
				terminal.Information("Sent Command [" + command + "] [" + aws.StringValue(resp.Command.CommandId) + "] to instances [" + strings.Join(regionInstanceNames[region], ", ") + "] in [" + region + "]!")
			}

			targetCount := int(aws.Int64Value(resp.Command.TargetCount))

			for {
				cmdInvocations, err := GetRegionCommandInvocationsByCommandID(region, aws.StringValue(resp.Command.CommandId), true)
				if err != nil {
					terminal.ErrorLine(err.Error())
					break
				}

				if len(cmdInvocations) != targetCount || !cmdInvocations.Finished() {
					terminal.Notice("Waiting for response from [" + region + "]..")
					time.Sleep(time.Second * 10)
				} else {
					terminal.Information("Recieved a response from [" + region + "]!")
					*cmdInvocationsCombined = append(*cmdInvocationsCombined, cmdInvocations...)
					break
				}
			}

		}(region, instanceIds, command)
	}

	wg.Wait()

	return cmdInvocationsCombined, nil
}

//
func (i *CommandInvocations) PrintOutput() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No SSM Command Invocations Found!")
		return
	}

	terminal.Information(fmt.Sprintf("Found [%d] SSM Command Invocations!", len(*i)))
	terminal.HR()

	for _, invocation := range *i {
		fmt.Println("\n   Instance: " + invocation.InstanceName + ", " + invocation.InstanceID)
		fmt.Println("     Region: " + invocation.Region)
		fmt.Println("     Status: " + invocation.StatusDetails)
		fmt.Println("  Requested: " + humanize.Time(invocation.RequestedDateTime))
		fmt.Println(" Command ID: " + invocation.CommandID)
		fmt.Println("    Comment: " + invocation.Comment)

		for pluginIndex, plugin := range invocation.CommandPlugins {
			fmt.Printf("\n Command #%d [%s] Output:\n", pluginIndex+1, plugin.Name)
			fmt.Println(plugin.Output)
			terminal.HR()
		}
	}
}

// Finished checks if any of the command invocations within the slice are still in progress, and returns true when they are not.
func (i *CommandInvocations) Finished() bool {
	if len(*i) == 0 {
		return false
	}

	for _, invocation := range *i {
		if invocation.Status == "" || invocation.Status == "InProgress" {
			return false
		}
	}

	return true
}

// DeregisterInstances deregisters an EC2 instances from SSM Inventory
func DeregisterInstances(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	invList := new(Inventory)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionInventory(region, invList, search)
	} else {
		invList, _ = GetInventory(search)
	}

	if err != nil {
		return errors.New("Error gathering Inventory list")
	}

	if len(*invList) > 0 {
		// Print the table
		invList.PrintTable()
	} else {
		return errors.New("No Inventory found!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to deregister these Instances?") {
		return errors.New("Aborting!")
	}

	// Deregister 'Em
	err = deregisterInstances(invList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func deregisterInstances(invList *Inventory, dryRun bool) (err error) {
	if !dryRun {
		for _, entity := range *invList {
			sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(entity.Region)}))
			svc := ssm.New(sess)

			params := &ssm.DeregisterManagedInstanceInput{
				InstanceId: aws.String(entity.EntityID),
			}

			_, err := svc.DeregisterManagedInstance(params)
			if err != nil {
				fmt.Println(err)
				if awsErr, ok := err.(awserr.Error); ok {
					return errors.New(awsErr.Message())
				}
				return err
			}

			terminal.Delta("Deregistered Instance [" + entity.EntityID + "] named [" + entity.InstanceName + "] in [" + entity.Region + "]!")
		}
	}

	return nil
}

// Marshal parses the response from the aws sdk into an awsm CommandInvocation
func (i *CommandInvocation) Marshal(invocation *ssm.CommandInvocation, region string, instList *Instances) {

	i.CommandID = aws.StringValue(invocation.CommandId)
	i.Comment = aws.StringValue(invocation.Comment)
	i.Status = aws.StringValue(invocation.Status)
	i.StatusDetails = aws.StringValue(invocation.StatusDetails)
	i.DocumentName = aws.StringValue(invocation.DocumentName)
	i.InstanceID = aws.StringValue(invocation.InstanceId)
	i.InstanceName = aws.StringValue(invocation.InstanceName)
	i.RequestedDateTime = aws.TimeValue(invocation.RequestedDateTime)
	i.ServiceRole = aws.StringValue(invocation.ServiceRole)
	i.Region = region

	if i.InstanceName == "" {
		instance := instList.GetInstanceName(i.InstanceID)
		i.InstanceName = instance
	}

	// Handle plugins
	plugins := make(CommandPlugins, len(invocation.CommandPlugins))
	for ind, plugin := range invocation.CommandPlugins {
		plugins[ind].Marshal(plugin)
		i.CommandPlugins = append(i.CommandPlugins, models.CommandPlugin(plugins[ind]))
	}
}

// Marshal parses the response from the aws sdk into an awsm CommandPlugin
func (p *CommandPlugin) Marshal(plugin *ssm.CommandPlugin) {
	p.Name = aws.StringValue(plugin.Name)
	p.Output = aws.StringValue(plugin.Output)
	p.ResponseCode = int(aws.Int64Value(plugin.ResponseCode))
	p.ResponseStartDateTime = aws.TimeValue(plugin.ResponseStartDateTime)
	p.ResponseFinishDateTime = aws.TimeValue(plugin.ResponseFinishDateTime)
}

// Marshal parses the response from the aws sdk into an awsm SSMInstance
func (i *SSMInstance) Marshal(instance *ssm.InstanceInformation, region string, instList *Instances) {
	i.ComputerName = aws.StringValue(instance.ComputerName)
	i.InstanceID = aws.StringValue(instance.InstanceId)
	i.Name = instList.GetInstanceName(i.InstanceID)
	i.Class = instList.GetInstanceClass(i.InstanceID)
	i.IPAddress = aws.StringValue(instance.IPAddress)
	i.LastPingDateTime = aws.TimeValue(instance.LastPingDateTime)
	i.PingStatus = aws.StringValue(instance.PingStatus)
	i.PlatformName = aws.StringValue(instance.PlatformName)
	i.PlatformType = aws.StringValue(instance.PlatformType)
	i.PlatformVersion = aws.StringValue(instance.PlatformVersion)
	i.AgentVersion = aws.StringValue(instance.AgentVersion)
	i.IsLatestVersion = aws.BoolValue(instance.IsLatestVersion)
	i.ResourceType = aws.StringValue(instance.ResourceType)
	i.Region = region
}

// Marshal parses the response from the aws sdk into an awsm Entity
func (e *Entity) Marshal(entity *ssm.InventoryResultEntity, region string, instList *Instances) {

	e.EntityID = aws.StringValue(entity.Id)
	e.Region = region

	if instanceInfo, ok := entity.Data["AWS:InstanceInformation"]; ok {

		e.InstanceName = instList.GetInstanceName(e.EntityID)
		e.InstanceClass = instList.GetInstanceClass(e.EntityID)

		e.ContentHash = aws.StringValue(instanceInfo.ContentHash)
		e.TypeName = aws.StringValue(instanceInfo.TypeName)
		e.SchemaVersion = aws.StringValue(instanceInfo.SchemaVersion)

		layout := "2006-01-02T15:04:05Z"
		captureTime, _ := time.Parse(layout, aws.StringValue(instanceInfo.CaptureTime))
		e.CaptureTime = captureTime

		if len(instanceInfo.Content) == 1 {
			content := instanceInfo.Content[0]
			e.ComputerName = aws.StringValue(content["ComputerName"])
			e.IpAddress = aws.StringValue(content["IpAddress"])
			e.PlatformName = aws.StringValue(content["PlatformName"])
			e.PlatformType = aws.StringValue(content["PlatformType"])
			e.PlatformVersion = aws.StringValue(content["PlatformVersion"])
			e.ResourceType = aws.StringValue(content["ResourceType"])
			e.AgentType = aws.StringValue(content["AgentType"])
			e.AgentVersion = aws.StringValue(content["AgentVersion"])
		}
	}
}

// PrintTable Prints an ascii table of the list of Command Invocations and their output
func (i *CommandInvocations) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Command Invocations Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for invocationIndex, invocation := range *i {
		models.ExtractAwsmTable(invocationIndex, invocation, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

// PrintTable Prints an ascii table of the list of the Inventory
func (i *SSMInstances) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No SSM Instances Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, entity := range *i {
		models.ExtractAwsmTable(index, entity, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}

// PrintTable Prints an ascii table of the list of the Inventory
func (i *Inventory) PrintTable() {
	if len(*i) == 0 {
		terminal.ShowErrorMessage("Warning", "No Inventory Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*i))

	for index, entity := range *i {
		models.ExtractAwsmTable(index, entity, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
