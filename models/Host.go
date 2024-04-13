package models

type Host struct {
	Id int16 `json:"id" xorm:"smallint pk autoincr"`
}
