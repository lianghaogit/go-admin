package dao

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"go-admin/model"

	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// 数据库表结构读取（按方言：sqlite / mysql / postgres）
// ---------------------------------------------------------------------------

// DbTableMeta 数据库表元信息（用于导入选择）
type DbTableMeta struct {
	TableName    string `json:"tableName"`
	TableComment string `json:"tableComment"`
	CreateTime   string `json:"createTime"`
	Engine       string `json:"engine"`
}

// ColumnMeta 列元信息（不同方言的信息表映射为统一结构）
type ColumnMeta struct {
	CID          int    `gorm:"column:cid" json:"cid"`
	Name         string `gorm:"column:name" json:"name"`
	Type         string `gorm:"column:type" json:"type"`
	NotNull      int    `gorm:"column:notnull" json:"notnull"`
	DefaultValue string `gorm:"column:dflt_value" json:"defaultValue"`
	PK           int    `gorm:"column:pk" json:"pk"`
	Comment      string `gorm:"column:comment" json:"comment"` // 列注释（PG: col_description；MySQL: column_comment）
}

// ListDbTables 列出数据库中未被导入的业务表（排除 sys_ 内置表与各方言系统表）
func ListDbTables(keyword string) ([]DbTableMeta, error) {
	var tables []struct {
		Name    string `gorm:"column:name"`
		SQLText string `gorm:"column:sql"`
	}
	// 各方言返回 name + 注释（sql 别名承载注释文本，SQLite 无注释时为空）
	var listErr error
	switch Dialect() {
	case "mysql":
		listErr = DB.Raw(`SELECT table_name AS name, IFNULL(table_comment, '') AS sql
			FROM information_schema.tables
			WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE' ORDER BY table_name`).Scan(&tables).Error
	case "postgres":
		listErr = DB.Raw(`SELECT tablename AS name, COALESCE(obj_description(('public.' || quote_ident(tablename))::regclass, 'pg_class'), '') AS sql
			FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename`).Scan(&tables).Error
	default:
		listErr = DB.Raw(`SELECT name, sql FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`).Scan(&tables).Error
	}
	if listErr != nil {
		return nil, listErr
	}
	// 已导入的表
	var imported []string
	if err := DB.Model(&model.GenTable{}).Pluck("table_name", &imported).Error; err != nil {
		return nil, err
	}
	importedSet := make(map[string]bool, len(imported))
	for _, t := range imported {
		importedSet[t] = true
	}
	// 内置表排除前缀
	excludePrefix := []string{"sys_"}
	result := make([]DbTableMeta, 0, len(tables))
	for _, t := range tables {
		name := t.Name
		skip := false
		for _, p := range excludePrefix {
			if strings.HasPrefix(name, p) {
				skip = true
				break
			}
		}
		if skip || importedSet[name] {
			continue
		}
		comment := parseTableComment(t.SQLText)
		if keyword != "" && !strings.Contains(name, keyword) && !strings.Contains(comment, keyword) {
			continue
		}
		result = append(result, DbTableMeta{TableName: name, TableComment: comment, CreateTime: time.Now().Format("2006-01-02 15:04:05")})
	}
	return result, nil
}

// parseTableComment 从元数据注释字段提取表注释（SQLite 建表语句一般无 COMMENT，返回空由上层回退到表名）
func parseTableComment(sqlText string) string {
	if Dialect() == "sqlite" {
		return ""
	}
	return sqlText // mysql: table_comment；postgres: obj_description
}

