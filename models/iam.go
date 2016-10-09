package models

import "time"

type IAMUser struct {
	UserName              string    `json:"userName"`
	UserId                string    `json:"userId"`
	CreateDate            time.Time `json:"createDate"`
	CreatedHuman          string    `json:"createdHumand"`
	Arn                   string    `json:"arn"`
	PasswordLastUsed      time.Time `json:"passwordLastUsed"`
	PasswordLastUsedHuman string    `json:"passwordLastUsedHuman"`
}
