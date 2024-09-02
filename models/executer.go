package models

import (
	"time"
	"xorm.io/xorm"
)

// Executor 执行组模型
type Executor struct {
	Id            int         `json:"id" xorm:"int pk autoincr"`
	ExecutorName  string      `json:"executor_name" xorm:"varchar(64) notnull"`           // 执行器名称
	ExecutorTitle string      `json:"executor_title" xorm:"varchar(32) notnull"`          // 执行器标题
	AddressType   AddressType `json:"address_type" xorm:"tinyint(4) notnull default (0)"` // 执行器地址类型 0：自动录入 1：手动录入
	Address       string      `json:"address" xorm:"varchar(64) notnull"`                 // 执行器地址
	UpdateTime    time.Time   `json:"update_time" xorm:"datetime default null"`           // 更新时间
	BaseModel     `xorm:"-"`
}

func (exec *Executor) Create() (insertId int, err error) {
	_, err = Db.Insert(exec)
	if err == nil {
		insertId = exec.Id
	}

	return
}

// UpdateBean 更新
func (exec *Executor) UpdateBean(id int) (int64, error) {
	return Db.ID(id).Cols("executor_name", "executor_title", "address_type", "address", "update_time").Update(exec)
}

// Update 更新
func (exec *Executor) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(exec).ID(id).Update(data)
}

// Delete 删除
func (exec *Executor) Delete(id int) (int64, error) {
	return Db.ID(id).Delete(exec)
}

// Search 查询
func (exec *Executor) Search(id int) error {
	_, err := Db.ID(id).Get(exec)

	return err
}

func (exec *Executor) GetAddress(id int) (string, error) {
	exec.Id = id
	_, err := Db.Cols("address").Get(exec)

	return exec.Address, err
}

func GetAllAddress() ([]string, error) {
	list := make([]Executor, 0)
	err := Db.Cols("address").Find(&list)
	if err != nil {
		return nil, err
	}
	addresses := make([]string, 0, len(list))
	for _, item := range list {
		addresses = append(addresses, item.Address)
	}

	return addresses, nil
}

// SelectByName 判断执行器名称是否存在
func (exec *Executor) SelectByName(name string) (bool, error) {
	count, err := Db.Where("executor_name = ?", name).Count(exec)

	return count > 0, err
}

// List 获取所有执行器
func (exec *Executor) List(params CommonMap) ([]Executor, error) {
	exec.parsePageAndPageSize(params)
	list := make([]Executor, 0)
	session := Db.Asc("id")
	exec.parseWhere(session, params)
	err := session.Limit(exec.PageSize, exec.pageLimitOffset()).Find(&list)

	return list, err
}

// AllList 获取所有执行器
func (exec *Executor) AllList() ([]Executor, error) {
	list := make([]Executor, 0)
	err := Db.Cols("id,executor_name,address").Desc("id").Find(&list)

	return list, err
}

// Total 获取执行器总数
func (exec *Executor) Total(params CommonMap) (int64, error) {
	session := Db.NewSession()
	exec.parseWhere(session, params)
	return session.Count(exec)
}

func (exec *Executor) parseWhere(session *xorm.Session, params CommonMap) {
	if len(params) == 0 {
		return
	}
	id, ok := params["Id"]
	if ok && id.(int) > 0 {
		session.And("id = ?", id)
	}
	name, ok := params["Name"]
	if ok && name.(string) != "" {
		session.And("executor_name = ?", name)
	}
}
