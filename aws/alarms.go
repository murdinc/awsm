package aws

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/murdinc/cli"
)

type Alarms []Alarm

type Alarm struct {
	Name        string
	Description string
	State       string
	Trigger     string
	Period      string
	EvalPeriods string
	Actions     string
	Dimensions  string
	Namespace   string
	Region      string
}

func GetAlarms() (*Alarms, error) {
	var wg sync.WaitGroup

	alList := new(Alarms)
	regions := GetRegionList()

	for _, region := range regions {
		wg.Add(1)

		go func(region *ec2.Region) {
			defer wg.Done()
			err := GetRegionAlarms(region.RegionName, alList)
			if err != nil {
				cli.ShowErrorMessage("Error gathering alarm list", err.Error())
			}
		}(region)
	}
	wg.Wait()

	return alList, nil
}

func GetRegionAlarms(region *string, alList *Alarms) error {
	svc := cloudwatch.New(session.New(&aws.Config{Region: region}))
	result, err := svc.DescribeAlarms(&cloudwatch.DescribeAlarmsInput{})

	if err != nil {
		return err
	}

	al := make(Alarms, len(result.MetricAlarms))
	for i, alarm := range result.MetricAlarms {

		al[i] = Alarm{
			Name:        aws.StringValue(alarm.AlarmName),
			Description: aws.StringValue(alarm.AlarmDescription),
			State:       aws.StringValue(alarm.StateValue),
			Trigger:     aws.StringValue(alarm.MetricName),
			Period:      string(aws.Int64Value(alarm.Period)),
			EvalPeriods: string(*alarm.EvaluationPeriods),
			//Actions:     fmt.Sprint(*alarm.AlarmActions), // TODO
			//Dimensions: *alarm.Dimensions, // TODO
			Namespace: aws.StringValue(alarm.Namespace),
			Region:    fmt.Sprintf(*region),
		}
	}
	*alList = append(*alList, al[:]...)

	return nil
}

func (i *Alarms) PrintTable() {
	collumns := []string{"Name", "Description", "State", "Trigger", "Period", "EvalPeriods", "Actions", "Dimensions", "Namespace", "Region"}

	rows := make([][]string, len(*i))
	for index, val := range *i {
		rows[index] = []string{
			val.Name,
			val.Description,
			val.State,
			val.Trigger,
			val.Period,
			val.EvalPeriods,
			val.Actions,
			val.Dimensions,
			val.Namespace,
			val.Region,
		}
	}

	printTable(collumns, rows)
}