// tableColumns 按方言读取指定表的列元信息
func tableColumns(tableName string) ([]ColumnMeta, error) {
	var cols []ColumnMeta
	var err error
	switch Dialect() {
	case "mysql":
		err = DB.Raw(`SELECT ordinal_position AS cid, column_name AS name, column_type AS type,
				IF(is_nullable = 'NO', 1, 0) AS notnull, column_default AS dflt_value,
				IF(column_key = 'PRI', 1, 0) AS pk, IFNULL(column_comment, '') AS comment
			FROM information_schema.columns
			WHERE table_schema = DATABASE() AND table_name = ? ORDER BY ordinal_position`, tableName).Scan(&cols).Error
	case "postgres":
		// format_type 带长度/精度（numeric(10,2)、varchar(64)、timestamp with time zone），
		// 比 information_schema.data_type 更完整；列注释走 col_description
		err = DB.Raw(`SELECT a.attnum AS cid,
				a.attname AS name,
				format_type(a.atttypid, a.atttypmod) AS type,
				CASE WHEN a.attnotnull THEN 1 ELSE 0 END AS notnull,
				COALESCE(pg_get_expr(d.adbin, d.adrelid), '') AS dflt_value,
				CASE WHEN EXISTS (
					SELECT 1 FROM pg_index i
					WHERE i.indrelid = a.attrelid AND i.indisprimary AND a.attnum = ANY(i.indkey)
				) THEN 1 ELSE 0 END AS pk,
				COALESCE(col_description(a.attrelid, a.attnum), '') AS comment
			FROM pg_attribute a
			LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
			WHERE a.attrelid = ('public.' || quote_ident(?))::regclass
				AND a.attnum > 0 AND NOT a.attisdropped
			ORDER BY a.attnum`, tableName).Scan(&cols).Error
	default:
		err = DB.Raw(fmt.Sprintf("PRAGMA table_info(%s)", quoteIdent(tableName))).Scan(&cols).Error
	}
	return cols, err
}

func quoteIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "") + "`"
}

// ---------------------------------------------------------------------------
// 命名与类型转换
// ---------------------------------------------------------------------------

// snakeToPascal user_name -> UserName
func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	return b.String()
}

// snakeToCamel user_name -> userName
func snakeToCamel(s string) string {
	p := snakeToPascal(s)
	if p == "" {
		return p
	}
	return strings.ToLower(p[:1]) + p[1:]
}

// toSnakeCase TestCustomer -> test_customer（Pascal/驼峰转蛇形）
func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// tableToClassName 表名转结构体类名：sys_user -> User；t_biz_order -> BizOrder；user -> User
func tableToClassName(tableName string) string {
	name := tableName
	for _, prefix := range []string{"sys_", "t_"} {
		if strings.HasPrefix(name, prefix) {
			name = strings.TrimPrefix(name, prefix)
			break
		}
	}
	return snakeToPascal(name)
}

// dbTypeToGoType 列类型映射到 Go 类型（时间字段统一 string，与项目既有约定一致）
// 兼容 sqlite 声明类型、MySQL column_type 与 PostgreSQL format_type（后者会带长度/精度与修饰词）
func dbTypeToGoType(decl string, colName string) string {
	t := strings.ToUpper(strings.TrimSpace(decl))
	// 截取第一个 '(' 前的类型名：varchar(64) -> VARCHAR，numeric(10,2) -> NUMERIC
	if i := strings.Index(t, "("); i > 0 {
		t = t[:i]
	}
	t = strings.TrimSpace(t)
	// PostgreSQL 的带时区/不带时区时间类型先归一
	if strings.Contains(t, "WITH TIME ZONE") || strings.Contains(t, "WITHOUT TIME ZONE") {
		return "model.LocalTime"
	}
	lower := strings.ToLower(colName)
	if strings.Contains(lower, "time") || strings.Contains(lower, "date") ||
		strings.HasPrefix(t, "TIMESTAMP") || t == "DATE" || t == "TIME" {
		return "model.LocalTime"
	}
	switch t {
	case "BOOLEAN", "BOOL":
		return "bool"
	case "INTEGER", "INT", "INT2", "INT4", "INT8", "BIGINT", "SMALLINT", "TINYINT", "MEDIUMINT",
		"SERIAL", "BIGSERIAL", "SMALLSERIAL":
		return "int64"
	case "REAL", "DOUBLE", "DOUBLE PRECISION", "FLOAT", "FLOAT4", "FLOAT8",
		"NUMERIC", "DECIMAL", "MONEY":
		return "float64"
	default:
		return "string"
	}
}

