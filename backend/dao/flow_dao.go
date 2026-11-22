package dao

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-admin/model"

	"gorm.io/gorm"
)

// ========== 工作流/审批引擎（自研状态机） ==========
// 状态：实例 1运行中 2已通过 3已驳回 4已撤销 5已终止；任务 0待办 1通过 2驳回 3转出 4加签 5撤销 6已读
// 节点模式：one 或签（任一通过即过） / any 会签（全部通过才过）
// 审批人类型：role 按角色 / user 指定人 / dept_leader 发起人部门负责人（按部门 leader 字段匹配用户昵称）

// ---------- 查询辅助 ----------

func flowNodesOf(defID uint) []model.FlowNode {
	var nodes []model.FlowNode
	_ = DB.Where("flow_def_id = ?", defID).Order("node_order ASC, flow_node_id ASC").Find(&nodes).Error
	return nodes
}

func flowNodeByID(id uint) (model.FlowNode, error) {
	var n model.FlowNode
	if err := DB.First(&n, id).Error; err != nil {
		return n, err
	}
	return n, nil
}

func flowDefByID(id uint) (model.FlowDef, error) {
	var d model.FlowDef
	if err := DB.First(&d, id).Error; err != nil {
		return d, err
	}
	return d, nil
}

func flowUserByID(id uint) (model.User, error) {
	var u model.User
	if err := DB.First(&u, id).Error; err != nil {
		return u, err
	}
	return u, nil
}

// parseFlowIDs 解析逗号分隔的 ID 字符串
func parseFlowIDs(s string) []uint {
	var ids []uint
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.ParseUint(p, 10, 64); err == nil {
			ids = append(ids, uint(v))
		}
	}
	return ids
}

func joinFlowIDs(ids []uint) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatUint(uint64(id), 10))
	}
	return strings.Join(parts, ",")
}

// ---------- 审批人解析 ----------

// flowApproverUsers 解析节点审批人：返回 (用户ID集合, 姓名Map, 错误)
func flowApproverUsers(node model.FlowNode, applicant model.User) ([]model.User, error) {
	var users []model.User
	switch node.ApproverType {
	case "role":
		roleIDs := parseFlowIDs(node.ApproverIDs)
		if len(roleIDs) == 0 {
			return nil, fmt.Errorf("节点「%s」未配置角色", node.NodeName)
		}
		if err := DB.Raw(`SELECT DISTINCT u.* FROM sys_user u
			JOIN sys_user_role ur ON u.id = ur.user_id
			WHERE ur.role_id IN ? AND u.status = '0'`, roleIDs).Scan(&users).Error; err != nil {
			return nil, err
		}
	case "user":
		userIDs := parseFlowIDs(node.ApproverIDs)
		for _, uid := range userIDs {
			var u model.User
			if err := DB.Where("id = ? AND status = '0'", uid).First(&u).Error; err == nil {
				users = append(users, u)
			}
		}
	case "dept_leader":
		// 发起人部门负责人：按 sys_dept.leader（姓名）匹配用户昵称/用户名
		var leader, deptName string
		_ = DB.Table("sys_dept").Where("id = ?", applicant.DeptID).Pluck("leader", &leader).Error
		_ = DB.Table("sys_dept").Where("id = ?", applicant.DeptID).Pluck("dept_name", &deptName).Error
		if deptName == "" {
			deptName = "申请人部门"
		}
		if strings.TrimSpace(leader) == "" {
			return nil, fmt.Errorf("部门负责人路由失败：部门「%s」未配置负责人（leader）", deptName)
		}
		if err := DB.Where("(nickname = ? OR username = ?) AND status = '0'", leader, leader).Find(&users).Error; err != nil {
			return nil, err
		}
		if len(users) == 0 {
			return nil, fmt.Errorf("部门负责人路由失败：部门「%s」的负责人「%s」未匹配到启用用户", deptName, leader)
		}
		return users, nil
	default:
		return nil, fmt.Errorf("节点「%s」审批人类型未知：%s", node.NodeName, node.ApproverType)
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("节点「%s」无可用的审批人", node.NodeName)
	}
	return users, nil
}

// flowNodeTimeoutAt 计算任务超时时间（节点优先，其次流程定义；0=不限）
func flowNodeTimeoutAt(node model.FlowNode, def model.FlowDef, created time.Time) *time.Time {
	hours := node.TimeoutHours
	if hours <= 0 {
		hours = def.TimeoutHours
	}
	if hours <= 0 {
		return nil
	}
	t := created.Add(time.Duration(hours) * time.Hour)
	return &t
}

// ---------- 记录 ----------

func flowAddRecord(db *gorm.DB, instanceID, nodeID uint, nodeName, action, operatorName, targetName, comment string, operatorID uint) error {
	r := model.FlowRecord{
		InstanceID:   instanceID,
		NodeID:       nodeID,
		NodeName:     nodeName,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Action:       action,
		TargetName:   targetName,
		Comment:      comment,
	}
	return db.Create(&r).Error
}

// flowCreateNodeTasks 为节点生成审批任务（或签/会签均为每人一条，完成判断在 approve 中区分）
func flowCreateNodeTasks(tx *gorm.DB, instance model.FlowInstance, node model.FlowNode, def model.FlowDef, applicant model.User) error {
	// 部门快照：后续节点（推进/退回/重新提交）的 dept_leader 路由以发起时部门为准，
	// 避免申请人中途换部门导致审批路由漂移
	if instance.ApplicantDeptID > 0 {
		applicant.DeptID = instance.ApplicantDeptID
	}
	users, err := flowApproverUsers(node, applicant)
	if err != nil {
		return err
	}
	overdue := flowNodeTimeoutAt(node, def, time.Now())
	for _, u := range users {
		t := model.FlowTask{
			InstanceID:   instance.ID,
			NodeID:       node.ID,
			NodeName:     node.NodeName,
			NodeOrder:    node.NodeOrder,
			TaskType:     "1",
			AssigneeID:   u.ID,
			AssigneeName: u.Nickname,
			Status:       "0",
			OverdueAt:    overdue,
		}
		if err := tx.Create(&t).Error; err != nil {
			return err
		}
	}
	return nil
}

