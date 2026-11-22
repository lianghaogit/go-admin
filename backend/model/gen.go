package model

// GenTable 代码生成-表配置
type GenTable struct {
	TableID      uint        `gorm:"primaryKey" json:"tableId"`
	TblName      string      `gorm:"column:table_name;size:64" json:"tableName"` // 表名称
	TableComment string      `gorm:"size:255" json:"tableComment"`               // 表描述
	ClassName    string      `gorm:"size:64" json:"className"`                   // 实体类名称（Go 结构体名）
	ModuleName   string      `gorm:"size:64" json:"moduleName"`                  // 前端模块名（如 system）
	PackageName  string      `gorm:"size:64" json:"packageName"`                 // 后端包名（api/dao/model 所属目录名）
	FunctionName string      `gorm:"size:64" json:"functionName"`                // 功能名称（如 用户）
	Author       string      `gorm:"size:64" json:"author"`                      // 生成作者
	TplCategory  string      `gorm:"size:16;default:crud" json:"tplCategory"`    // 模板类型：crud 单表 / tree 树形 / sub 主子表
	TreeCode     string      `gorm:"size:64" json:"treeCode"`                    // 树形：显示编码列
	TreeName     string      `gorm:"size:64" json:"treeName"`                    // 树形：显示名称列
	TreeParent   string      `gorm:"size:64" json:"treeParent"`                  // 树形：父级列（默认 parent_id）
	SubTableName string      `gorm:"size:64" json:"subTableName"`                // 主子表：子表表名
	SubTableFk   string      `gorm:"size:64" json:"subTableFk"`                  // 主子表：子表关联主表的外键列
	Remark       string      `gorm:"size:255" json:"remark"`
	CreateTime   LocalTime   `json:"createTime"`
	UpdateTime   LocalTime   `json:"updateTime"`
	Columns      []GenColumn `gorm:"foreignKey:TableID" json:"columns"`
}

// TableName 指定表名
func (GenTable) TableName() string { return "sys_gen_table" }

// GenColumn 代码生成-列配置
type GenColumn struct {
	ColumnID      uint   `gorm:"primaryKey" json:"columnId"`
	TableID       uint   `gorm:"index" json:"tableId"`
	ColumnName    string `gorm:"size:64" json:"columnName"`             // 列名
	ColumnComment string `gorm:"size:255" json:"columnComment"`         // 列注释
	ColumnType    string `gorm:"size:64" json:"columnType"`             // 原始列类型
	GoType        string `gorm:"size:32" json:"goType"`                 // Go 类型（int64/string/float64）
	IsPk          int    `gorm:"default:0" json:"isPk"`                 // 是否主键 1是
	IsIncrement   int    `gorm:"default:0" json:"isIncrement"`          // 是否自增 1是
	IsRequired    int    `gorm:"default:0" json:"isRequired"`           // 是否必填（校验）
	IsList        int    `gorm:"default:0" json:"isList"`               // 是否列表显示
	IsQuery       int    `gorm:"default:0" json:"isQuery"`              // 是否查询条件
	QueryType     string `gorm:"size:16;default:EQ" json:"queryType"`   // 查询方式 EQ/NE/GT/LT/LIKE
	IsForm        int    `gorm:"default:0" json:"isForm"`               // 是否表单显示
	FormType      string `gorm:"size:32;default:input" json:"formType"` // 表单类型 input/textarea/select/date
	Sort          int    `gorm:"default:0" json:"sort"`
}

// TableName 指定表名
func (GenColumn) TableName() string { return "sys_gen_column" }