// tableComment 读取表注释（PostgreSQL: obj_description；MySQL: table_comment；SQLite 无注释）
func tableComment(tableName string) string {
	var c string
	switch Dialect() {
	case "postgres":
		err := DB.Raw(`SELECT COALESCE(obj_description(('public.' || quote_ident(?))::regclass, 'pg_class'), '')`,
			tableName).Scan(&c).Error
		if err != nil {
			return ""
		}
	case "mysql":
		err := DB.Raw(`SELECT IFNULL(table_comment, '') FROM information_schema.tables
			WHERE table_schema = DATABASE() AND table_name = ?`, tableName).Scan(&c).Error
		if err != nil {
			return ""
		}
	}
	return strings.TrimSpace(c)
}

// ---------------------------------------------------------------------------
// 导入与配置
// ---------------------------------------------------------------------------

// ImportTables 将数据库表导入生成配置（写入 gen_table + gen_column）
func ImportTables(tableNames []string) (int, error) {
	imported := 0
	err := DB.Transaction(func(tx *gorm.DB) error {
		for _, name := range tableNames {
			name = strings.TrimSpace(name)
			if name == "" || strings.HasPrefix(name, "sys_") || strings.HasPrefix(name, "sqlite_") {
				continue
			}
			var cnt int64
			if err := tx.Model(&model.GenTable{}).Where("table_name = ?", name).Count(&cnt).Error; err != nil {
				return err
			}
			if cnt > 0 {
				continue
			}
			cols, err := tableColumns(name)
			if err != nil {
				return err
			}
			if len(cols) == 0 {
				continue
			}
			table := model.GenTable{
				TblName: name,
				TableComment: func() string {
					if c := tableComment(name); c != "" {
						return c
					}
					return name
				}(),
				ClassName:    tableToClassName(name),
				ModuleName:   "system",
				PackageName:  "system",
				FunctionName: tableToClassName(name),
				Author:       "go-admin",
				CreateTime:   model.LocalTime(time.Now()),
				UpdateTime:   model.LocalTime(time.Now()),
			}
			if err := tx.Create(&table).Error; err != nil {
				return err
			}
			// 分析列
			for _, c := range cols {
				col := buildColumn(c, table.TableID)
				if err := tx.Create(&col).Error; err != nil {
					return err
				}
			}
			imported++
		}
		return nil
	})
	return imported, err
}

// isTimeColumn 判断列是否时间列：列名含 time/date，或类型本身是时间类型
// （PostgreSQL 里常见 created_at/updated_at 命名，不能只看列名）
func isTimeColumn(c ColumnMeta) bool {
	lower := strings.ToLower(c.Name)
	if strings.Contains(lower, "time") || strings.Contains(lower, "date") {
		return true
	}
	t := strings.ToUpper(strings.TrimSpace(c.Type))
	if i := strings.Index(t, "("); i > 0 {
		t = t[:i]
	}
	t = strings.TrimSpace(t)
	return strings.HasPrefix(t, "TIMESTAMP") || t == "DATE" || t == "TIME" ||
		strings.HasSuffix(t, "TIME ZONE")
}

