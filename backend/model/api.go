package model

// Api 接口权限表（数据库可配置的接口-权限映射）
type Api struct {
	APIID      uint      `gorm:"primaryKey" json:"apiId"`
	APIName    string    `gorm:"size:64" json:"apiName"`
	Path       string    `gorm:"size:200" json:"path"`           // 支持 :param 模板，如 /api/system/user/:id
	Method     string    `gorm:"size:8" json:"method"`           // GET/POST/PUT/DELETE
	Perms      string    `gorm:"size:64" json:"perms"`           // 权限字符，空表示仅需登录
	Status     string    `gorm:"size:8;default:0" json:"status"` // 0启用 1停用（停用后不限制）
	Remark     string    `gorm:"size:255" json:"remark"`
	CreateTime LocalTime `gorm:"column:create_time" json:"createTime"`
}

// TableName 指定表名
func (Api) TableName() string { return "sys_api" }
