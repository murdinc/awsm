package aws

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/config"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

// ScalingPolicies represents a slice of Scaling Policies
type ScalingPolicies []ScalingPolicy

// ScalingPolicy represents a single Scaling Policy
type ScalingPolicy models.ScalingPolicy

// GetScalingPolicies returns a slice of Scaling Policies based on the given search term
func GetScalingPolicies(search string) (*ScalingPolicies, []error) {
	var wg sync.WaitGroup
	var errs []error

	spList := new(ScalingPolicies)
	regions := GetRegionListWithoutIgnored()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionScalingPolicies(*region.RegionName, spList, search)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering scaling policy list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return spList, errs
}

// GetRegionScalingPolicies returns a slice of Scaling Policies for a region into the given ScalingPolicies slice
func GetRegionScalingPolicies(region string, spList *ScalingPolicies, search string) error {

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	svc := autoscaling.New(sess)

	result, err := svc.DescribePolicies(&autoscaling.DescribePoliciesInput{})
	if err != nil {
		return err
	}

	sp := make(ScalingPolicies, len(result.ScalingPolicies))
	for i, policy := range result.ScalingPolicies {
		sp[i].Marshal(policy, region)
	}

	if search != "" {
		term := regexp.MustCompile(search)
	Loop:
		for i, g := range sp {
			rAsg := reflect.ValueOf(g)

			for k := 0; k < rAsg.NumField(); k++ {
				sVal := rAsg.Field(k).String()

				if term.MatchString(sVal) {
					*spList = append(*spList, sp[i])
					continue Loop
				}
			}
		}
	} else {
		*spList = append(*spList, sp[:]...)
	}

	return nil
}

// GetPolicyNameByArn returns the name of a Scaling Policy given the provided arn of that policy
func (s *ScalingPolicies) GetPolicyNameByArn(arn string) string {
	for _, policy := range *s {
		if policy.Arn == arn && policy.Name != "" {
			return policy.Name
		} else if policy.Arn == arn {
			return policy.Arn
		}
	}
	return arn
}

// Marshal parses the response from the aws sdk into an awsm ScalingPolicy
func (s *ScalingPolicy) Marshal(policy *autoscaling.ScalingPolicy, region string) {
	adjustment := int(aws.Int64Value(policy.ScalingAdjustment))
	adjustmentStr := fmt.Sprint(adjustment)
	if adjustment >= 1 {
		adjustmentStr = fmt.Sprintf("+%d", adjustment)
	}

	var alarmArns []string
	var alarmNames []string

	for _, alarm := range policy.Alarms {
		arnStr := aws.StringValue(alarm.AlarmARN)
		alarmArns = append(alarmArns, arnStr)

		arn, err := ParseArn(arnStr)
		if err == nil {
			alarmNames = append(alarmNames, arn.Resource)
		} else {
			alarmNames = append(alarmNames, "?????")
		}
	}

	s.Name = aws.StringValue(policy.PolicyName)
	s.Arn = aws.StringValue(policy.PolicyARN)
	s.AdjustmentType = aws.StringValue(policy.AdjustmentType)
	s.Adjustment = adjustment
	s.AdjustmentStr = adjustmentStr
	s.Cooldown = fmt.Sprint(aws.Int64Value(policy.Cooldown))
	s.AutoScaleGroupName = aws.StringValue(policy.AutoScalingGroupName)
	s.AlarmArns = alarmArns
	s.AlarmNames = strings.Join(alarmNames, ", ")

	s.Region = region
}

