package models

type User struct {
	Id         int    `json:"id" xorm:"int(11) notnull pk autoincr"`
	Username   string `json:"username" xorm:"varchar(50) notnull"` // 用户名
	Password   string `json:"password" xorm:"varchar(50) notnull"` // 密码
	Role       int8   `json:"role" xorm:"tinyint(4) notnull"`      // 角色 0：普通用户 1：管理员
	Permission string `json:"permission" xorm:"varchar(255)"`      // 权限 执行器列表 以逗号分隔
}
