package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type AlarmClasses map[string]AlarmClass

type AlarmClass struct {
	AlarmDescription        string   `json:"alarmDescription"`
	AlarmActions            []string `json:"alarmActions"`
	OKActions               []string `json:"okActions"`
	InsufficientDataActions []string `json:"insufficientDataActions"`
	MetricName              string   `json:"metricName"`
	Namespace               string   `json:"namespace`
	Statistic               string   `json:"statistic"`
	Period                  int      `json:"period"`
	EvaluationPeriods       int      `json:"evaluationPeriods"`
	Threshold               float64  `json:"threshold"`
	ComparisonOperator      string   `json:"comparisonOperator"`
	ActionsEnabled          bool     `json:"actionsEnabled"`
	Unit                    string   `json:"unit"`
}

func DefaultAlarms() AlarmClasses {
	defaultAlarms := make(AlarmClasses)

	defaultAlarms["cpuHigh"] = AlarmClass{
		AlarmDescription:   "Scale-up based on CPU",
		OKActions:          []string{},
		MetricName:         "CPUUtilization",
		Namespace:          "AWS/EC2",
		Statistic:          "Average",
		Period:             60,
		EvaluationPeriods:  2,
		Threshold:          60.0,
		ComparisonOperator: "GreaterThanThreshold",
		ActionsEnabled:     true,
	}

	defaultAlarms["cpuLow"] = AlarmClass{
		AlarmDescription:   "Scale-down based on CPU",
		OKActions:          []string{},
		MetricName:         "CPUUtilization",
		Namespace:          "AWS/EC2",
		Statistic:          "Average",
		Period:             600,
		EvaluationPeriods:  2,
		Threshold:          20.0,
		ComparisonOperator: "LessThanThreshold",
		ActionsEnabled:     true,
	}

	return defaultAlarms
}

func LoadAlarmClass(name string) (AlarmClass, error) {
	cfgs := make(AlarmClasses)
	item, err := GetItemByName("alarms", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllAlarmClasses() (AlarmClasses, error) {
	cfgs := make(AlarmClasses)
	items, err := GetItemsByType("alarms")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c AlarmClasses) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "alarms/", "", -1)
		cfg := new(AlarmClass)

		for _, attribute := range item.Attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "AlarmDescription":
				cfg.AlarmDescription = val

			case "OKActions":
				cfg.OKActions = append(cfg.OKActions, val)

			case "AlarmActions":
				cfg.AlarmActions = append(cfg.AlarmActions, val)

			case "InsufficientDataActions":
				cfg.InsufficientDataActions = append(cfg.InsufficientDataActions, val)

			case "MetricName":
				cfg.MetricName = val

			case "Namespace":
				cfg.Namespace = val

			case "Statistic":
				cfg.Statistic = val

			case "Period":
				cfg.Period, _ = strconv.Atoi(val)

			case "EvaluationPeriods":
				cfg.EvaluationPeriods, _ = strconv.Atoi(val)

			case "Threshold":
				cfg.Threshold, _ = strconv.ParseFloat(val, 64)

			case "ComparisonOperator":
				cfg.ComparisonOperator = val

			case "ActionsEnabled":
				cfg.ActionsEnabled, _ = strconv.ParseBool(val)

			case "Unit":
				cfg.Unit = val

			}
		}

		c[name] = *cfg
	}
}
