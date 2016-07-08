package aws

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/terminal"
	"github.com/olekukonko/tablewriter"
)

type Alarms []Alarm

type Alarm struct {
	Name        string
	Arn         string
	Description string
	State       string
	Trigger     string
	Period      string
	EvalPeriods string
	ActionArns  []string
	ActionNames string
	Dimensions  string
	Namespace   string
	Region      string
}

func GetAlarms() (*Alarms, []error) {
	var wg sync.WaitGroup
	var errs []error

	alList := new(Alarms)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionAlarms(*region.RegionName, alList)
			if err != nil {
				terminal.ShowErrorMessage(fmt.Sprintf("Error gathering alarm list for region [%s]", *region.RegionName), err.Error())
				errs = append(errs, err)
			}
		}(region)
	}
	wg.Wait()

	return alList, errs
}

func GetRegionAlarms(region string, alList *Alarms) error {
	svc := cloudwatch.New(session.New(&aws.Config{Region: aws.String(region)}))
	result, err := svc.DescribeAlarms(&cloudwatch.DescribeAlarmsInput{})
	if err != nil {
		return err
	}

	al := make(Alarms, len(result.MetricAlarms))
	for i, alarm := range result.MetricAlarms {
		al[i].Marshal(alarm, region)
	}
	*alList = append(*alList, al[:]...)

	return nil
}

func (a *Alarm) Marshal(alarm *cloudwatch.MetricAlarm, region string) {
	var dimensions []string
	var operator string

	for _, dim := range alarm.Dimensions {
		dimensions = append(dimensions, aws.StringValue(dim.Name)+" = "+aws.StringValue(dim.Value))
	}

	switch aws.StringValue(alarm.ComparisonOperator) {
	case "GreaterThanThreshold":
		operator = ">"

	case "GreaterThanOrEqualToThreshold":
		operator = ">="

	case "LessThanThreshold":
		operator = "<"

	case "LessThanOrEqualToThreshold":
		operator = "<="
	}

	var actionArns []string
	var actionNames []string

	for _, action := range alarm.AlarmActions {
		arnStr := aws.StringValue(action)
		actionArns = append(actionArns, arnStr)

		arn, err := ParseArn(arnStr)
		if err == nil {
			actionNames = append(actionNames, arn.PolicyName)
		} else {
			actionNames = append(actionNames, "??????")
		}
	}

	a.Name = aws.StringValue(alarm.AlarmName)
	a.Arn = aws.StringValue(alarm.AlarmArn)
	a.Description = aws.StringValue(alarm.AlarmDescription)
	a.State = aws.StringValue(alarm.StateValue)
	a.Trigger = fmt.Sprintf("%s %s %d (%s)", aws.StringValue(alarm.MetricName), operator, int(aws.Float64Value(alarm.Threshold)), aws.StringValue(alarm.Statistic))
	a.Period = fmt.Sprint(aws.Int64Value(alarm.Period))
	a.EvalPeriods = fmt.Sprint(aws.Int64Value(alarm.EvaluationPeriods))
	a.ActionArns = actionArns
	a.ActionNames = strings.Join(actionNames, ", ")
	a.Dimensions = strings.Join(dimensions, ", ")
	a.Namespace = aws.StringValue(alarm.Namespace)
	a.Region = region
}

func (a *Alarms) PrintTable() {
	table := tablewriter.NewWriter(os.Stdout)

	rows := make([][]string, len(*a))
	for index, val := range *a {
		rows[index] = []string{
			val.Name,
			val.Description,
			val.State,
			val.Trigger,
			val.Period,
			val.EvalPeriods,
			val.ActionNames,
			val.Dimensions,
			val.Namespace,
			val.Region,
		}
	}

	table.SetHeader([]string{"Name", "Description", "State", "Trigger", "Period", "EvalPeriods", "Actions", "Dimensions", "Namespace", "Region"})

	table.AppendBulk(rows)
	table.Render()
}
