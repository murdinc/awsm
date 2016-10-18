package models

import "time"

type IAMUser struct {
	UserName              string    `json:"userName"`
	UserID                string    `json:"userID"`
	CreateDate            time.Time `json:"createDate"`
	CreatedHuman          string    `json:"createdHumand"`
	Arn                   string    `json:"arn"`
	PasswordLastUsed      time.Time `json:"passwordLastUsed"`
	PasswordLastUsedHuman string    `json:"passwordLastUsedHuman"`
}
