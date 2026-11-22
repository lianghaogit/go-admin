package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

var userExportHeaders = []string{"用户编号", "用户名", "用户昵称", "部门", "手机号码", "邮箱", "状态", "备注"}

var operLogExportHeaders = []string{"编号", "模块标题", "操作类型", "请求方式", "操作人员", "操作IP", "操作地址", "状态", "耗时(ms)", "操作时间"}
var loginLogExportHeaders = []string{"编号", "用户名", "登录IP", "状态", "消息", "登录时间"}

// ExportOperLogs 导出操作日志（xlsx，支持搜索条件与时间范围，不分页）
func ExportOperLogs(c *gin.Context) {
	q := dao.OperLogQuery{
		PageNum:   1,
		PageSize:  100000,
		Title:     c.Query("title"),
		OperName:  c.Query("operName"),
		Status:    c.Query("status"),
		BeginTime: c.Query("beginTime"),
		EndTime:   c.Query("endTime"),
	}
	logs, _, err := dao.ListOperLogs(q)
	if err != nil {
		utils.Fail(c, 500, "导出失败")
		return
	}
	bizLabels := map[int]string{0: "其它", 1: "新增", 2: "修改", 3: "删除", 4: "查询"}
	rows := make([][]interface{}, 0, len(logs))
	for _, l := range logs {
		status := "成功"
		if l.Status == "1" {
			status = "失败"
		}
		rows = append(rows, []interface{}{
			l.ID, l.Title, bizLabels[l.BusinessType], l.RequestMethod, l.OperName,
			l.OperIP, l.OperURL, status, l.CostTime, l.OperTime.Format("2006-01-02 15:04:05"),
		})
	}
	data, err := utils.BuildExcel("操作日志", operLogExportHeaders, rows)
	if err != nil {
		utils.Fail(c, 500, "生成文件失败")
		return
	}
	filename := fmt.Sprintf("操作日志_%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ExportLoginLogs 导出登录日志（xlsx，支持过滤条件与时间范围，不分页）
func ExportLoginLogs(c *gin.Context) {
	q := dao.LoginLogQuery{
		PageNum:   1,
		PageSize:  100000,
		Username:  c.Query("username"),
		IP:        c.Query("ipaddr"),
		Status:    c.Query("status"),
		BeginTime: c.Query("beginTime"),
		EndTime:   c.Query("endTime"),
	}
	logs, _, err := dao.ListLoginLogs(q)
	if err != nil {
		utils.Fail(c, 500, "导出失败")
		return
	}
	rows := make([][]interface{}, 0, len(logs))
	for _, l := range logs {
		status := "成功"
		if l.Status == "1" {
			status = "失败"
		}
		rows = append(rows, []interface{}{
			l.ID, l.Username, l.IP, status, l.Msg, l.LoginTime.Format("2006-01-02 15:04:05"),
		})
	}
	data, err := utils.BuildExcel("登录日志", loginLogExportHeaders, rows)
	if err != nil {
		utils.Fail(c, 500, "生成文件失败")
		return
	}
	filename := fmt.Sprintf("登录日志_%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ExportUsers 导出用户列表（xlsx，数据权限过滤，不分页）
func ExportUsers(c *gin.Context) {
	q := dao.UserQuery{
		Username: c.Query("username"),
		Phone:    c.Query("phone"),
		Status:   c.Query("status"),
	}
	scopeSQL, scopeArgs, _ := dao.DataScopeUserCond(c.GetUint("userId"))
	users, err := dao.ListUsersAll(q, scopeSQL, scopeArgs)
	if err != nil {
		utils.Fail(c, 500, "导出失败")
		return
	}
	rows := make([][]interface{}, 0, len(users))
	for _, u := range users {
		deptName := ""
		if u.Dept != nil {
			deptName = u.Dept.DeptName
		}
		status := "正常"
		if u.Status == "1" {
			status = "停用"
		}
		rows = append(rows, []interface{}{
			u.ID, u.Username, u.Nickname, deptName, u.Phone, u.Email, status, u.Remark,
		})
	}
	data, err := utils.BuildExcel("用户数据", userExportHeaders, rows)
	if err != nil {
		utils.Fail(c, 500, "生成文件失败")
		return
	}
	filename := fmt.Sprintf("用户数据_%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ImportUserTemplate 下载用户导入模板
func ImportUserTemplate(c *gin.Context) {
	example := [][]interface{}{{
		1, "zhangsan", "张三", "研发部", "13800001111", "zhangsan@example.com", "正常", "示例：导入后请删除此行",
	}}
	data, err := utils.BuildExcel("用户数据", userExportHeaders, example)
	if err != nil {
		utils.Fail(c, 500, "生成模板失败")
		return
	}
	filename := fmt.Sprintf("用户导入模板_%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// 导入限制：防止超大文件耗尽内存
const (
	maxImportSize = 5 << 20 // 5MB
	maxImportRows = 5000
)

// ImportUsers 导入用户（xlsx：用户名/昵称/部门/手机/邮箱/状态/备注，updateSupport=true 时用户名已存在则更新）
func ImportUsers(c *gin.Context) {
	updateSupport := c.PostForm("updateSupport") == "true"

	file, err := c.FormFile("file")
	if err != nil {
		utils.Fail(c, 500, "请选择要上传的文件")
		return
	}
	src, err := file.Open()
	if err != nil {
		utils.Fail(c, 500, "文件读取失败")
		return
	}
	defer src.Close()
	if file.Size > maxImportSize {
		utils.Fail(c, 500, "文件过大，最大支持 5MB")
		return
	}

	f, err := excelize.OpenReader(src)
	if err != nil {
		utils.Fail(c, 500, "文件解析失败，请上传 xlsx 格式文件")
		return
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	rows, err := f.GetRows(sheet)
	if err != nil || len(rows) < 2 {
		utils.Fail(c, 500, "文件中没有数据")
		return
	}
	if len(rows)-1 > maxImportRows {
		utils.Fail(c, 500, "数据行数超过上限（最多 5000 行）")
		return
	}

	successNum := 0
	failNum := 0
	failMessages := make([]string, 0)

	phoneRe := regexp.MustCompile(`^1[3-9]\d{9}$`)
	for i, row := range rows[1:] { // 跳过表头
		line := i + 2
		get := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		username := get(1)
		nickname := get(2)
		deptName := get(3)
		phone := get(4)
		email := get(5)
		statusText := get(6)
		remark := get(7)

		fail := func(msg string) {
			failNum++
			failMessages = append(failMessages, fmt.Sprintf("第 %d 行：%s", line, msg))
		}

		if username == "" {
			fail("用户名不能为空")
			continue
		}
		if nickname == "" {
			fail("用户昵称不能为空")
			continue
		}
		if phone != "" && !phoneRe.MatchString(phone) {
			fail("手机号码格式不正确")
			continue
		}
		status := "0"
		if statusText != "" {
			if statusText == "停用" || statusText == "1" {
				status = "1"
			} else if statusText == "正常" || statusText == "0" {
				status = "0"
			} else {
				fail("状态只能为 正常/停用")
				continue
			}
		}

		// 部门按名称解析（不存在则留空归属）
		deptID := uint(0)
		if deptName != "" {
			dept, err := dao.GetDeptByName(deptName)
			if err != nil {
				fail("部门不存在：" + deptName)
				continue
			}
			deptID = dept.ID
		}

		// 用户名唯一性
		existing, err := dao.GetUserByUsername(username)
		if err == nil && existing.ID > 0 {
			if !updateSupport {
				fail("用户名已存在（如需覆盖导入请勾选“更新支持”）")
				continue
			}
			if existing.Username == "admin" {
				fail("不允许通过导入修改超级管理员")
				continue
			}
			if !dao.UserInDataScope(c.GetUint("userId"), existing.ID) {
				fail("无权修改该用户（超出数据权限范围）")
				continue
			}
			if err := dao.UpdateUserImport(existing.ID, nickname, email, phone, status, remark, deptID); err != nil {
				fail("更新失败：" + err.Error())
				continue
			}
			successNum++
			continue
		}

		// 新增：默认密码 sys.user.initPassword（缺省 admin123）
		initPwd := dao.GetConfigValue("sys.user.initPassword")
		if initPwd == "" {
			initPwd = "admin123"
		}
		hash, err := utils.HashPassword(initPwd)
		if err != nil {
			fail("密码加密失败")
			continue
		}
		user := &model.User{
			Username: username, Password: hash, Nickname: nickname,
			Email: email, Phone: phone, Status: status, Remark: remark, DeptID: deptID,
		}
		if err := dao.CreateUser(user, nil); err != nil {
			fail("新增失败：" + err.Error())
			continue
		}
		successNum++
	}

	utils.OkData(c, gin.H{
		"successNum":   successNum,
		"failureNum":   failNum,
		"failMessages": failMessages,
	})
}
