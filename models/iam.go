package models

import "time"

// IAMUser represents an Identity and Access Management (IAM) User
type IAMUser struct {
	UserName         string    `json:"userName" awsmTable:"User Name"`
	UserID           string    `json:"userID" awsmTable:"User ID"`
	CreateDate       time.Time `json:"createDate" awsmTable:"Created"`
	Arn              string    `json:"arn" awsmTable:"ARN"`
	PasswordLastUsed time.Time `json:"passwordLastUsed" awsmTable:"Last Used"`
}

// IAMRole represents an Identity and Access Management (IAM) Role
type IAMRole struct {
	RoleName                 string    `json:"roleName" awsmTable:"Role Name"`
	RoleID                   string    `json:"roleID" awsmTable:"Role ID"`
	CreateDate               time.Time `json:"createDate" awsmTable:"Created"`
	AssumeRolePolicyDocument string    `json:"assumeRolePolicyDocument"`
	Arn                      string    `json:"arn" awsmTable:"ARN"`
}

// IAMInstanceProfile represents an Identity and Access Management (IAM) Profile
type IAMInstanceProfile struct {
	ProfileName string    `json:"profileName" awsmTable:"Profile Name"`
	ProfileID   string    `json:"profileID" awsmTable:"Profile ID"`
	CreateDate  time.Time `json:"createDate" awsmTable:"Created"`
	Arn         string    `json:"arn" awsmTable:"ARN"`
}

// IAMPolicy represents an Identity and Access Management (IAM) Policy
type IAMPolicy struct {
	PolicyName       string    `json:"policyName" awsmTable:"Policy Name"`
	PolicyID         string    `json:"policyID" awsmTable:"Policy ID"`
	Description      string    `json:"description" awsmTable:"Description"`
	IsAttachable     bool      `json:"isAttachable" awsmTable:"Attachable"`
	AttachmentCount  int       `json:"attachmentCount" awsmTable:"Attachment Count"`
	CreateDate       time.Time `json:"createDate" awsmTable:"Created"`
	DefaultVersionId string    `json:"defaultVersionId" awsmTable:"Version"`
	Arn              string    `json:"arn" awsmTable:"ARN"`
}

// IAMPolicyDocument represents an Identity and Access Management (IAM) Policy Document
type IAMPolicyDocument struct {
	Document         string    `json:"document" awsmTable:"Document"`
	CreateDate       time.Time `json:"createDate" awsmTable:"Created"`
	VersionId        string    `json:"versionId" awsmTable:"Version ID"`
	IsDefaultVersion bool      `json:"isDefaultVersion" awsmTable:"Default Version"`
}
