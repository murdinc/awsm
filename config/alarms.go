package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/simpledb"
)

// AlarmClasses is a map of Alarm Classes
type AlarmClasses map[string]AlarmClass

// AlarmClass is a single Alarm Class
type AlarmClass struct {
	AlarmDescription        string   `json:"alarmDescription" awsmList:"Alarm Description"`
	AlarmActions            []string `json:"alarmActions" awsmList:"Alarm Actions"`
	OKActions               []string `json:"okActions" awsmList:"OK Actions"`
	InsufficientDataActions []string `json:"insufficientDataActions" awsmList:"Insufficient Data Actions"`
	MetricName              string   `json:"metricName" awsmList:"Metric Name"`
	Namespace               string   `json:"namespace awsmList:"Namespace"`
	Statistic               string   `json:"statistic" awsmList:"Statistic"`
	Period                  int      `json:"period" awsmList:"Period"`
	EvaluationPeriods       int      `json:"evaluationPeriods" awsmList:"Evaluation Periods"`
	Threshold               float64  `json:"threshold" awsmList:"Threshold"`
	ComparisonOperator      string   `json:"comparisonOperator" awsmList:"Comparison Operator"`
	ActionsEnabled          bool     `json:"actionsEnabled" awsmList:"Actions Enabled"`
	Unit                    string   `json:"unit" awsmList:"Unit"`
}

// DefaultAlarms returns the defauly Alarm Classes
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

// SaveAlarmClass reads unmarshals a byte slice and inserts it into the db
func SaveAlarmClass(className string, data []byte) (class AlarmClass, err error) {
	err = json.Unmarshal(data, &class)
	if err != nil {
		return
	}

	err = InsertClasses("alarms", AlarmClasses{className: class})
	return
}

// LoadAlarmClass loads a single Alarm Class
func LoadAlarmClass(name string) (AlarmClass, error) {
	cfgs := make(AlarmClasses)
	item, err := GetItemByName("alarms", name)
	if err != nil {
		return cfgs[name], err
	}

	cfgs.Marshal([]*simpledb.Item{item})
	return cfgs[name], nil
}

// LoadAllAlarmClasses loads all Alarm Classes
func LoadAllAlarmClasses() (AlarmClasses, error) {
	cfgs := make(AlarmClasses)
	items, err := GetItemsByType("alarms")
	if err != nil {
		return cfgs, err
	}

	cfgs.Marshal(items)
	return cfgs, nil
}

// Marshal puts the items from simpledb into an AlarmClass struct
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