// buildColumn 根据列元信息生成默认配置
func buildColumn(c ColumnMeta, tableID uint) model.GenColumn {
	goType := dbTypeToGoType(c.Type, c.Name)
	lower := strings.ToLower(c.Name)
	comment := strings.TrimSpace(c.Comment)
	if comment == "" {
		comment = c.Name // 无注释时回退列名，避免生成代码里出现空注释
	}
	col := model.GenColumn{
		TableID:       tableID,
		ColumnName:    c.Name,
		ColumnComment: comment,
		ColumnType:    c.Type,
		GoType:        goType,
		IsPk:          boolToInt(c.PK > 0),
		Sort:          c.CID,
	}
	// 主键：整数主键视为自增
	if c.PK > 0 {
		col.IsIncrement = boolToInt(goType == "int64")
		col.IsRequired = 0
		col.IsList = 1
		col.IsQuery = 0
		col.IsForm = 0
		return col
	}
	// 布尔列：列表展示 + 开关表单；不参与查询过滤（false 无法表达"不过滤"）
	if goType == "bool" {
		col.IsList = 1
		col.IsQuery = 0
		col.IsForm = 1
		col.FormType = "switch"
		return col
	}
	// 时间字段：默认表单 date；不做等值/模糊查询（时间范围由列表接口单独处理）
	if isTimeColumn(c) {
		col.IsList = 1
		col.IsQuery = 0
		col.IsForm = 1
		col.FormType = "date"
		return col
	}
	col.IsList = 1
	// 字符串列默认 LIKE 查询，数字列默认 EQ
	if goType == "string" {
		col.QueryType = "LIKE"
	} else {
		col.QueryType = "EQ"
	}
	col.IsQuery = 1
	col.IsForm = 1
	col.FormType = "input"
	// 长文本列（PostgreSQL text / MySQL text 系列）用多行输入框
	if strings.Contains(strings.ToUpper(c.Type), "TEXT") {
		col.FormType = "textarea"
	}
	col.IsRequired = boolToInt(lower == "status" || lower == "sort" || lower == "order_num" || lower == "type" || lower == "name" || lower == "code")
	return col
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ListGenTables 分页查询已导入的生成表
func ListGenTables(q struct {
	PageNum, PageSize    int
	TableName, ClassName string
}) ([]model.GenTable, int64, error) {
	var tables []model.GenTable
	var total int64
	pageNum, pageSize := normPage(q.PageNum, q.PageSize)
	db := DB.Model(&model.GenTable{})
	if q.TableName != "" {
		db = db.Where(likeExprEsc("table_name"), "%"+escapeLike(q.TableName)+"%")
	}
	if q.ClassName != "" {
		db = db.Where(likeExprEsc("class_name"), "%"+escapeLike(q.ClassName)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("table_id asc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&tables).Error
	return tables, total, err
}

// GetGenTableByID 查询生成表配置（含列）
func GetGenTableByID(id uint) (*model.GenTable, error) {
	var table model.GenTable
	if err := DB.First(&table, id).Error; err != nil {
		return nil, err
	}
	var cols []model.GenColumn
	if err := DB.Where("table_id = ?", id).Order("sort asc").Find(&cols).Error; err != nil {
		return nil, err
	}
	table.Columns = cols
	return &table, nil
}

// GenColumnReq 列配置请求体（前端字段）
type GenColumnReq struct {
	ColumnID      uint   `json:"columnId"`
	TableID       uint   `json:"tableId"`
	ColumnName    string `json:"columnName"`
	ColumnComment string `json:"columnComment"`
	ColumnType    string `json:"columnType"`
	GoType        string `json:"goType"`
	IsPk          int    `json:"isPk"`
	IsIncrement   int    `json:"isIncrement"`
	IsRequired    int    `json:"isRequired"`
	IsList        int    `json:"isList"`
	IsQuery       int    `json:"isQuery"`
	QueryType     string `json:"queryType"`
	IsForm        int    `json:"isForm"`
	FormType      string `json:"formType"`
	Sort          int    `json:"sort"`
}

// SaveGenTable 保存表配置（表字段 + 列配置全量替换）
func SaveGenTable(table *model.GenTable, cols []GenColumnReq) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		table.UpdateTime = model.LocalTime(time.Now())
		if err := tx.Save(table).Error; err != nil {
			return err
		}
		// 删除旧列后重建
		if err := tx.Where("table_id = ?", table.TableID).Delete(&model.GenColumn{}).Error; err != nil {
			return err
		}
		for _, col := range cols {
			col.TableID = table.TableID
			item := model.GenColumn{
				TableID:       col.TableID,
				ColumnName:    col.ColumnName,
				ColumnComment: col.ColumnComment,
				ColumnType:    col.ColumnType,
				GoType:        col.GoType,
				IsPk:          col.IsPk,
				IsIncrement:   col.IsIncrement,
				IsRequired:    col.IsRequired,
				IsList:        col.IsList,
				IsQuery:       col.IsQuery,
				QueryType:     col.QueryType,
				IsForm:        col.IsForm,
				FormType:      col.FormType,
				Sort:          col.Sort,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteGenTable 删除生成表配置（连带列配置）
func DeleteGenTable(id uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("table_id = ?", id).Delete(&model.GenColumn{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.GenTable{}, id).Error
	})
}

// ---------------------------------------------------------------------------
// 代码生成数据组装
// ---------------------------------------------------------------------------

// GenField 模板字段
type GenField struct {
	GoField     string // UserName
	JsonField   string // userName
	ColumnName  string // user_name
	ColumnType  string // 数据库原始类型（TEXT / INTEGER ...）
	GoType      string // int64 / string / float64
	Comment     string
	IsPk        bool
	IsIncrement bool
	IsRequired  bool
	IsList      bool
	IsQuery     bool
	QueryType   string // EQ / NE / GT / LT / LIKE
	IsForm      bool
	FormType    string // input / textarea / select / date
	Sort        int
	// 派生
	GormTag      string // gorm 标签片段
	QueryTag     string // 查询条件注释
	TsType       string // TypeScript 类型（number/string/boolean）
	IsCreateTime bool   // create_time / created_at
	IsUpdateTime bool   // update_time / updated_at
}

// GenData 模板数据
type GenData struct {
	PackageName  string // system
	ModuleName   string // system（前端模块）
	ClassName    string // User
	ClassVar     string // user（小驼峰）
	ClassSnake   string // user（蛇形）——Go 文件命名
	TableName    string // sys_user
	TableComment string
	FunctionName string // 用户
	Author       string
	IDField      string // 主键 GoField（如 Id）
	IDJsonField  string // 主键 json（userId）
	IDGoType     string // 主键 Go 类型
	HasTime      bool   // 是否有 create/update 时间列
	HasAutoTime  bool   // 是否存在 create_time/update_time 类列（模板需要 time 包）
	TplCategory  string // crud / tree / sub
	Fields       []GenField

	// 树形结构
	IsTree              bool
	TreeCode            string // 编码列（数据库列名）
	TreeCodeLabel       string // 编码列注释
	TreeName            string // 名称列
	TreeNameLabel       string // 名称列注释
	TreeNameJsonField   string // 名称列 json 字段（树选择器 label）
	TreeCodeJsonField   string // 编码列 json 字段
	TreeParent          string // 父级列
	TreeParentGoField   string
	TreeParentJsonField string
	TreeParentGoType    string

	// 主子表
	IsSub           bool
	SubTableName    string
	SubTableComment string
	SubClassName    string
	SubClassVar     string
	SubClassSnake   string
	SubFkColumn     string // 子表外键列
	SubFkGoField    string
	SubFkJsonField  string
	SubIDField      string
	SubIDJsonField  string
	SubIDGoType     string
	SubFields       []GenField
}

// makeGenField 把列配置转成模板字段
func makeGenField(c model.GenColumn) GenField {
	f := GenField{
		GoField:     snakeToPascal(c.ColumnName),
		JsonField:   snakeToCamel(c.ColumnName),
		ColumnName:  c.ColumnName,
		ColumnType:  c.ColumnType,
		GoType:      c.GoType,
		Comment:     c.ColumnComment,
		IsPk:        c.IsPk == 1,
		IsIncrement: c.IsIncrement == 1,
		IsRequired:  c.IsRequired == 1,
		IsList:      c.IsList == 1,
		IsQuery:     c.IsQuery == 1,
		QueryType:   c.QueryType,
		IsForm:      c.IsForm == 1,
		FormType:    c.FormType,
		Sort:        c.Sort,
	}
	tags := []string{}
	if f.IsPk {
		tags = append(tags, "primaryKey")
	}
	if c.ColumnType != "" {
		tags = append(tags, fmt.Sprintf("column:%s", c.ColumnName))
	}
	lowerName := strings.ToLower(c.ColumnName)
	switch lowerName {
	case "create_time", "created_at":
		f.IsCreateTime = true
	case "update_time", "updated_at":
		f.IsUpdateTime = true
	}
	f.GormTag = "`gorm:\"" + strings.Join(tags, ";") + "\" json:\"" + f.JsonField + "\"`"
	switch {
	case strings.HasPrefix(f.GoType, "int"), strings.HasPrefix(f.GoType, "uint"), f.GoType == "float64":
		f.TsType = "number"
	case f.GoType == "bool":
		f.TsType = "boolean"
	default: // string / time.Time / 其他
		f.TsType = "string"
	}
	return f
}

// fieldExprs 用于模板计算的中间产物
type fieldExpr struct {
	col model.GenColumn
	f   GenField
}

// BuildGenData 组装代码生成模板数据
func BuildGenData(tableID uint) (*GenData, error) {
	table, err := GetGenTableByID(tableID)
	if err != nil {
		return nil, err
	}
	if table.ClassName == "" {
		table.ClassName = tableToClassName(table.TblName)
	}
	category := strings.ToLower(strings.TrimSpace(table.TplCategory))
	if category == "" {
		category = "crud"
	}
	if category != "crud" && category != "tree" && category != "sub" {
		return nil, fmt.Errorf("不支持的生成类型: %s（可选 crud/tree/sub）", table.TplCategory)
	}
	d := &GenData{
		PackageName:  table.PackageName,
		ModuleName:   table.ModuleName,
		ClassName:    table.ClassName,
		ClassVar:     snakeToCamel(table.ClassName),
		ClassSnake:   toSnakeCase(table.ClassName),
		TableName:    table.TblName,
		TableComment: table.TableComment,
		FunctionName: table.FunctionName,
		Author:       table.Author,
		TplCategory:  category,
	}
	for _, c := range table.Columns {
		f := makeGenField(c)
		d.Fields = append(d.Fields, f)
		if f.IsPk {
			d.IDField = f.GoField
			d.IDJsonField = f.JsonField
			d.IDGoType = f.GoType
		}
		// 时间列检测
		lower := strings.ToLower(f.ColumnName)
		if strings.Contains(lower, "create_time") || strings.Contains(lower, "update_time") ||
			strings.Contains(lower, "created_at") || strings.Contains(lower, "updated_at") {
			d.HasTime = true
		}
		if f.IsCreateTime || f.IsUpdateTime {
			d.HasAutoTime = true
		}
	}
	if d.IDField == "" {
		return nil, fmt.Errorf("表 %s 未配置主键", table.TblName)
	}

	// 树形结构：父级/名称/编码列
	if category == "tree" {
		d.IsTree = true
		parentCol := strings.TrimSpace(table.TreeParent)
		if parentCol == "" {
			parentCol = "parent_id"
		}
		var parentField *GenField
		for i := range d.Fields {
			if d.Fields[i].ColumnName == parentCol {
				parentField = &d.Fields[i]
				break
			}
		}
		if parentField == nil {
			return nil, fmt.Errorf("树形表 %s 未找到父级列 %s", table.TblName, parentCol)
		}
		d.TreeParent = parentField.ColumnName
		d.TreeParentGoField = parentField.GoField
		d.TreeParentJsonField = parentField.JsonField
		d.TreeParentGoType = parentField.GoType
		// 名称列：优先配置，其次取第一个非主键字符串列
		nameCol := strings.TrimSpace(table.TreeName)
		if nameCol == "" {
			for i := range d.Fields {
				f := d.Fields[i]
				if !f.IsPk && f.GoType == "string" {
					nameCol = f.ColumnName
					break
				}
			}
		}
		if nameCol == "" {
			return nil, fmt.Errorf("树形表 %s 未找到可用的名称列", table.TblName)
		}
		nameField := findGenField(d.Fields, nameCol)
		if nameField == nil {
			return nil, fmt.Errorf("树形表 %s 未找到名称列 %s", table.TblName, nameCol)
		}
		d.TreeName = nameField.ColumnName
		d.TreeNameLabel = nameField.Comment
		d.TreeNameJsonField = nameField.JsonField
		// 编码列：优先配置，其次 name 列，最后名称列本身
		codeCol := strings.TrimSpace(table.TreeCode)
		if codeCol == "" {
			codeCol = nameCol
		}
		codeField := findGenField(d.Fields, codeCol)
		if codeField == nil {
			return nil, fmt.Errorf("树形表 %s 未找到编码列 %s", table.TblName, codeCol)
		}
		d.TreeCode = codeField.ColumnName
		d.TreeCodeLabel = codeField.Comment
		d.TreeCodeJsonField = codeField.JsonField
		// 构树要求主键与父级字段类型一致（整型）
		if !isIntGoType(d.IDGoType) || d.TreeParentGoType != d.IDGoType {
			return nil, fmt.Errorf("树形表 %s 的主键(%s)与父级列(%s)类型需一致且为整型", table.TblName, d.IDGoType, d.TreeParentGoType)
		}
	}

	// 主子表：读取子表列配置
	if category == "sub" {
		d.IsSub = true
		subTable := strings.TrimSpace(table.SubTableName)
		if subTable == "" {
			return nil, fmt.Errorf("主子表 %s 未配置子表", table.TblName)
		}
		fkCol := strings.TrimSpace(table.SubTableFk)
		if fkCol == "" {
			return nil, fmt.Errorf("主子表 %s 未配置子表外键列", table.TblName)
		}
		d.SubTableName = subTable
		d.SubTableComment = tableComment(subTable)
		if d.SubTableComment == "" {
			d.SubTableComment = subTable
		}
		d.SubClassName = tableToClassName(subTable)
		d.SubClassVar = snakeToCamel(d.SubClassName)
		d.SubClassSnake = toSnakeCase(d.SubClassName)
		cols, err := tableColumns(subTable)
		if err != nil {
			return nil, fmt.Errorf("读取子表 %s 列失败: %v", subTable, err)
		}
		if len(cols) == 0 {
			return nil, fmt.Errorf("子表 %s 不存在或没有列", subTable)
		}
		fkFound := false
		for _, c := range cols {
			col := buildColumn(c, table.TableID)
			col.ColumnComment = strings.TrimSpace(c.Comment)
			f := makeGenField(col)
			// 子表外键列不进入表单与列表（由主表维护）
			if f.ColumnName == fkCol {
				fkFound = true
				d.SubFkColumn = f.ColumnName
				d.SubFkGoField = f.GoField
				d.SubFkJsonField = f.JsonField
				f.IsForm = false
				f.IsQuery = false
				f.IsList = false
			}
			if f.IsPk {
				d.SubIDField = f.GoField
				d.SubIDJsonField = f.JsonField
				d.SubIDGoType = f.GoType
				f.IsList = false
				f.IsForm = false
			}
			d.SubFields = append(d.SubFields, f)
		}
		if !fkFound {
			return nil, fmt.Errorf("子表 %s 未找到外键列 %s", subTable, fkCol)
		}
		if d.SubIDField == "" {
			return nil, fmt.Errorf("子表 %s 未配置主键", subTable)
		}
		// 子表外键与主表主键类型必须一致，否则生成代码无法直接赋值
		var fkGen *GenField
		for i := range d.SubFields {
			if d.SubFields[i].ColumnName == fkCol {
				fkGen = &d.SubFields[i]
				break
			}
		}
		if fkGen == nil || fkGen.GoType != d.IDGoType {
			fkGo := ""
			if fkGen != nil {
				fkGo = fkGen.GoType
			}
			return nil, fmt.Errorf("子表外键 %s 类型(%s)需与主表主键类型(%s)一致", fkCol, fkGo, d.IDGoType)
		}
	}
	return d, nil
}

// isIntGoType 判断是否整型 Go 类型
func isIntGoType(t string) bool {
	switch t {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	}
	return false
}

// findGenField 按列名查找模板字段
func findGenField(fields []GenField, column string) *GenField {
	for i := range fields {
		if fields[i].ColumnName == column {
			return &fields[i]
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// 模板渲染与打包
// ---------------------------------------------------------------------------

// genTemplatesDir 模板目录（相对 exe 所在目录）
func genTemplatesDir() string {
	return "templates"
}

// readTemplateFile 定位模板文件：优先可执行文件目录，其次工作目录（与 config.yaml 双兜底一致）
func readTemplateFile(path string) (string, error) {
	candidates := []string{}
	if dir, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(dir), path))
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, path))
		// go test 在 dao 包目录下执行，模板位于上一级
		candidates = append(candidates, filepath.Join(wd, "..", path))
	}
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("模板文件不存在: %s", path)
}

func newTemplate(name string) *template.Template {
	return template.New(name).Funcs(template.FuncMap{
		"lower": strings.ToLower,
		// modelType 生成 model 包内文件时去掉包名前缀（model.LocalTime -> LocalTime）
		"modelType": func(s string) string { return strings.TrimPrefix(s, "model.") },
	})
}

// renderTemplate 渲染单个模板
func renderTemplate(tmplName string, data *GenData) (string, error) {
	path := genTemplatesDir() + "/" + tmplName
	content, err := readTemplateFile(path)
	if err != nil {
		return "", err
	}
	return renderTemplateString(tmplName, content, data)
}

func renderTemplateString(name, content string, data *GenData) (string, error) {
	tmpl, err := newTemplate(name).Parse(content)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// renderTemplateMap 渲染一组模板，返回 map[文件名]内容
func renderTemplateMap(data *GenData) (map[string]string, error) {
	files := []string{
		"model.tmpl", "dao.tmpl", "api.tmpl", "router.tmpl",
		"vue_api.tmpl", "readme.tmpl",
	}
	switch data.TplCategory {
	case "tree":
		files = append(files, "vue_index_tree.tmpl")
	case "sub":
		files = append(files, "vue_index_sub.tmpl")
	default:
		files = append(files, "vue_index.tmpl")
	}
	out := make(map[string]string, len(files))
	for _, f := range files {
		s, err := renderTemplate(f, data)
		if err != nil {
			return nil, fmt.Errorf("渲染模板 %s 失败: %v", f, err)
		}
		out[f] = s
	}
	return out, nil
}

// GenZipFiles 生成 zip 下载文件列表（文件路径 -> 内容）
func GenZipFiles(data *GenData) (map[string]string, error) {
	rendered, err := renderTemplateMap(data)
	if err != nil {
		return nil, err
	}
	base := data.ModuleName + "/" + data.ClassVar
	indexTpl := "vue_index.tmpl"
	switch data.TplCategory {
	case "tree":
		indexTpl = "vue_index_tree.tmpl"
	case "sub":
		indexTpl = "vue_index_sub.tmpl"
	}
	files := map[string]string{
		"README.txt": rendered["readme.tmpl"],
		"backend/model/" + data.ClassSnake + ".go":   rendered["model.tmpl"],
		"backend/dao/" + data.ClassSnake + "_dao.go": rendered["dao.tmpl"],
		"backend/api/" + data.ClassSnake + "_api.go": rendered["api.tmpl"],
		"backend/router.go.txt":                      rendered["router.tmpl"],
		"frontend/src/api/" + data.ClassVar + ".ts":  rendered["vue_api.tmpl"],
		"frontend/src/views/" + base + "/index.vue":  rendered[indexTpl],
	}
	return files, nil
}
