package models

import "time"

// IAMUser represents an Identity and Access Management (IAM) User
type IAMUser struct {
	UserName              string    `json:"userName" awsmTable:"User Name"`
	UserID                string    `json:"userID" awsmTable:"User ID"`
	CreateDate            time.Time `json:"createDate"`
	CreatedHuman          string    `json:"createdHumand" awsmTable:"Created"`
	Arn                   string    `json:"arn" awsmTable:"ARN"`
	PasswordLastUsed      time.Time `json:"passwordLastUsed"`
	PasswordLastUsedHuman string    `json:"passwordLastUsedHuman" awsmTable:"Last Used"`
}
