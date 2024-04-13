package models

import "time"

type XxlJobLogGlue struct {
	Id         int       `json:"id" xorm:"int(11) notnull pk autoincr"`
	JobId      int       `json:"job_id" xorm:"int(11) notnull"`            // 任务主键ID
	GlueType   string    `json:"glue_type" xorm:"varchar(50)"`             // GLUE类型	#com.xxl.job.core.glue.GlueTypeEnum
	GlueSource string    `json:"glue_source" xorm:"mediumtext"`            // GLUE源代码
	GlueRemark string    `json:"glue_remark" xorm:"varchar(128) notnull"`  // GLUE备注
	AddTime    time.Time `json:"add_time" xorm:"datetime created"`         // 添加时间
	UpdateTime time.Time `json:"update_time" xorm:"datetime default null"` // 更新时间
}