// UpdateScalingPolicies
func UpdateScalingPolicies(search, region string, dryRun bool) (err error) {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	spList := new(ScalingPolicies)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionScalingPolicies(region, spList, search)
	} else {
		spList, _ = GetScalingPolicies(search)
	}

	if err != nil {
		return errors.New("Error gathering Scaling Policies list")
	}

	if len(*spList) > 0 {
		// Print the table
		spList.PrintTable()
	} else {
		return errors.New("No Scaling Policies found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to update these Scaling Policies?") {
		return errors.New("Aborting!")
	}

	// Update 'Em

	_, err = updateScalingPolicies(spList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func updateScalingPolicies(spList *ScalingPolicies, dryRun bool) (arn string, err error) {
	for _, sp := range *spList {

		terminal.Delta("Updating Scaling Policy [" + sp.Name + "] in AutoScale Group [" + sp.AutoScaleGroupName + "] in [" + sp.Region + "]")

		// Verify the scaling policy class input
		cfg, err := config.LoadScalingPolicyClass(sp.Name)
		if err != nil {
			return arn, err
		}
		terminal.Information("Found Scaling Policy class configuration for [" + sp.Name + "]")

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(sp.Region)}))
		svc := autoscaling.New(sess)

		// Create the scaling policy
		params := &autoscaling.PutScalingPolicyInput{
			AdjustmentType:       aws.String(cfg.AdjustmentType),
			AutoScalingGroupName: aws.String(sp.AutoScaleGroupName),
			PolicyName:           aws.String(sp.Name),
			ScalingAdjustment:    aws.Int64(int64(cfg.ScalingAdjustment)),
			Cooldown:             aws.Int64(int64(cfg.Cooldown)),
		}

		if !dryRun {
			resp, err := svc.PutScalingPolicy(params)

			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					return arn, errors.New(awsErr.Message())
				}
				return arn, err
			}

			arn = aws.StringValue(resp.PolicyARN)

			terminal.Delta("Updated Scaling Policy  [" + sp.Name + "] with ARN [" + arn + "] in [" + sp.Region + "]!")
		} else {
			terminal.Notice("Params:")
			fmt.Println(params.String())
		}
	}

	return arn, nil
}

// CreateScalingPolicy creates a new Scaling Policy given the provided class, and region
func CreateScalingPolicy(class, asgSearch string, dryRun bool) error {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// Verify the scaling policy class input
	cfg, err := config.LoadScalingPolicyClass(class)
	if err != nil {
		return err
	}
	terminal.Information("Found Scaling Policy class configuration for [" + class + "]")

	asgList, _ := GetAutoScaleGroups(asgSearch)

	if len(*asgList) > 0 {
		// Print the table
		asgList.PrintTable()
	} else {
		return errors.New("No AutoScaling Groups found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to create this Scaling Policy in these AutoScaling Groups?") {
		return errors.New("Aborting!")
	}

	_, err = createScalingPolicy(class, cfg, asgList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")
	return nil

}

// private function without terminal prompts
func createScalingPolicy(name string, cfg config.ScalingPolicyClass, asgList *AutoScaleGroups, dryRun bool) (arn string, err error) {

	for _, asg := range *asgList {

		terminal.Delta("Adding Scaling Policy [" + name + "] to AutoScale Group [" + asg.Name + "] in [" + asg.Region + "]")

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(asg.Region)}))
		svc := autoscaling.New(sess)

		// Create the scaling policy
		params := &autoscaling.PutScalingPolicyInput{
			AdjustmentType:       aws.String(cfg.AdjustmentType),
			AutoScalingGroupName: aws.String(asg.Name),
			PolicyName:           aws.String(name),
			ScalingAdjustment:    aws.Int64(int64(cfg.ScalingAdjustment)),
			Cooldown:             aws.Int64(int64(cfg.Cooldown)),
		}

		if !dryRun {
			resp, err := svc.PutScalingPolicy(params)

			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					return arn, errors.New(awsErr.Message())
				}
				return arn, err
			}

			arn = aws.StringValue(resp.PolicyARN)

			terminal.Delta("Created Scaling Policy  [" + name + "] with ARN [" + arn + "] in [" + asg.Region + "]!")
		} else {
			terminal.Notice("Params:")
			fmt.Println(params.String())
		}
	}

	return arn, nil
}

