package model

import "time"

type Transaction struct {
	Id          int       `json:"id"`
	User_id     int       `json:"userId"`
	Amount      int       `json:"amount"`
	Description string    `json:"description"`
	Order_id    int       `json:"orderId"`
	Service_id  int       `json:"serviceId"`
	Closed_date time.Time `json:"closedDate"`
	Success_flg bool      `json:"-"`
	Type        string    `json:"-"`
}

type AccountTransaction struct {
	Amount      int       `json:"amount"`
	Description string    `json:"description"`
	Order_id    int       `json:"orderId"`
	Service     string    `json:"service"`
	Closed_date time.Time `json:"closedDate"`
}
