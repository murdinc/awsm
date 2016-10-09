package models

type ARN struct {
	Arn                  string `json:"arn"`
	Partition            string `json:"partition"`
	Service              string `json:"service"`
	Region               string `json:"region"`
	AccountId            string `json:"accountId"`
	PolicyId             string `json:"policyId"`
	GroupId              string `json:"groupId"`
	AutoScalingGroupName string `json:"autoScalingGroupName"`
	PolicyName           string `json:"policyName"`
	ResourceType         string `json:"resourceType"`
	Resource             string `json:"resource"`
}
