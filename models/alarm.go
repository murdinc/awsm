package models

// Alarm represents a CloudWatch Alarm
type Alarm struct {
	Name        string   `json:"name" awsmTable:"Name"`
	Arn         string   `json:"arn"`
	Description string   `json:"description" awsmTable:"Description"`
	State       string   `json:"state" awsmTable:"State"`
	Trigger     string   `json:"trigger" awsmTable:"Trigger"`
	Period      string   `json:"period" awsmTable:"Period"`
	EvalPeriods string   `json:"evalPeriods" awsmTable:"Evaluation Periods"`
	ActionArns  []string `json:"actionArns"`
	ActionNames string   `json:"actionNames" awsmTable:"Actions"`
	Dimensions  string   `json:"dimensions" awsmTable:"Dimensions"`
	Namespace   string   `json:"namespace" awsmTable:"Namespace"`
	Region      string   `json:"region" awsmTable:"Region"`
}
