package model

type UserAccount struct {
	User_id          int `json:"id"`
	Balance          int `json:"balance"`
	Reserved_balance int `json:"-"`
}
