package models

// ScalingPolicy represents an AutoScaling Group Scaling Policy
type ScalingPolicy struct {
	Name               string   `json:"name" awsmTable:"Name"`
	Arn                string   `json:"arn"`
	AdjustmentType     string   `json:"adjustmentType" awsmTable:"Adjustment Type"`
	Adjustment         int      `json:"adjustment"`
	AdjustmentStr      string   `json:"adjustmentStr" awsmTable:"Adjustment"`
	Cooldown           string   `json:"cooldown" awsmTable:"Cooldown"`
	AutoScaleGroupName string   `json:"autoScaleGroupName" awsmTable:"Autoscale Group Name"`
	AlarmArns          []string `json:"alarmArns" awsmTable:"Alarm ARNs"`
	AlarmNames         string   `json:"alarmNames" awsmTable:"Alarm Names"`
	Region             string   `json:"region" awsmTable:"Region"`
}