// flowFinishInstance 实例收尾（通过/驳回/撤销/终止）：清待办、置抄送已读、写完成时间
func flowFinishInstance(tx *gorm.DB, instanceID uint, status string) error {
	// 只撤销审批类待办（1 审批 / 3 加签），抄送任务（2）另行置已读
	if err := tx.Model(&model.FlowTask{}).
		Where("instance_id = ? AND task_type <> '2' AND status = '0'", instanceID).
		Update("status", "5").Error; err != nil {
		return err
	}
	if err := tx.Model(&model.FlowTask{}).
		Where("instance_id = ? AND task_type = '2' AND status = '0'", instanceID).
		Update("status", "6").Error; err != nil { // 未读抄送置已读
		return err
	}
	now := time.Now()
	if err := tx.Model(&model.FlowInstance{}).Where("flow_instance_id = ?", instanceID).
		Updates(map[string]interface{}{"status": status, "finished_at": now, "current_node_id": 0, "current_order": 0}).Error; err != nil {
		return err
	}
	return nil
}

// flowAdvance 推进到下一节点（当前节点已全部完成）：无下一节点则实例通过
func flowAdvance(tx *gorm.DB, instance model.FlowInstance, node model.FlowNode, def model.FlowDef) error {
	nodes := flowNodesOf(def.ID)
	var next *model.FlowNode
	for i := range nodes {
		if nodes[i].NodeOrder > node.NodeOrder {
			next = &nodes[i]
			break
		}
	}
	if next == nil {
		return flowFinishInstance(tx, instance.ID, "2")
	}
	// 幂等：下一节点已有审批任务说明并发请求已推进过，直接返回，避免重复生成任务
	var exists int64
	if err := tx.Model(&model.FlowTask{}).
		Where("instance_id = ? AND node_id = ? AND task_type IN ('1','3') AND status = '0'", instance.ID, next.ID).
		Count(&exists).Error; err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	// 更新实例当前节点，并生成下一节点任务（审批人路由需申请人的部门快照）
	var applicant model.User
	_ = DB.First(&applicant, instance.ApplicantID)
	if err := flowCreateNodeTasks(tx, instance, *next, def, applicant); err != nil {
		return err
	}
	return tx.Model(&model.FlowInstance{}).Where("flow_instance_id = ?", instance.ID).
		Updates(map[string]interface{}{"current_node_id": next.ID, "current_order": next.NodeOrder}).Error
}

// flowRejectFinish 驳回处理：reject_to=1 且存在上一节点则退回上一节点重新审批，否则驳回回发起人
func flowRejectFinish(tx *gorm.DB, instance model.FlowInstance, node model.FlowNode, def model.FlowDef, operatorName string, operatorID uint) error {
	if node.RejectTo == "1" && node.NodeOrder > 1 {
		nodes := flowNodesOf(def.ID)
		var prev *model.FlowNode
		for i := range nodes {
			if nodes[i].NodeOrder < node.NodeOrder {
				prev = &nodes[i]
			}
		}
		if prev != nil {
			var applicant model.User
			_ = DB.First(&applicant, instance.ApplicantID)
			if err := flowCreateNodeTasks(tx, instance, *prev, def, applicant); err != nil {
				return err
			}
			// 时间线标记「退回」：退回上一节点重新审批
			if err := flowAddRecord(tx, instance.ID, prev.ID, prev.NodeName, "12", operatorName, "", "退回至「"+prev.NodeName+"」重新审批", operatorID); err != nil {
				return err
			}
			return tx.Model(&model.FlowInstance{}).Where("flow_instance_id = ?", instance.ID).
				Updates(map[string]interface{}{"current_node_id": prev.ID, "current_order": prev.NodeOrder, "status": "1"}).Error
		}
	}
	// 回发起人
	if err := flowFinishInstance(tx, instance.ID, "3"); err != nil {
		return err
	}
	return tx.Model(&model.FlowInstance{}).Where("flow_instance_id = ?", instance.ID).
		Update("reject_from_node_id", node.ID).Error
}

// flowResolvePendingTarget 重新提交恢复节点：优先驳回来源节点，否则第一节点
func flowResolvePendingTarget(tx *gorm.DB, instanceID, fallbackNodeID uint) (model.FlowNode, error) {
	if fallbackNodeID > 0 {
		n, err := flowNodeByID(fallbackNodeID)
		if err == nil {
			return n, nil
		}
	}
	var def model.FlowDef
	if err := DB.Table("sys_flow_def").Where("flow_def_id = ?",
		DB.Table("sys_flow_instance").Select("flow_def_id").Where("flow_instance_id = ?", instanceID)).First(&def).Error; err != nil {
		return model.FlowNode{}, err
	}
	nodes := flowNodesOf(def.ID)
	if len(nodes) == 0 {
		return model.FlowNode{}, fmt.Errorf("流程未配置节点")
	}
	return nodes[0], nil
}

// ---------- 发起与审批动作 ----------

