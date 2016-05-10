package config

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
