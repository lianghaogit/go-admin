package api

import (
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListNotices 分页查询通知公告
func ListNotices(c *gin.Context) {
	q := dao.NoticeQuery{
		PageNum:  pageNum(c),
		PageSize: pageSize(c),
		Title:    c.Query("noticeTitle"),
		Type:     c.Query("noticeType"),
		Status:   c.Query("status"),
	}
	notices, total, err := dao.ListNotices(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, notices, total)
}

// ListActiveNotices 查询正常状态公告（首页展示）
func ListActiveNotices(c *gin.Context) {
	notices, err := dao.ListActiveNotices(5)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkData(c, notices)
}

// GetNotice 查询公告详情
func GetNotice(c *gin.Context) {
	n, err := dao.GetNoticeByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "公告不存在")
		return
	}
	utils.OkData(c, n)
}

// NoticeReq 公告请求体
type NoticeReq struct {
	NoticeID uint   `json:"noticeId"`
	Title    string `json:"noticeTitle"`
	Type     string `json:"noticeType"`
	Content  string `json:"noticeContent"`
	Status   string `json:"status"`
	Remark   string `json:"remark"`
}

// CreateNotice 新增公告
func CreateNotice(c *gin.Context) {
	var req NoticeReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Title == "" || req.Content == "" {
		utils.Fail(c, 500, "公告标题和内容不能为空")
		return
	}
	n := &model.Notice{
		Title: req.Title, Type: req.Type, Content: req.Content,
		Status: req.Status, Remark: req.Remark, CreateBy: c.GetString("username"),
	}
	if n.Type == "" {
		n.Type = "1"
	}
	if n.Status == "" {
		n.Status = "0"
	}
	if err := dao.CreateNotice(n); err != nil {
		utils.Fail(c, 500, "新增公告失败")
		return
	}
	utils.Ok(c)
}

// UpdateNotice 修改公告
func UpdateNotice(c *gin.Context) {
	var req NoticeReq
	if err := c.ShouldBindJSON(&req); err != nil || req.NoticeID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetNoticeByID(req.NoticeID)
	if err != nil {
		utils.Fail(c, 500, "公告不存在")
		return
	}
	n := &model.Notice{
		ID: old.ID, Title: req.Title, Type: req.Type, Content: req.Content,
		Status: req.Status, Remark: req.Remark, CreateBy: old.CreateBy,
		CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	if n.Type == "" {
		n.Type = old.Type
	}
	if n.Status == "" {
		n.Status = old.Status
	}
	if err := dao.UpdateNotice(n); err != nil {
		utils.Fail(c, 500, "修改公告失败")
		return
	}
	utils.Ok(c)
}

// DeleteNotice 删除公告
func DeleteNotice(c *gin.Context) {
	if err := dao.DeleteNotice(idParam(c)); err != nil {
		utils.Fail(c, 500, "删除公告失败")
		return
	}
	utils.Ok(c)
}
