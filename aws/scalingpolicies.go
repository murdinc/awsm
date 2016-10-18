package aws

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/awsm/models"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type ScalingPolicies []ScalingPolicy

type ScalingPolicy models.ScalingPolicy

func GetScalingPolicies() (*ScalingPolicies, []error) {
	var wg sync.WaitGroup
	var errs []error

	spList := new(ScalingPolicies)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionScalingPolicies(*region.RegionName, spList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering scaling policy list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return spList, errs
}

func GetRegionScalingPolicies(region string, spList *ScalingPolicies) error {
	svc := autoscaling.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribePolicies(&autoscaling.DescribePoliciesInput{})
	if err != nil {
		return err
	}

	sp := make(ScalingPolicies, len(result.ScalingPolicies))
	for i, policy := range result.ScalingPolicies {
		sp[i].Marshal(policy, region)
	}
	*spList = append(*spList, sp[:]...)

	return nil
}

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
