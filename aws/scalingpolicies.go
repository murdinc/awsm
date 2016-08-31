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
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type ScalingPolicies []ScalingPolicy

type ScalingPolicy struct {
	Name               string   `json:"name"`
	Arn                string   `json:"arn"`
	AdjustmentType     string   `json:"adjustmentType"`
	Adjustment         int      `json:"adjustment"`
	AdjustmentStr      string   `json:"adjustmentStr"`
	Cooldown           string   `json:"cooldown"`
	AutoScaleGroupName string   `json:"autoScaleGroupName"`
	AlarmArns          []string `json:"alarmArns"`
	AlarmNames         string   `json:"alarmNames"`
	Region             string   `json:"region"`
}

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

func (p *ScalingPolicies) GetPolicyNameByArn(arn string) string {
	for _, policy := range *p {
		if policy.Arn == arn && policy.Name != "" {
			return policy.Name
		} else if policy.Arn == arn {
			return policy.Arn
		}
	}
	return arn
}

func (p *ScalingPolicy) Marshal(policy *autoscaling.ScalingPolicy, region string) {
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

	p.Name = aws.StringValue(policy.PolicyName)
	p.Arn = aws.StringValue(policy.PolicyARN)
	p.AdjustmentType = aws.StringValue(policy.AdjustmentType)
	p.Adjustment = adjustment
	p.AdjustmentStr = adjustmentStr
	p.Cooldown = fmt.Sprint(aws.Int64Value(policy.Cooldown))
	p.AutoScaleGroupName = aws.StringValue(policy.AutoScalingGroupName)
	p.AlarmArns = alarmArns
	p.AlarmNames = strings.Join(alarmNames, ", ")

	p.Region = region
}

func (i *ScalingPolicies) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.AdjustmentType,
			val.AdjustmentStr,
			val.Cooldown + " sec.",
			val.AutoScaleGroupName,
			val.AlarmNames,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Adjustment Type", "Adjustment", "Cooldown", "AutoScaling Group Name", "Alarms", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