// FlowStartReq 发起申请请求
type FlowStartReq struct {
	FlowDefID uint   `json:"flowDefId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CCUserIDs []uint `json:"ccUserIds"`
}

// StartFlowInstance 发起流程实例
func StartFlowInstance(operator model.User, req FlowStartReq) (uint, error) {
	if req.FlowDefID == 0 || strings.TrimSpace(req.Title) == "" {
		return 0, fmt.Errorf("请选择流程并填写标题")
	}
	def, err := flowDefByID(req.FlowDefID)
	if err != nil {
		return 0, fmt.Errorf("流程不存在")
	}
	if def.Status != "0" {
		return 0, fmt.Errorf("流程已停用，无法发起")
	}
	nodes := flowNodesOf(def.ID)
	if len(nodes) == 0 {
		return 0, fmt.Errorf("流程「%s」未配置审批节点", def.DefName)
	}
	now := time.Now()
	inst := model.FlowInstance{
		FlowDefID:       def.ID,
		DefName:         def.DefName,
		Title:           req.Title,
		Content:         req.Content,
		ApplicantID:     operator.ID,
		ApplicantName:   operator.Nickname,
		ApplicantDeptID: operator.DeptID,
		CCUserIDs:       joinFlowIDs(req.CCUserIDs),
		Status:          "1",
		CurrentNodeID:   nodes[0].ID,
		CurrentOrder:    nodes[0].NodeOrder,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&inst).Error; err != nil {
			return err
		}
		if err := flowCreateNodeTasks(tx, inst, nodes[0], def, operator); err != nil {
			return err
		}
		// 抄送任务（排除申请人本人 + 去重）
		seen := make(map[uint]bool)
		for _, uid := range req.CCUserIDs {
			if uid == 0 || uid == operator.ID || seen[uid] {
				continue
			}
			seen[uid] = true
			u, err := flowUserByID(uid)
			if err != nil {
				continue
			}
			t := model.FlowTask{
				InstanceID:   inst.ID,
				NodeID:       nodes[0].ID,
				NodeName:     nodes[0].NodeName,
				NodeOrder:    nodes[0].NodeOrder,
				TaskType:     "2",
				AssigneeID:   u.ID,
				AssigneeName: u.Nickname,
				Status:       "0",
			}
			if err := tx.Create(&t).Error; err != nil {
				return err
			}
			if err := flowAddRecord(tx, inst.ID, nodes[0].ID, nodes[0].NodeName, "11", operator.Nickname, u.Nickname, "抄送", operator.ID); err != nil {
				return err
			}
		}
		if err := flowAddRecord(tx, inst.ID, 0, "发起", "1", operator.Nickname, "", "发起申请", operator.ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return inst.ID, nil
}

// ApproveFlowTask 审批通过（或签：任一过即节点完成；会签：全部通过才完成）
func ApproveFlowTask(taskID, operatorID uint, comment string) error {
	var task model.FlowTask
	if err := DB.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("任务不存在")
	}
	if task.Status != "0" {
		return fmt.Errorf("该任务已被处理")
	}
	var operator model.User
	if err := DB.First(&operator, operatorID).Error; err != nil {
		return fmt.Errorf("操作人不存在")
	}
	if task.AssigneeID != operatorID {
		return fmt.Errorf("无权处理该任务")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		// 乐观锁：仅待办可处理，防并发重复审批
		res := tx.Model(&model.FlowTask{}).Where("flow_task_id = ? AND status = '0'", task.ID).
			Updates(map[string]interface{}{"status": "1", "comment": comment, "passed_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("该任务已被处理")
		}
		node, err := flowNodeByID(task.NodeID)
		if err != nil {
			return err
		}
		var inst model.FlowInstance
		if err := tx.First(&inst, task.InstanceID).Error; err != nil {
			return err
		}
		def, _ := flowDefByID(inst.FlowDefID)
		if err := flowAddRecord(tx, inst.ID, node.ID, node.NodeName, "2", operator.Nickname, "", comment, operatorID); err != nil {
			return err
		}
		// 完成判断：同节点同实例的审批类任务（含加签）是否全部通过
		var pending int64
		if err := tx.Model(&model.FlowTask{}).
			Where("instance_id = ? AND node_id = ? AND task_type IN ('1','3') AND status = '0'", inst.ID, node.ID).
			Count(&pending).Error; err != nil {
			return err
		}
		if node.NodeMode == "any" && pending > 0 {
			return nil // 会签：仍有未通过任务，等待
		}
		// 或签：其余审批待办撤销（status=5），抄送任务不参与；会签已全过也走此分支（pending=0）
		if node.NodeMode == "one" {
			if err := tx.Model(&model.FlowTask{}).
				Where("instance_id = ? AND node_id = ? AND task_type <> '2' AND status = '0'", inst.ID, node.ID).
				Update("status", "5").Error; err != nil {
				return err
			}
		}
		return flowAdvance(tx, inst, node, def)
	})
}

// RejectFlowTask 驳回（reject_to=1 退回上一节点；否则驳回回发起人）
func RejectFlowTask(taskID, operatorID uint, comment string) error {
	var task model.FlowTask
	if err := DB.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("任务不存在")
	}
	if task.Status != "0" {
		return fmt.Errorf("该任务已被处理")
	}
	var operator model.User
	if err := DB.First(&operator, operatorID).Error; err != nil {
		return fmt.Errorf("操作人不存在")
	}
	if task.AssigneeID != operatorID {
		return fmt.Errorf("无权处理该任务")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&model.FlowTask{}).Where("flow_task_id = ? AND status = '0'", task.ID).
			Updates(map[string]interface{}{"status": "2", "comment": comment, "passed_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("该任务已被处理")
		}
		node, err := flowNodeByID(task.NodeID)
		if err != nil {
			return err
		}
		var inst model.FlowInstance
		if err := tx.First(&inst, task.InstanceID).Error; err != nil {
			return err
		}
		def, _ := flowDefByID(inst.FlowDefID)
		if err := flowAddRecord(tx, inst.ID, node.ID, node.NodeName, "3", operator.Nickname, "", comment, operatorID); err != nil {
			return err
		}
		// 撤销同节点其余审批待办（抄送任务保持未读/已读状态）
		if err := tx.Model(&model.FlowTask{}).
			Where("instance_id = ? AND node_id = ? AND task_type <> '2' AND status = '0'", inst.ID, node.ID).
			Update("status", "5").Error; err != nil {
			return err
		}
		return flowRejectFinish(tx, inst, node, def, operator.Nickname, operatorID)
	})
}

// TransferFlowTask 转办：原任务置转出，为被转办人新建同节点审批任务
func TransferFlowTask(taskID, operatorID, toUserID uint, comment string) error {
	var task model.FlowTask
	if err := DB.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("任务不存在")
	}
	if task.Status != "0" {
		return fmt.Errorf("该任务已被处理")
	}
	var operator model.User
	if err := DB.First(&operator, operatorID).Error; err != nil {
		return fmt.Errorf("操作人不存在")
	}
	if task.AssigneeID != operatorID {
		return fmt.Errorf("无权处理该任务")
	}
	to, err := flowUserByID(toUserID)
	if err != nil {
		return fmt.Errorf("被转办人不存在")
	}
	if to.ID == operatorID {
		return fmt.Errorf("不能转办给自己")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		// 防重复指派：目标人在当前节点已有待办任务（含加签），拒绝转办
		var dup int64
		_ = tx.Model(&model.FlowTask{}).
			Where("instance_id = ? AND node_id = ? AND assignee_id = ? AND status = '0'", task.InstanceID, task.NodeID, to.ID).
			Count(&dup).Error
		if dup > 0 {
			return fmt.Errorf("该用户已是当前节点审批人")
		}
		res := tx.Model(&model.FlowTask{}).Where("flow_task_id = ? AND status = '0'", task.ID).
			Update("status", "3")
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("该任务已被处理")
		}
		nt := model.FlowTask{
			InstanceID:   task.InstanceID,
			NodeID:       task.NodeID,
			NodeName:     task.NodeName,
			NodeOrder:    task.NodeOrder,
			TaskType:     "1",
			AssigneeID:   to.ID,
			AssigneeName: to.Nickname,
			Status:       "0",
			OverdueAt:    task.OverdueAt,
		}
		if err := tx.Create(&nt).Error; err != nil {
			return err
		}
		if err := flowAddRecord(tx, task.InstanceID, task.NodeID, task.NodeName, "4", operator.Nickname, to.Nickname, comment, operatorID); err != nil {
			return err
		}
		return nil
	})
}

// AddSignFlowTask 加签：原任务保留，为加签人新建并行任务（参与节点完成判断）
func AddSignFlowTask(taskID, operatorID, toUserID uint, comment string) error {
	var task model.FlowTask
	if err := DB.First(&task, taskID).Error; err != nil {
		return fmt.Errorf("任务不存在")
	}
	if task.Status != "0" {
		return fmt.Errorf("该任务已被处理")
	}
	node, err := flowNodeByID(task.NodeID)
	if err != nil {
		return err
	}
	if node.CanAddSign != "1" {
		return fmt.Errorf("当前节点不允许加签")
	}
	var operator model.User
	if err := DB.First(&operator, operatorID).Error; err != nil {
		return fmt.Errorf("操作人不存在")
	}
	if task.AssigneeID != operatorID {
		return fmt.Errorf("无权处理该任务")
	}
	to, err := flowUserByID(toUserID)
	if err != nil {
		return fmt.Errorf("被加签人不存在")
	}
	if to.ID == operatorID {
		return fmt.Errorf("不能加签给自己")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		// 防重复加签：目标人已是当前节点待办审批人，拒绝再加签
		var dup int64
		_ = tx.Model(&model.FlowTask{}).
			Where("instance_id = ? AND node_id = ? AND assignee_id = ? AND status = '0'", task.InstanceID, task.NodeID, to.ID).
			Count(&dup).Error
		if dup > 0 {
			return fmt.Errorf("该用户已是当前节点审批人")
		}
		nt := model.FlowTask{
			InstanceID:   task.InstanceID,
			NodeID:       task.NodeID,
			NodeName:     task.NodeName,
			NodeOrder:    task.NodeOrder,
			TaskType:     "3",
			AssigneeID:   to.ID,
			AssigneeName: to.Nickname,
			Status:       "0",
			OverdueAt:    task.OverdueAt,
		}
		if err := tx.Create(&nt).Error; err != nil {
			return err
		}
		if err := flowAddRecord(tx, task.InstanceID, node.ID, node.NodeName, "5", operator.Nickname, to.Nickname, comment, operatorID); err != nil {
			return err
		}
		return nil
	})
}

// WithdrawFlowInstance 申请人撤回（仅运行中且本人）
func WithdrawFlowInstance(instanceID, applicantID uint) error {
	var inst model.FlowInstance
	if err := DB.First(&inst, instanceID).Error; err != nil {
		return fmt.Errorf("申请不存在")
	}
	if inst.ApplicantID != applicantID {
		return fmt.Errorf("仅申请人本人可撤回")
	}
	if inst.Status != "1" {
		return fmt.Errorf("当前状态不可撤回")
	}
	// 撤回语义：一旦本轮审批人处理过（通过/驳回/转办/加签），申请人不能再撤回
	// 驳回后重新提交（action=10）会开启新一轮，仅需检查该轮之后的记录
	var acted int64
	actedQ := DB.Model(&model.FlowRecord{}).Where("instance_id = ? AND action IN ('2','3','4','5')", instanceID)
	var lastReapply model.FlowRecord
	if err := DB.Where("instance_id = ? AND action = '10'", instanceID).
		Order("flow_record_id DESC").First(&lastReapply).Error; err == nil {
		actedQ = actedQ.Where("flow_record_id > ?", lastReapply.ID)
	}
	if err := actedQ.Count(&acted).Error; err != nil {
		return err
	}
	if acted > 0 {
		return fmt.Errorf("已有审批人处理，无法撤回")
	}
	var applicant model.User
	_ = DB.First(&applicant, applicantID)
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := flowFinishInstance(tx, instanceID, "4"); err != nil {
			return err
		}
		return flowAddRecord(tx, instanceID, inst.CurrentNodeID, "", "6", applicant.Nickname, "", "撤回申请", applicantID)
	})
}

// ReapplyFlowInstance 驳回后重新提交：从驳回来源节点（或第一节点）重新生成任务
func ReapplyFlowInstance(instanceID, applicantID uint, title, content string) error {
	var inst model.FlowInstance
	if err := DB.First(&inst, instanceID).Error; err != nil {
		return fmt.Errorf("申请不存在")
	}
	if inst.ApplicantID != applicantID {
		return fmt.Errorf("仅申请人本人可重新提交")
	}
	if inst.Status != "3" {
		return fmt.Errorf("仅已驳回的申请可重新提交")
	}
	var applicant model.User
	_ = DB.First(&applicant, applicantID)
	def, err := flowDefByID(inst.FlowDefID)
	if err != nil {
		return err
	}
	if def.Status != "0" {
		return fmt.Errorf("流程已停用，无法重新提交")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{"title": title, "content": content, "status": "1", "reject_from_node_id": 0, "updated_at": time.Now()}
		if err := tx.Model(&model.FlowInstance{}).Where("flow_instance_id = ?", instanceID).Updates(updates).Error; err != nil {
			return err
		}
		node, err := flowResolvePendingTarget(tx, instanceID, inst.RejectFromNodeID)
		if err != nil {
			return err
		}
		if err := flowCreateNodeTasks(tx, inst, node, def, applicant); err != nil {
			return err
		}
		if err := tx.Model(&model.FlowInstance{}).Where("flow_instance_id = ?", instanceID).
			Updates(map[string]interface{}{"current_node_id": node.ID, "current_order": node.NodeOrder}).Error; err != nil {
			return err
		}
		return flowAddRecord(tx, instanceID, node.ID, node.NodeName, "10", applicant.Nickname, "", "重新提交", applicantID)
	})
}

// TerminateFlowInstance 管理员终止（仅运行中）
func TerminateFlowInstance(instanceID, operatorID uint, comment string) error {
	var inst model.FlowInstance
	if err := DB.First(&inst, instanceID).Error; err != nil {
		return fmt.Errorf("申请不存在")
	}
	if inst.Status != "1" {
		return fmt.Errorf("仅运行中的申请可终止")
	}
	var operator model.User
	_ = DB.First(&operator, operatorID)
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := flowFinishInstance(tx, instanceID, "5"); err != nil {
			return err
		}
		return flowAddRecord(tx, instanceID, inst.CurrentNodeID, "", "8", operator.Nickname, "", comment, operatorID)
	})
}

// UrgeFlowInstance 催办：写入催办记录（overdue watcher 自动催办一次）
func UrgeFlowInstance(instanceID, operatorID uint) error {
	var inst model.FlowInstance
	if err := DB.First(&inst, instanceID).Error; err != nil {
		return fmt.Errorf("申请不存在")
	}
	if inst.Status != "1" {
		return fmt.Errorf("仅运行中的申请可催办")
	}
	var operator model.User
	_ = DB.First(&operator, operatorID)
	// 权限：仅申请人本人 / 当前节点审批人（待办任务 assignee）/ 管理员可催办
	if !IsAdmin(operatorID) && inst.ApplicantID != operatorID {
		var onNode int64
		_ = DB.Model(&model.FlowTask{}).
			Where("instance_id = ? AND node_id = ? AND assignee_id = ? AND status = '0'", instanceID, inst.CurrentNodeID, operatorID).
			Count(&onNode).Error
		if onNode == 0 {
			return fmt.Errorf("无权催办该申请")
		}
	}
	// 判重：同实例同节点已有催办记录则跳过（与自动超时催办口径一致）；事务内写记录避免 SQLITE_BUSY
	return DB.Transaction(func(tx *gorm.DB) error {
		var cnt int64
		_ = tx.Model(&model.FlowRecord{}).Where("instance_id = ? AND node_id = ? AND action = '9'", instanceID, inst.CurrentNodeID).Count(&cnt).Error
		if cnt == 0 {
			return flowAddRecord(tx, instanceID, inst.CurrentNodeID, "", "9", operator.Nickname, "", "催办审批", operatorID)
		}
		return nil
	})
}

// MarkCcRead 抄送标记已读
func MarkCcRead(taskID, operatorID uint) error {
	res := DB.Model(&model.FlowTask{}).
		Where("flow_task_id = ? AND assignee_id = ? AND task_type = '2' AND status = '0'", taskID, operatorID).
		Update("status", "6")
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("任务不存在或已被处理")
	}
	return nil
}

// ---------- 流程定义 CRUD ----------

// FlowNodeInput 节点输入（设计保存用）
type FlowNodeInput struct {
	FlowNodeID   uint   `json:"flowNodeId"`
	NodeOrder    int    `json:"nodeOrder"`
	NodeName     string `json:"nodeName"`
	ApproverType string `json:"approverType"`
	ApproverIDs  string `json:"approverIds"`
	NodeMode     string `json:"nodeMode"`
	CanAddSign   string `json:"canAddSign"`
	TimeoutHours int    `json:"timeoutHours"`
	RejectTo     string `json:"rejectTo"`
	Remark       string `json:"remark"`
}

// FlowDefInput 流程定义输入
type FlowDefInput struct {
	FlowDefID    uint            `json:"flowDefId"`
	DefName      string          `json:"defName"`
	DefKey       string          `json:"defKey"`
	Category     string          `json:"category"`
	Status       string          `json:"status"`
	TimeoutHours int             `json:"timeoutHours"`
	Remark       string          `json:"remark"`
	Nodes        []FlowNodeInput `json:"nodes"`
}

// ListFlowDefs 流程定义列表（节点预载，支持名称/标识模糊 + 状态 + 分类过滤）
func ListFlowDefs(pageNum, pageSize int, keyword, status, category string) ([]model.FlowDef, int64, error) {
	q := DB.Model(&model.FlowDef{})
	if k := strings.TrimSpace(keyword); k != "" {
		like := "%" + escapeLike(k) + "%"
		q = q.Where("("+likeExprEsc("def_name")+" OR "+likeExprEsc("def_key")+")", like, like)
	}
	if s := strings.TrimSpace(status); s != "" {
		q = q.Where("status = ?", s)
	}
	if c := strings.TrimSpace(category); c != "" {
		q = q.Where("category = ?", c)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	pageNum, pageSize = normPage(pageNum, pageSize)
	var defs []model.FlowDef
	if err := q.Order("flow_def_id DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&defs).Error; err != nil {
		return nil, 0, err
	}
	for i := range defs {
		defs[i].Nodes = flowNodesOf(defs[i].ID)
	}
	return defs, total, nil
}

// ListEnabledFlowDefs 启用中的流程（发起申请下拉）
func ListEnabledFlowDefs() ([]model.FlowDef, error) {
	var defs []model.FlowDef
	if err := DB.Where("status = '0'").Order("flow_def_id ASC").Find(&defs).Error; err != nil {
		return nil, err
	}
	return defs, nil
}

// GetFlowDef 流程定义详情（含节点）
func GetFlowDef(id uint) (model.FlowDef, error) {
	def, err := flowDefByID(id)
	if err != nil {
		return def, err
	}
	def.Nodes = flowNodesOf(id)
	return def, nil
}

// SaveFlowDef 新建/更新流程定义（更新时若存在运行中实例则拒绝改节点）
func SaveFlowDef(input FlowDefInput, createBy string) (uint, error) {
	if strings.TrimSpace(input.DefName) == "" || strings.TrimSpace(input.DefKey) == "" {
		return 0, fmt.Errorf("请填写流程名称与标识")
	}
	if len(input.Nodes) == 0 {
		return 0, fmt.Errorf("至少配置一个审批节点")
	}
	// 校验节点
	for i := range input.Nodes {
		n := &input.Nodes[i]
		if strings.TrimSpace(n.NodeName) == "" {
			return 0, fmt.Errorf("第 %d 步节点名称不能为空", i+1)
		}
		switch n.ApproverType {
		case "role", "user":
			if strings.TrimSpace(n.ApproverIDs) == "" {
				return 0, fmt.Errorf("节点「%s」未配置审批人", n.NodeName)
			}
		case "dept_leader":
		default:
			return 0, fmt.Errorf("节点「%s」审批人类型未知", n.NodeName)
		}
		if n.NodeMode != "any" {
			n.NodeMode = "one"
		}
		if n.CanAddSign != "1" {
			n.CanAddSign = "0"
		}
		if n.RejectTo != "1" {
			n.RejectTo = "0"
		}
		if n.NodeOrder <= 0 {
			n.NodeOrder = i + 1
		}
	}
	var defID uint
	err := DB.Transaction(func(tx *gorm.DB) error {
		if input.FlowDefID > 0 {
			var old model.FlowDef
			if err := tx.First(&old, input.FlowDefID).Error; err != nil {
				return fmt.Errorf("流程不存在")
			}
			// 存在运行中实例时禁止修改节点结构（实例用节点快照推进，避免配置漂移）
			var running int64
			if err := tx.Model(&model.FlowInstance{}).Where("flow_def_id = ? AND status = '1'", input.FlowDefID).Count(&running).Error; err != nil {
				return err
			}
			if running > 0 {
				return fmt.Errorf("该流程存在运行中的实例，禁止修改节点；请先处理相关申请")
			}
			def := model.FlowDef{
				ID:           input.FlowDefID,
				DefName:      input.DefName,
				DefKey:       input.DefKey,
				Category:     input.Category,
				Status:       input.Status,
				TimeoutHours: input.TimeoutHours,
				Remark:       input.Remark,
				UpdatedAt:    time.Now(),
			}
			if def.Status != "1" {
				def.Status = "0"
			}
			if err := tx.Model(&model.FlowDef{}).Where("flow_def_id = ?", input.FlowDefID).Updates(map[string]interface{}{
				"def_name": def.DefName, "def_key": def.DefKey, "category": def.Category,
				"status": def.Status, "timeout_hours": def.TimeoutHours, "remark": def.Remark, "updated_at": def.UpdatedAt,
			}).Error; err != nil {
				return err
			}
			if err := tx.Where("flow_def_id = ?", input.FlowDefID).Delete(&model.FlowNode{}).Error; err != nil {
				return err
			}
			defID = input.FlowDefID
		} else {
			// 标识唯一
			var cnt int64
			if err := tx.Model(&model.FlowDef{}).Where("def_key = ?", input.DefKey).Count(&cnt).Error; err != nil {
				return err
			}
			if cnt > 0 {
				return fmt.Errorf("流程标识「%s」已存在", input.DefKey)
			}
			def := model.FlowDef{
				DefName: input.DefName, DefKey: input.DefKey, Category: input.Category,
				Status: "0", TimeoutHours: input.TimeoutHours, Remark: input.Remark, CreateBy: createBy,
			}
			if err := tx.Create(&def).Error; err != nil {
				return err
			}
			defID = def.ID
		}
		// 重建节点（按 order 排序）
		sorted := make([]FlowNodeInput, len(input.Nodes))
		copy(sorted, input.Nodes)
		sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].NodeOrder < sorted[j].NodeOrder })
		for i := range sorted {
			n := sorted[i]
			node := model.FlowNode{
				FlowDefID:    defID,
				NodeOrder:    n.NodeOrder,
				NodeName:     n.NodeName,
				ApproverType: n.ApproverType,
				ApproverIDs:  strings.TrimSpace(n.ApproverIDs),
				NodeMode:     n.NodeMode,
				CanAddSign:   n.CanAddSign,
				TimeoutHours: n.TimeoutHours,
				RejectTo:     n.RejectTo,
				Remark:       n.Remark,
			}
			if err := tx.Create(&node).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return defID, err
}

// DeleteFlowDef 删除流程定义（存在实例时拒绝）
func DeleteFlowDef(id uint) error {
	var cnt int64
	if err := DB.Model(&model.FlowInstance{}).Where("flow_def_id = ?", id).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return fmt.Errorf("该流程已有申请记录，禁止删除")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("flow_def_id = ?", id).Delete(&model.FlowNode{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.FlowDef{}, id).Error
	})
}

// ChangeFlowDefStatus 启停流程
func ChangeFlowDefStatus(id uint, status string) error {
	if status != "1" {
		status = "0"
	}
	res := DB.Model(&model.FlowDef{}).Where("flow_def_id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("流程不存在")
	}
	return nil
}

// ---------- 实例/任务查询 ----------

// FlowTaskVO 任务列表展示对象（附带实例标题/流程名/申请人）
type FlowTaskVO struct {
	model.FlowTask
	DefName       string `json:"defName"`
	Title         string `json:"title"`
	ApplicantName string `json:"applicantName"`
}

// flowTasksWithInst 组装任务列表的实例维度字段（批量查实例，避免 N+1）
func flowTasksWithInst(tasks []model.FlowTask) []FlowTaskVO {
	if len(tasks) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.InstanceID)
	}
	var insts []model.FlowInstance
	DB.Where("flow_instance_id IN ?", ids).Find(&insts)
	byID := make(map[uint]model.FlowInstance, len(insts))
	for _, i := range insts {
		byID[i.ID] = i
	}
	vos := make([]FlowTaskVO, 0, len(tasks))
	for _, t := range tasks {
		vo := FlowTaskVO{FlowTask: t}
		if i, ok := byID[t.InstanceID]; ok {
			vo.DefName = i.DefName
			vo.Title = i.Title
			vo.ApplicantName = i.ApplicantName
		}
		vos = append(vos, vo)
	}
	return vos
}

// ListFlowApply 我的申请列表（支持标题模糊 + 状态过滤）
func ListFlowApply(applicantID uint, pageNum, pageSize int, keyword, status string) ([]model.FlowInstance, int64, error) {
	q := DB.Model(&model.FlowInstance{}).Where("applicant_id = ?", applicantID)
	if k := strings.TrimSpace(keyword); k != "" {
		like := "%" + escapeLike(k) + "%"
		q = q.Where(likeExprEsc("title"), like)
	}
	if s := strings.TrimSpace(status); s != "" {
		q = q.Where("status = ?", s)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	pageNum, pageSize = normPage(pageNum, pageSize)
	var list []model.FlowInstance
	if err := q.Order("flow_instance_id DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListFlowInstances 全部实例（管理员视角，支持标题模糊 + 状态过滤）
func ListFlowInstances(pageNum, pageSize int, keyword, status string) ([]model.FlowInstance, int64, error) {
	q := DB.Model(&model.FlowInstance{})
	if k := strings.TrimSpace(keyword); k != "" {
		like := "%" + escapeLike(k) + "%"
		q = q.Where(likeExprEsc("title"), like)
	}
	if s := strings.TrimSpace(status); s != "" {
		q = q.Where("status = ?", s)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	pageNum, pageSize = normPage(pageNum, pageSize)
	var list []model.FlowInstance
	if err := q.Order("flow_instance_id DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListFlowTodo 我的待办（审批任务，支持标题模糊过滤）
func ListFlowTodo(assigneeID uint, pageNum, pageSize int, keyword string) ([]FlowTaskVO, int64, error) {
	q := DB.Model(&model.FlowTask{}).
		Where("assignee_id = ? AND task_type IN ('1','3') AND status = '0'", assigneeID)
	if k := strings.TrimSpace(keyword); k != "" {
		like := "%" + escapeLike(k) + "%"
		q = q.Where("instance_id IN (SELECT flow_instance_id FROM sys_flow_instance WHERE "+likeExprEsc("title")+")", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	pageNum, pageSize = normPage(pageNum, pageSize)
	var list []model.FlowTask
	if err := q.Order("flow_task_id DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return flowTasksWithInst(list), total, nil
}

// ListFlowDone 我处理的（已办，支持标题模糊过滤）
func ListFlowDone(assigneeID uint, pageNum, pageSize int, keyword string) ([]FlowTaskVO, int64, error) {
	q := DB.Model(&model.FlowTask{}).
		Where("assignee_id = ? AND task_type IN ('1','3') AND status IN ('1','2','3')", assigneeID)
	if k := strings.TrimSpace(keyword); k != "" {
		like := "%" + escapeLike(k) + "%"
		q = q.Where("instance_id IN (SELECT flow_instance_id FROM sys_flow_instance WHERE "+likeExprEsc("title")+")", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	pageNum, pageSize = normPage(pageNum, pageSize)
	var list []model.FlowTask
	if err := q.Order("flow_task_id DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return flowTasksWithInst(list), total, nil
}

// ListFlowCc 抄送我的（支持标题模糊过滤）
func ListFlowCc(assigneeID uint, pageNum, pageSize int, keyword string) ([]FlowTaskVO, int64, error) {
	q := DB.Model(&model.FlowTask{}).
		Where("assignee_id = ? AND task_type = '2'", assigneeID)
	if k := strings.TrimSpace(keyword); k != "" {
		like := "%" + escapeLike(k) + "%"
		q = q.Where("instance_id IN (SELECT flow_instance_id FROM sys_flow_instance WHERE "+likeExprEsc("title")+")", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	pageNum, pageSize = normPage(pageNum, pageSize)
	var list []model.FlowTask
	if err := q.Order("flow_task_id DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return flowTasksWithInst(list), total, nil
}

// FlowInstanceDetail 实例详情（含定义节点、任务、记录）
type FlowInstanceDetail struct {
	Instance model.FlowInstance `json:"instance"`
	Nodes    []model.FlowNode   `json:"nodes"`
	Tasks    []model.FlowTask   `json:"tasks"`
	Records  []model.FlowRecord `json:"records"`
}

// GetFlowInstanceDetail 实例详情：节点进度 + 任务 + 审批记录
// 数据权限：仅申请人本人 / 实例任务任一处理人（含审批、抄送）/ 管理员可查看，防横向越权（IDOR）
func GetFlowInstanceDetail(id, operatorID uint) (FlowInstanceDetail, error) {
	var detail FlowInstanceDetail
	if err := DB.First(&detail.Instance, id).Error; err != nil {
		return detail, fmt.Errorf("申请不存在")
	}
	if !IsAdmin(operatorID) && detail.Instance.ApplicantID != operatorID {
		var mine int64
		if err := DB.Model(&model.FlowTask{}).Where("instance_id = ? AND assignee_id = ?", id, operatorID).Count(&mine).Error; err != nil || mine == 0 {
			return detail, fmt.Errorf("无权查看该申请")
		}
	}
	detail.Nodes = flowNodesOf(detail.Instance.FlowDefID)
	if err := DB.Where("instance_id = ?", id).Order("flow_task_id ASC").Find(&detail.Tasks).Error; err != nil {
		return detail, err
	}
	if err := DB.Where("instance_id = ?", id).Order("flow_record_id ASC").Find(&detail.Records).Error; err != nil {
		return detail, err
	}
	return detail, nil
}

// ---------- 超时催办 watcher ----------

// StartOverdueWatcher 启动审批超时催办协程：每分钟扫描运行中实例的待办任务，
// 超过 overdue_at 且未催办过的任务写入一次催办记录（action=9，判重：同任务已有催办记录则跳过）
func StartOverdueWatcher(stop <-chan struct{}) {
	backgroundWG.Add(1)
	go func() {
		defer backgroundWG.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				watchOverdue()
			}
		}
	}()
}

func watchOverdue() {
	now := time.Now()
	var tasks []model.FlowTask
	if err := DB.Where("status = '0' AND task_type IN ('1','3') AND overdue_at IS NOT NULL AND overdue_at < ?", now).
		Find(&tasks).Error; err != nil || len(tasks) == 0 {
		return
	}
	// 批量取运行中实例与已有催办记录，避免逐任务查询（N+1）
	instanceIDs := make([]uint, 0, len(tasks))
	for _, t := range tasks {
		instanceIDs = append(instanceIDs, t.InstanceID)
	}
	var running []model.FlowInstance
	_ = DB.Select("flow_instance_id, status").Where("flow_instance_id IN ? AND status = '1'", instanceIDs).Find(&running).Error
	runningSet := make(map[uint]bool, len(running))
	for _, i := range running {
		runningSet[i.ID] = true
	}
	var records []struct {
		InstanceID uint `gorm:"column:instance_id"`
		NodeID     uint `gorm:"column:node_id"`
	}
	_ = DB.Model(&model.FlowRecord{}).Select("instance_id, node_id").
		Where("action = '9' AND instance_id IN ?", instanceIDs).Scan(&records).Error
	urged := make(map[[2]uint]bool, len(records))
	for _, r := range records {
		urged[[2]uint{r.InstanceID, r.NodeID}] = true
	}
	for _, t := range tasks {
		if !runningSet[t.InstanceID] {
			continue
		}
		key := [2]uint{t.InstanceID, t.NodeID}
		if urged[key] {
			continue
		}
		// 事务内再判重并写记录，避免与手动催办并发重复写入
		err := DB.Transaction(func(tx *gorm.DB) error {
			var cnt int64
			if err := tx.Model(&model.FlowRecord{}).Where("instance_id = ? AND node_id = ? AND action = '9'", t.InstanceID, t.NodeID).Count(&cnt).Error; err != nil {
				return err
			}
			if cnt > 0 {
				return nil
			}
			return flowAddRecord(tx, t.InstanceID, t.NodeID, t.NodeName, "9", "系统", t.AssigneeName, "审批超时，自动催办", 0)
		})
		if err == nil {
			urged[key] = true
		}
	}
}
