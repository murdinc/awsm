package config

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

type AlarmClassConfigs map[string]AlarmClassConfig

type AlarmClassConfig struct {
	AlarmDescription        string
	AlarmActions            []string
	OKActions               []string
	InsufficientDataActions []string
	MetricName              string
	Namespace               string
	Statistic               string
	Period                  int
	EvaluationPeriods       int
	Threshold               float64
	ComparisonOperator      string
	ActionsEnabled          bool
	Unit                    string
}

func DefaultAlarms() AlarmClassConfigs {
	defaultAlarms := make(AlarmClassConfigs)

	defaultAlarms["cpuHigh"] = AlarmClassConfig{
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

	defaultAlarms["cpuLow"] = AlarmClassConfig{
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

func LoadAlarmClass(name string) (AlarmClassConfig, error) {
	cfgs := make(AlarmClassConfigs)
	item, err := GetItemByName("alarms", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

func LoadAllAlarmClasses() (AlarmClassConfigs, error) {
	cfgs := make(AlarmClassConfigs)
	items, err := GetItemsByType("alarms")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

func (c AlarmClassConfigs) Marshal(items []*simpledb.Item) {
	for _, item := range items {
		name := strings.Replace(*item.Name, "alarms/", "", -1)
		cfg := new(AlarmClassConfig)

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
