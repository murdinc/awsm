package config

import "github.com/aws/aws-sdk-go/service/simpledb"

type AlarmClassConfigs map[string]AlarmClassConfig

type AlarmClassConfig struct {
	AlarmDescription   string
	OKActions          []string //???????
	MetricName         string
	Namespace          string
	Statistic          string
	Period             int
	EvaluationPeriods  int
	Threshold          float32 //?
	ComparisonOperator string  //?
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
	}

	return defaultAlarms
}

func (c *AlarmClassConfig) LoadConfig(class string) error {
	data, err := GetClassConfig("alarms", class)
	if err != nil {
		return err
	}

	c.Marshal(data.Attributes)

	return nil

}

func (c *AlarmClassConfig) Marshal(attributes []*simpledb.Attribute) {
	// TODO
	/*
		for _, attribute := range attributes {

			val := *attribute.Value

			switch *attribute.Name {

			case "InstanceType":
				c.InstanceType = val

			case "SecurityGroups":
				c.SecurityGroups = append(c.SecurityGroups, val)

			case "Subnet":
				c.Subnet = val

			case "PublicIpAddress":
				c.PublicIpAddress, _ = strconv.ParseBool(val)

			case "AMI":
				c.AMI = val

			case "Keys":
				c.Keys = append(c.SecurityGroups, val)

			}
		}
	*/

}
