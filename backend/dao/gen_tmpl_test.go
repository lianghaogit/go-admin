package dao

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// tmplField 构造模板字段
func tmplField(goField, jsonField, column, goType, comment string, opts ...string) GenField {
	f := GenField{
		GoField:    goField,
		JsonField:  jsonField,
		ColumnName: column,
		ColumnType: "text",
		GoType:     goType,
		Comment:    comment,
		IsList:     true,
		TsType:     "string",
	}
	if goType == "int64" || goType == "float64" {
		f.TsType = "number"
	}
	f.GormTag = "`gorm:\"column:" + column + "\" json:\"" + jsonField + "\"`"
	for _, o := range opts {
		switch o {
		case "pk":
			f.IsPk = true
			f.GormTag = "`gorm:\"primaryKey;column:" + column + "\" json:\"" + jsonField + "\"`"
		case "increment":
			f.IsIncrement = true
		case "form":
			f.IsForm = true
			f.FormType = "input"
		case "query":
			f.IsQuery = true
			f.QueryType = "LIKE"
		case "create":
			f.IsCreateTime = true
			f.TsType = "string"
		}
	}
	return f
}

func baseGenData(category, class string) *GenData {
	d := &GenData{
		PackageName:  "system",
		ModuleName:   "system",
		ClassName:    class,
		ClassVar:     strings.ToLower(class),
		ClassSnake:   strings.ToLower(class),
		TableName:    "demo_" + strings.ToLower(class),
		TableComment: class + "演示表",
		FunctionName: class,
		Author:       "test",
		IDField:      "Id",
		IDJsonField:  "id",
		IDGoType:     "int64",
		TplCategory:  category,
		Fields: []GenField{
			tmplField("Id", "id", "id", "int64", "主键", "pk", "increment"),
			tmplField("Name", "name", "name", "string", "名称", "form", "query"),
			tmplField("Status", "status", "status", "int64", "状态", "form", "query"),
		},
	}
	return d
}

func TestGenTemplatesContainersCrud(t *testing.T) {
	d := baseGenData("crud", "Book")
	files, err := GenZipFiles(d)
	if err != nil {
		t.Fatalf("crud 渲染失败: %v", err)
	}
	assertGoFilesParse(t, files)
	if !strings.Contains(files["frontend/src/views/system/book/index.vue"], "el-pagination") {
		t.Error("单表页面缺少分页")
	}
	// 数值字段必须用 el-input-number，否则页面提交字符串会让后端 JSON 反序列化失败
	if !strings.Contains(files["frontend/src/views/system/book/index.vue"], `<el-input-number v-model="form.status"`) {
		t.Error("单表页面的数值字段未使用 el-input-number")
	}
	if !strings.Contains(files["backend/api/book_api.go"], "genreg.RegisterModule") {
		t.Error("api 模板缺少自注册")
	}
	api := files["backend/api/book_api.go"]
	if !strings.Contains(api, "ParentID:  1700") {
		t.Error("生成菜单未挂到系统工具(1700)")
	}
	for _, want := range []string{"ExportBook", "ImportBook", `Path: "/export"`, `Path: "/importTemplate"`, `Path: "/importData"`, "SyncImportSequence"} {
		if !strings.Contains(api, want) {
			t.Errorf("api 模板缺少导入导出：%s", want)
		}
	}
	dao := files["backend/dao/book_dao.go"]
	for _, want := range []string{"ListBookAll", "func ImportBook("} {
		if !strings.Contains(dao, want) {
			t.Errorf("dao 模板缺少导入导出：%s", want)
		}
	}
	vue := files["frontend/src/views/system/book/index.vue"]
	for _, want := range []string{"导入", "导出", "ExcelImportDialog", "importBookTemplate", "exportBook"} {
		if !strings.Contains(vue, want) {
			t.Errorf("单表页面缺少导入导出 UI：%s", want)
		}
	}
}

func TestGenTemplatesSkipsSequenceForNonIncrementPK(t *testing.T) {
	d := baseGenData("crud", "Doc")
	d.Fields[0] = tmplField("Id", "id", "id", "string", "主键", "pk")
	files, err := GenZipFiles(d)
	if err != nil {
		t.Fatalf("非自增主键渲染失败: %v", err)
	}
	if strings.Contains(files["backend/api/doc_api.go"], "SyncImportSequence") {
		t.Error("非自增主键不应生成 PG 序列同步调用")
	}
}

