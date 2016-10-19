package models

import "time"

// IAMUser represents an Identity and Access Management (IAM) User
type IAMUser struct {
	UserName              string    `json:"userName"`
	UserID                string    `json:"userID"`
	CreateDate            time.Time `json:"createDate"`
	CreatedHuman          string    `json:"createdHumand"`
	Arn                   string    `json:"arn"`
	PasswordLastUsed      time.Time `json:"passwordLastUsed"`
	PasswordLastUsedHuman string    `json:"passwordLastUsedHuman"`
}
