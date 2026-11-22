package api

import (
	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// flowOperator 当前操作人（完整用户对象，含部门/昵称）
func flowOperator(c *gin.Context) *model.User {
	userID := c.GetUint("userId")
	u, err := dao.GetUserByID(userID)
	if err != nil {
		return &model.User{ID: userID, Username: c.GetString("username"), Nickname: c.GetString("username")}
	}
	return u
}

// ---------- 流程定义 ----------

// ListFlowDefs 流程定义列表
func ListFlowDefs(c *gin.Context) {
	defs, total, err := dao.ListFlowDefs(pageNum(c), pageSize(c), c.Query("keyword"), c.Query("status"), c.Query("category"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, defs, total)
}

// ListEnabledFlowDefs 启用中的流程（发起下拉）
func ListEnabledFlowDefs(c *gin.Context) {
	defs, err := dao.ListEnabledFlowDefs()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkData(c, defs)
}

// GetFlowDef 流程定义详情
func GetFlowDef(c *gin.Context) {
	def, err := dao.GetFlowDef(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "流程不存在")
		return
	}
	utils.OkData(c, def)
}

// CreateFlowDef 新增流程定义
func CreateFlowDef(c *gin.Context) {
	var req dao.FlowDefInput
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	id, err := dao.SaveFlowDef(req, c.GetString("username"))
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.OkData(c, gin.H{"flowDefId": id})
}

// UpdateFlowDef 更新流程定义
func UpdateFlowDef(c *gin.Context) {
	var req dao.FlowDefInput
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if req.FlowDefID == 0 {
		utils.Fail(c, 400, "缺少流程 ID")
		return
	}
	_, err := dao.SaveFlowDef(req, c.GetString("username"))
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// DeleteFlowDef 删除流程定义
func DeleteFlowDef(c *gin.Context) {
	if err := dao.DeleteFlowDef(idParam(c)); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// FlowStatusReq 启停请求
type FlowStatusReq struct {
	FlowDefID uint   `json:"flowDefId"`
	Status    string `json:"status"`
}

// ChangeFlowDefStatus 流程启停
func ChangeFlowDefStatus(c *gin.Context) {
	var req FlowStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if err := dao.ChangeFlowDefStatus(req.FlowDefID, req.Status); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// ---------- 发起与我的申请 ----------

// ListFlowApply 我的申请
func ListFlowApply(c *gin.Context) {
	list, total, err := dao.ListFlowApply(c.GetUint("userId"), pageNum(c), pageSize(c), c.Query("keyword"), c.Query("status"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, list, total)
}

// ListFlowInstances 全部实例（管理员）
func ListFlowInstances(c *gin.Context) {
	list, total, err := dao.ListFlowInstances(pageNum(c), pageSize(c), c.Query("keyword"), c.Query("status"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, list, total)
}

// StartFlow 发起申请
func StartFlow(c *gin.Context) {
	var req dao.FlowStartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	id, err := dao.StartFlowInstance(*flowOperator(c), req)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.OkData(c, gin.H{"flowInstanceId": id})
}

// WithdrawFlow 撤回申请
func WithdrawFlow(c *gin.Context) {
	if err := dao.WithdrawFlowInstance(idParam(c), c.GetUint("userId")); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// ReapplyReq 重新提交请求
type ReapplyReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ReapplyFlow 驳回后重新提交
func ReapplyFlow(c *gin.Context) {
	var req ReapplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if err := dao.ReapplyFlowInstance(idParam(c), c.GetUint("userId"), req.Title, req.Content); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// UrgeFlow 催办
func UrgeFlow(c *gin.Context) {
	if err := dao.UrgeFlowInstance(idParam(c), c.GetUint("userId")); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// FlowTerminateReq 终止请求
type FlowTerminateReq struct {
	FlowInstanceID uint   `json:"flowInstanceId"`
	Comment        string `json:"comment"`
}

// TerminateFlow 管理员终止（仅运行中）
func TerminateFlow(c *gin.Context) {
	req := FlowTerminateReq{FlowInstanceID: idParam(c)}
	_ = c.ShouldBindJSON(&req)
	if err := dao.TerminateFlowInstance(req.FlowInstanceID, c.GetUint("userId"), req.Comment); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// ---------- 待办/已办/抄送 ----------

// ListFlowTodo 我的待办
func ListFlowTodo(c *gin.Context) {
	list, total, err := dao.ListFlowTodo(c.GetUint("userId"), pageNum(c), pageSize(c), c.Query("keyword"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, list, total)
}

// ListFlowDone 我处理的
func ListFlowDone(c *gin.Context) {
	list, total, err := dao.ListFlowDone(c.GetUint("userId"), pageNum(c), pageSize(c), c.Query("keyword"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, list, total)
}

// ListFlowCc 抄送我的
func ListFlowCc(c *gin.Context) {
	list, total, err := dao.ListFlowCc(c.GetUint("userId"), pageNum(c), pageSize(c), c.Query("keyword"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, list, total)
}

// FlowTaskActionReq 任务动作请求（通过/驳回/转办/加签）
type FlowTaskActionReq struct {
	TaskID   uint   `json:"taskId"`
	Comment  string `json:"comment"`
	ToUserID uint   `json:"toUserId"`
}

// ApproveFlowTask 审批通过
func ApproveFlowTask(c *gin.Context) {
	var req FlowTaskActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if err := dao.ApproveFlowTask(req.TaskID, c.GetUint("userId"), req.Comment); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// RejectFlowTask 驳回
func RejectFlowTask(c *gin.Context) {
	var req FlowTaskActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if err := dao.RejectFlowTask(req.TaskID, c.GetUint("userId"), req.Comment); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// TransferFlowTask 转办
func TransferFlowTask(c *gin.Context) {
	var req FlowTaskActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if req.ToUserID == 0 {
		utils.Fail(c, 400, "请选择被转办人")
		return
	}
	if err := dao.TransferFlowTask(req.TaskID, c.GetUint("userId"), req.ToUserID, req.Comment); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// AddSignFlowTask 加签
func AddSignFlowTask(c *gin.Context) {
	var req FlowTaskActionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if req.ToUserID == 0 {
		utils.Fail(c, 400, "请选择被加签人")
		return
	}
	if err := dao.AddSignFlowTask(req.TaskID, c.GetUint("userId"), req.ToUserID, req.Comment); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// MarkCcRead 抄送已读
func MarkCcRead(c *gin.Context) {
	if err := dao.MarkCcRead(idParam(c), c.GetUint("userId")); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Ok(c)
}

// GetFlowInstanceDetail 实例详情
func GetFlowInstanceDetail(c *gin.Context) {
	detail, err := dao.GetFlowInstanceDetail(idParam(c), c.GetUint("userId"))
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.OkData(c, detail)
}