func TestGenTemplatesTree(t *testing.T) {
	d := baseGenData("tree", "Category")
	d.IsTree = true
	d.TreeCode = "name"
	d.TreeCodeLabel = "名称"
	d.TreeCodeJsonField = "name"
	d.TreeName = "name"
	d.TreeNameLabel = "名称"
	d.TreeNameJsonField = "name"
	d.TreeParent = "parent_id"
	d.TreeParentGoField = "ParentId"
	d.TreeParentJsonField = "parentId"
	d.TreeParentGoType = "int64"
	d.Fields = append(d.Fields, tmplField("ParentId", "parentId", "parent_id", "int64", "父级ID", "form"))
	files, err := GenZipFiles(d)
	if err != nil {
		t.Fatalf("tree 渲染失败: %v", err)
	}
	assertGoFilesParse(t, files)
	dao := files["backend/dao/category_dao.go"]
	if !strings.Contains(dao, "buildCategoryTree") || !strings.Contains(dao, "存在下级节点") {
		t.Error("树形 dao 缺少构树或删除校验")
	}
	if !strings.Contains(dao, "上级节点不能是自身或其下级") || !strings.Contains(dao, "depth > 50") {
		t.Error("树形 dao 缺少父级成环校验或递归深度保护")
	}
	vue := files["frontend/src/views/system/category/index.vue"]
	if !strings.Contains(vue, "tree-props") || !strings.Contains(vue, "el-tree-select") {
		t.Error("树形页面缺少 tree-props 或上级选择器")
	}
	if !strings.Contains(vue, `<el-input-number v-model="form.status"`) {
		t.Error("树形页面的数值字段未使用 el-input-number")
	}
	if !strings.Contains(vue, "ExcelImportDialog") || !strings.Contains(vue, "exportCategory") {
		t.Error("树形页面缺少导入导出 UI")
	}
}

func TestGenTemplatesSub(t *testing.T) {
	d := baseGenData("sub", "Order")
	d.IsSub = true
	d.SubTableName = "demo_order_item"
	d.SubTableComment = "订单明细"
	d.SubClassName = "DemoOrderItem"
	d.SubClassVar = "demoOrderItem"
	d.SubClassSnake = "demo_order_item"
	d.SubFkColumn = "order_id"
	d.SubFkGoField = "OrderId"
	d.SubFkJsonField = "orderId"
	d.SubIDField = "Id"
	d.SubIDJsonField = "id"
	d.SubIDGoType = "int64"
	d.SubFields = []GenField{
		tmplField("Id", "id", "id", "int64", "主键", "pk"),
		tmplField("OrderId", "orderId", "order_id", "int64", "订单ID"),
		tmplField("ProductName", "productName", "product_name", "string", "商品名称", "form"),
		tmplField("Qty", "qty", "qty", "int64", "数量", "form"),
	}
	files, err := GenZipFiles(d)
	if err != nil {
		t.Fatalf("sub 渲染失败: %v", err)
	}
	assertGoFilesParse(t, files)
	dao := files["backend/dao/order_dao.go"]
	if !strings.Contains(dao, "DB.Transaction") || !strings.Contains(dao, "OrderId: item.Id") {
		t.Error("主子表 dao 缺少事务写入或外键赋值")
	}
	if !strings.Contains(dao, "replaceChildren ...bool") || !strings.Contains(dao, "len(r.Children) > 0") {
		t.Error("主子表导入更新缺少“未带明细保留原明细”保护")
	}
	model := files["backend/model/order.go"]
	if !strings.Contains(model, "type DemoOrderItem struct") || !strings.Contains(model, "Children []DemoOrderItem") {
		t.Error("主子表模型缺少子表结构或 children 字段")
	}
	vue := files["frontend/src/views/system/order/index.vue"]
	if !strings.Contains(vue, "添加明细") || !strings.Contains(vue, "children") {
		t.Error("主子表页面缺少子表编辑区")
	}
	if !strings.Contains(vue, `<el-input-number v-model="form.status"`) {
		t.Error("主子表页面主表数值字段未使用 el-input-number")
	}
	if !strings.Contains(vue, "ExcelImportDialog") || !strings.Contains(vue, "exportOrder") {
		t.Error("主子表页面缺少导入导出 UI")
	}
	if !strings.Contains(files["backend/api/order_api.go"], "所属主表行号") {
		t.Error("主子表导入缺少明细关联列说明")
	}
}

func TestGenTemplatesTimeFieldsPreserved(t *testing.T) {
	d := baseGenData("crud", "Book")
	d.HasAutoTime = true
	d.Fields = append(d.Fields, tmplField("CreateTime", "createTime", "create_time", "model.LocalTime", "创建时间", "create"))
	files, err := GenZipFiles(d)
	if err != nil {
		t.Fatalf("crud 时间列渲染失败: %v", err)
	}
	assertGoFilesParse(t, files)
	dao := files["backend/dao/book_dao.go"]
	if !strings.Contains(dao, `"time"`) || !strings.Contains(dao, "model.LocalTime(time.Now())") {
		t.Error("时间列缺少创建时间补全")
	}
	if !strings.Contains(dao, "updates := map[string]interface{}{") || strings.Contains(dao, "DB.Save(") {
		t.Error("修改应使用 Updates 映射，避免覆盖 create_time")
	}
	if strings.Contains(dao, `"create_time": r.CreateTime`) {
		t.Error("修改映射不应包含 create_time")
	}
}

// assertGoFilesParse 校验生成的后端 Go 文件能被标准库解析
func assertGoFilesParse(t *testing.T, files map[string]string) {
	t.Helper()
	for name, content := range files {
		if !strings.HasPrefix(name, "backend/") || !strings.HasSuffix(name, ".go") {
			continue
		}
		if _, err := parser.ParseFile(token.NewFileSet(), name, content, parser.AllErrors); err != nil {
			t.Errorf("%s 语法错误: %v\n%s", name, err, content)
		}
	}
}
