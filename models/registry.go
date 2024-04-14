package models

import "time"

// Registry 维护在线的执行器和调度中心机器地址信息，并将执行器地址更新到group中
type Registry struct {
	Id            int       `json:"id" xorm:"int(11) notnull pk autoincr" comment:"主键ID"`     // 主键ID
	RegistryGroup string    `json:"registry_group" xorm:"varchar(50) notnull" comment:"注册组"`  // 注册组
	RegistryKey   string    `json:"registry_key" xorm:"varchar(255) notnull" comment:"注册键"`   // 注册键
	RegistryValue string    `json:"registry_value" xorm:"varchar(255) notnull" comment:"注册值"` // 注册值
	UpdateTime    time.Time `json:"update_time" xorm:"datetime" comment:"更新时间"`               // 更新时间
}