// DeleteScalingPolicies deletes one or more Scaling Policies that match the provided search term and optionally the provided region
func DeleteScalingPolicies(search, region string, dryRun bool) (err error) {

	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	spList := new(ScalingPolicies)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionScalingPolicies(region, spList, search)
	} else {
		spList, _ = GetScalingPolicies(search)
	}

	if err != nil {
		return errors.New("Error gathering Scaling Policies list")
	}

	if len(*spList) > 0 {
		// Print the table
		spList.PrintTable()
	} else {
		return errors.New("No Scaling Policies found, Aborting!")
	}

	// Confirm
	if !terminal.PromptBool("Are you sure you want to delete these Scaling Policies?") {
		return errors.New("Aborting!")
	}

	// Delete 'Em

	err = deleteScalingPolicies(spList, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func deleteScalingPolicies(spList *ScalingPolicies, dryRun bool) (err error) {
	for _, policy := range *spList {
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(policy.Region)}))
		svc := autoscaling.New(sess)

		params := &autoscaling.DeletePolicyInput{
			AutoScalingGroupName: aws.String(policy.AutoScaleGroupName),
			PolicyName:           aws.String(policy.Name),
		}

		// Delete it!
		if !dryRun {
			_, err := svc.DeletePolicy(params)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					return errors.New(awsErr.Message())
				}
				return err
			}

			terminal.Delta("Deleted Scaling Policy [" + policy.Name + "] from [" + policy.AutoScaleGroupName + "] in [" + policy.Region + "]!")

		} else {
			fmt.Println(params)
		}

	}

	return nil
}

// ExecuteScalingPolicies
func ExecuteScalingPolicies(search, region string, force, dryRun bool) (err error) {
	// --dry-run flag
	if dryRun {
		terminal.Information("--dry-run flag is set, not making any actual changes!")
	}

	// --force flag
	if dryRun {
		terminal.Information("--force flag is set, not honoring cooldown!")
	}

	spList := new(ScalingPolicies)

	// Check if we were given a region or not
	if region != "" {
		err = GetRegionScalingPolicies(region, spList, search)
	} else {
		spList, _ = GetScalingPolicies(search)
	}

	if err != nil {
		return errors.New("Error gathering Scaling Policies list")
	}

	if len(*spList) > 0 {
		// Print the table
		spList.PrintTable()
	} else {
		return errors.New("No Scaling Policies found, Aborting!")
	}

	// Confirm
	if !force && !terminal.PromptBool("Are you sure you want to execute these Scaling Policies?") {
		return errors.New("Aborting!")
	}

	// Execute 'Em

	_, err = executeScalingPolicies(spList, force, dryRun)
	if err != nil {
		return err
	}

	terminal.Information("Done!")

	return nil
}

// Private function without the confirmation terminal prompts
func executeScalingPolicies(spList *ScalingPolicies, force, dryRun bool) (arn string, err error) {
	for _, sp := range *spList {

		terminal.Delta("Executing Scaling Policy [" + sp.Name + "] on AutoScale Group [" + sp.AutoScaleGroupName + "] in [" + sp.Region + "]")

		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(sp.Region)}))
		svc := autoscaling.New(sess)

		// Create the scaling policy
		params := &autoscaling.ExecutePolicyInput{
			AutoScalingGroupName: aws.String(sp.AutoScaleGroupName),
			HonorCooldown:        aws.Bool(!force),
			PolicyName:           aws.String(sp.Name),
		}

		if !dryRun {
			_, err := svc.ExecutePolicy(params)

			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					return arn, errors.New(awsErr.Message())
				}
				return arn, err
			}

			terminal.Delta("Executed Scaling Policy  [" + sp.Name + "] on AutoScale Group [" + sp.AutoScaleGroupName + "] in [" + sp.Region + "]!")
		} else {
			terminal.Notice("Params:")
			fmt.Println(params.String())
		}
	}

	return arn, nil
}

// PrintTable Prints an ascii table of the list of Scaling Policies
func (s *ScalingPolicies) PrintTable() {
	if len(*s) == 0 {
		terminal.ShowErrorMessage("Warning", "No Scaling Policies Found!")
		return
	}

	var header []string
	rows := make([][]string, len(*s))

	for index, sp := range *s {
		models.ExtractAwsmTable(index, sp, &header, &rows)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.AppendBulk(rows)
	table.Render()
}
