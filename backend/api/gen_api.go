package api

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"

	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListDbTables 查询未导入的数据库表
func ListDbTables(c *gin.Context) {
	tables, err := dao.ListDbTables(c.Query("tableName"))
	if err != nil {
		utils.Fail(c, 500, "查询数据库表失败")
		return
	}
	utils.OkData(c, tables)
}

// ImportGenTables 导入数据库表（tables=表名1,表名2）
func ImportGenTables(c *gin.Context) {
	names := strings.Split(c.Query("tables"), ",")
	n, err := dao.ImportTables(names)
	if err != nil {
		utils.Fail(c, 500, "导入失败: "+err.Error())
		return
	}
	if n == 0 {
		utils.Fail(c, 500, "没有可导入的表（已导入或非业务表会被跳过）")
		return
	}
	utils.OkMsg(c, fmt.Sprintf("成功导入 %d 张表", n))
}

// ListGenTables 分页查询已导入的生成表
func ListGenTables(c *gin.Context) {
	q := struct {
		PageNum, PageSize int
		TableName         string
		ClassName         string
	}{PageNum: pageNum(c), PageSize: pageSize(c), TableName: c.Query("tableName"), ClassName: c.Query("className")}
	tables, total, err := dao.ListGenTables(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, tables, total)
}

// GetGenTable 查询生成表详情（含列配置）
func GetGenTable(c *gin.Context) {
	table, err := dao.GetGenTableByID(paramUint(c, "tableId"))
	if err != nil {
		utils.Fail(c, 500, "生成表不存在")
		return
	}
	utils.OkData(c, table)
}

// GenTableReq 保存请求体
type GenTableReq struct {
	TableID      uint               `json:"tableId"`
	TableName    string             `json:"tableName"`
	TableComment string             `json:"tableComment"`
	ClassName    string             `json:"className"`
	ModuleName   string             `json:"moduleName"`
	PackageName  string             `json:"packageName"`
	FunctionName string             `json:"functionName"`
	Author       string             `json:"author"`
	Remark       string             `json:"remark"`
	TplCategory  string             `json:"tplCategory"`
	TreeCode     string             `json:"treeCode"`
	TreeName     string             `json:"treeName"`
	TreeParent   string             `json:"treeParent"`
	SubTableName string             `json:"subTableName"`
	SubTableFk   string             `json:"subTableFk"`
	Columns      []dao.GenColumnReq `json:"columns"`
}

// SaveGenTable 保存表配置（表信息 + 列配置）
func SaveGenTable(c *gin.Context) {
	var req GenTableReq
	if err := c.ShouldBindJSON(&req); err != nil || req.TableID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetGenTableByID(req.TableID)
	if err != nil {
		utils.Fail(c, 500, "生成表不存在")
		return
	}
	old.TableComment = req.TableComment
	old.ClassName = req.ClassName
	old.ModuleName = req.ModuleName
	old.PackageName = req.PackageName
	old.FunctionName = req.FunctionName
	old.Author = req.Author
	old.Remark = req.Remark
	old.TplCategory = req.TplCategory
	old.TreeCode = req.TreeCode
	old.TreeName = req.TreeName
	old.TreeParent = req.TreeParent
	old.SubTableName = req.SubTableName
	old.SubTableFk = req.SubTableFk
	cols := make([]dao.GenColumnReq, 0, len(req.Columns))
	for _, col := range req.Columns {
		col.TableID = old.TableID
		cols = append(cols, col)
	}
	if err := dao.SaveGenTable(old, cols); err != nil {
		utils.Fail(c, 500, "保存失败: "+err.Error())
		return
	}
	utils.Ok(c)
}

// DeleteGenTable 删除生成表配置
func DeleteGenTable(c *gin.Context) {
	if err := dao.DeleteGenTable(paramUint(c, "tableId")); err != nil {
		utils.Fail(c, 500, "删除失败")
		return
	}
	utils.Ok(c)
}

// PreviewGenTable 预览生成代码
func PreviewGenTable(c *gin.Context) {
	data, err := dao.BuildGenData(paramUint(c, "tableId"))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	files, err := dao.GenZipFiles(data)
	if err != nil {
		utils.Fail(c, 500, "生成预览失败: "+err.Error())
		return
	}
	utils.OkData(c, files)
}

// DownloadGenCode 下载生成代码（zip）
func DownloadGenCode(c *gin.Context) {
	data, err := dao.BuildGenData(paramUint(c, "tableId"))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	files, err := dao.GenZipFiles(data)
	if err != nil {
		utils.Fail(c, 500, "生成代码失败: "+err.Error())
		return
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := zw.Create(name)
		if err != nil {
			utils.Fail(c, 500, "打包失败")
			return
		}
		if _, err := f.Write([]byte(content)); err != nil {
			utils.Fail(c, 500, "打包失败")
			return
		}
	}
	if err := zw.Close(); err != nil {
		utils.Fail(c, 500, "打包失败")
		return
	}
	filename := data.ClassVar + ".zip"
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/zip")
	c.Data(200, "application/zip", buf.Bytes())
}
